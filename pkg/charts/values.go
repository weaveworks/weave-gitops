package charts

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
)

// ChartReference is a Helm chart reference, the SourceRef is a Flux
// SourceReference for the Helm chart.
type ChartReference struct {
	Chart     string
	Version   string
	SourceRef helmv2beta1.CrossNamespaceObjectReference
}

// HelmChartClient implements ChartClient using the Helm library packages.
type HelmChartClient struct {
	client.Client
	Namespace  string
	Repository *sourcev1beta1.HelmRepository
	CacheDir   string
}

// WithCacheDir configures the HelmChartClient to use the directory for the Helm
// repository cache.
func WithCacheDir(dir string) func(*HelmChartClient) {
	return func(h *HelmChartClient) {
		h.CacheDir = dir
	}
}

// NewHelmChartClient creates and returns a new HelmChartClient.
func NewHelmChartClient(kc client.Client, ns string, hr *sourcev1beta1.HelmRepository, opts ...func(*HelmChartClient)) *HelmChartClient {
	h := &HelmChartClient{
		Client:     kc,
		Namespace:  ns,
		Repository: hr,
	}
	for _, o := range opts {
		o(h)
	}
	return h
}

func (h *HelmChartClient) SetRepository(repo *sourcev1beta1.HelmRepository) {
	h.Repository = repo
}

// UpdateCache must be called before any calls to fetch charts.
//
// If the cache dir is empty, then it will use the default Helm cache directory
// for the repo cache.
func (h *HelmChartClient) UpdateCache(ctx context.Context) error {
	entry, err := h.entryForRepository(ctx)
	if err != nil {
		return err
	}
	r, err := repo.NewChartRepository(entry, DefaultChartGetters)
	if err != nil {
		return err
	}
	r.CachePath = h.CacheDir
	_, err = r.DownloadIndexFile()
	return err
}

func (h *HelmChartClient) loadChart(ctx context.Context, c *ChartReference) (*chart.Chart, error) {
	o, err := h.chartPathOptionsFromRepository(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("failed to configure client: %w", err)
	}

	chartLocation, err := o.LocateChart(c.Chart, h.envSettings())
	if err != nil {
		return nil, fmt.Errorf("locating chart %q: %w", c.Chart, err)
	}
	chart, err := loader.Load(chartLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart %q: %w", c.Chart, err)
	}
	return chart, nil
}

// FileFromChart fetches the named file from a chart.
func (h *HelmChartClient) FileFromChart(ctx context.Context, c *ChartReference, filename string) ([]byte, error) {
	chart, err := h.loadChart(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("loading %s from chart: %w", filename, err)
	}
	for _, v := range chart.Raw {
		if v.Name == filename {
			return v.Data, nil
		}
	}
	return nil, fmt.Errorf("failed to find file: %s", filename)
}

func credsForRepository(ctx context.Context, kc client.Client, ns string, hr *sourcev1beta1.HelmRepository) (string, string, error) {
	var secret corev1.Secret
	if err := kc.Get(ctx, types.NamespacedName{Name: hr.Spec.SecretRef.Name, Namespace: ns}, &secret); err != nil {
		return "", "", fmt.Errorf("repository authentication: %w", err)
	}
	return string(secret.Data["username"]), string(secret.Data["password"]), nil
}

func (h *HelmChartClient) chartPathOptionsFromRepository(ctx context.Context, c *ChartReference) (*action.ChartPathOptions, error) {
	// TODO: This should probably use Verify: true
	co := &action.ChartPathOptions{
		RepoURL: h.Repository.Spec.URL,
		Version: c.Version,
	}

	if h.Repository.Spec.SecretRef != nil {
		username, password, err := credsForRepository(ctx, h.Client, h.Namespace, h.Repository)
		if err != nil {
			return nil, err
		}
		co.Username = username
		co.Password = password
	}
	return co, nil
}

func (h *HelmChartClient) entryForRepository(ctx context.Context) (*repo.Entry, error) {
	entry := &repo.Entry{
		Name: h.Repository.GetName() + "-" + h.Repository.GetNamespace(),
		URL:  h.Repository.Spec.URL,
	}
	if h.Repository.Spec.SecretRef != nil {
		username, password, err := credsForRepository(ctx, h.Client, h.Namespace, h.Repository)
		if err != nil {
			return nil, err
		}
		entry.Username = username
		entry.Password = password
	}
	return entry, nil
}

func (h *HelmChartClient) envSettings() *cli.EnvSettings {
	conf := cli.New()
	conf.Debug = true
	if h.CacheDir != "" {
		conf.RepositoryCache = h.CacheDir
		conf.RepositoryConfig = path.Join(h.CacheDir, "/repository.yaml")
	}
	return conf
}

func ParseValues(chart string, version string, values string, clusterName string, helmRepo *sourcev1beta1.HelmRepository) (*helmv2beta1.HelmRelease, error) {
	decoded, err := base64.StdEncoding.DecodeString(values)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode values: %w", err)
	}
	vals := map[string]interface{}{}
	yaml.Unmarshal(decoded, &vals)
	jsonValues, err := json.Marshal(vals)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML values into JSON: %w", err)
	}

	hr := helmv2beta1.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", clusterName, chart),
			Namespace: "wego-system",
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: helmv2beta1.GroupVersion.Identifier(),
			Kind:       helmv2beta1.HelmReleaseKind,
		},
		Spec: helmv2beta1.HelmReleaseSpec{
			Chart: helmv2beta1.HelmChartTemplate{
				Spec: helmv2beta1.HelmChartTemplateSpec{
					Chart:   chart,
					Version: version,
					SourceRef: helmv2beta1.CrossNamespaceObjectReference{
						APIVersion: sourcev1beta1.GroupVersion.Identifier(),
						Kind:       sourcev1beta1.HelmRepositoryKind,
						Name:       helmRepo.ObjectMeta.Name,
						Namespace:  helmRepo.ObjectMeta.Namespace,
					},
				},
			},
			Interval: metav1.Duration{Duration: time.Minute},
			Values:   &apiextensionsv1.JSON{Raw: jsonValues},
		},
	}

	return &hr, nil
}
