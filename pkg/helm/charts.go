package helm

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"sort"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// ProfileAnnotation is the annotation that Helm charts must have to indicate
// that they provide a Profile.
const ProfileAnnotation = "weave.works/profile"
const LayerAnnotation = "weave.works/layer"

// NewRepoManager creates and returns a new RepoManager.
func NewRepoManager(kc client.Client, cacheDir string) *RepoManager {
	return &RepoManager{
		Client:   kc,
		CacheDir: cacheDir,
		envSettings: &cli.EnvSettings{
			Debug:            true,
			RepositoryCache:  cacheDir,
			RepositoryConfig: path.Join(cacheDir, "/repository.yaml"),
		},
	}
}

// RepoManager implements HelmRepoManager interface using the Helm library packages.
type RepoManager struct {
	client.Client
	CacheDir    string
	envSettings *cli.EnvSettings
}

// ChartReference is a Helm chart reference, the SourceRef is a Flux
// SourceReference for the Helm chart.
type ChartReference struct {
	Chart     string
	Version   string
	SourceRef helmv2beta1.CrossNamespaceObjectReference
}

// DefaultChartGetter provides default ways to get a chart index.yaml based on
// the URL scheme.
var defaultChartGetters = getter.Providers{
	getter.Provider{
		Schemes: []string{"http", "https"},
		New:     getter.NewHTTPGetter,
	},
}

type ChartPredicate func(*repo.ChartVersion) bool

// Profiles is a predicate for scanning charts with the ProfileAnnotation.
var Profiles = func(v *repo.ChartVersion) bool {
	return hasAnnotation(v.Metadata, ProfileAnnotation)
}

// GetCharts filters charts using the provided predicate.
// TODO: Add caching based on the Status Artifact Revision.
func (h *RepoManager) GetCharts(ctx context.Context, hr *sourcev1beta1.HelmRepository, pred ChartPredicate) ([]*pb.Profile, error) {
	chartRepo, err := fetchIndexFile(hr.Status.URL)
	if err != nil {
		return nil, fmt.Errorf("fetching profiles from HelmRepository %s/%s: %w",
			hr.GetName(), hr.GetNamespace(), err)
	}

	ps := make(map[string]*pb.Profile)

	for name, versions := range chartRepo.Entries {
		for _, v := range versions {
			if pred(v) {
				// if already added, update the versions array
				if p, ok := ps[name]; ok {
					p.AvailableVersions = append(p.AvailableVersions, v.Version)
				} else { // otherwise create a new profile and add to map
					p = &pb.Profile{
						Name:        name,
						Home:        v.Home,
						Sources:     v.Sources,
						Description: v.Description,
						Keywords:    v.Keywords,
						Icon:        v.Icon,
						KubeVersion: v.KubeVersion,
						Layer:       getLayer(v.Annotations),
					}
					for _, m := range v.Maintainers {
						p.Maintainers = append(p.Maintainers, &pb.Maintainer{
							Name:  m.Name,
							Email: m.Email,
							Url:   m.URL,
						})
					}
					p.AvailableVersions = append(p.AvailableVersions, v.Version)
					ps[name] = p
				}
			}
		}
	}

	profiles := []*pb.Profile{}

	for _, p := range ps {
		sort.Strings(p.AvailableVersions)
		profiles = append(profiles, p)
	}

	return profiles, nil
}

// GetValuesFile fetches the value file from a chart.
func (h *RepoManager) GetValuesFile(ctx context.Context, helmRepo *sourcev1beta1.HelmRepository, c *ChartReference, filename string) ([]byte, error) {
	if err := h.updateCache(ctx, helmRepo); err != nil {
		return nil, fmt.Errorf("updating cache: %w", err)
	}

	chart, err := h.loadChart(ctx, helmRepo, c)
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

func (h *RepoManager) updateCache(ctx context.Context, helmRepo *sourcev1beta1.HelmRepository) error {
	entry, err := h.entryForRepository(ctx, helmRepo)
	if err != nil {
		return fmt.Errorf("failed to build repository entry: %w", err)
	}

	r, err := repo.NewChartRepository(entry, defaultChartGetters)
	if err != nil {
		return fmt.Errorf("error creating chart repository: %w", err)
	}

	r.CachePath = h.CacheDir
	if _, err := r.DownloadIndexFile(); err != nil {
		return fmt.Errorf("error downloading index file: %w", err)
	}

	return nil
}

func (h *RepoManager) loadChart(ctx context.Context, helmRepo *sourcev1beta1.HelmRepository, c *ChartReference) (*chart.Chart, error) {
	o, err := h.chartPathOptionsFromRepository(ctx, helmRepo, c)
	if err != nil {
		return nil, fmt.Errorf("failed to configure client: %w", err)
	}

	chartLocation, err := o.LocateChart(c.Chart, h.envSettings)
	if err != nil {
		return nil, fmt.Errorf("locating chart %q: %w", c.Chart, err)
	}

	chart, err := loader.Load(chartLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart %q: %w", c.Chart, err)
	}

	return chart, nil
}

func (h *RepoManager) chartPathOptionsFromRepository(ctx context.Context, helmRepo *sourcev1beta1.HelmRepository, c *ChartReference) (*action.ChartPathOptions, error) {
	// TODO: This should probably use Verify: true
	co := &action.ChartPathOptions{
		RepoURL: helmRepo.Spec.URL,
		Version: c.Version,
	}

	if helmRepo.Spec.SecretRef != nil {
		username, password, err := credsForRepository(ctx, h.Client, helmRepo)
		if err != nil {
			return nil, err
		}

		co.Username = username
		co.Password = password
	}

	return co, nil
}

func (h *RepoManager) entryForRepository(ctx context.Context, helmRepo *sourcev1beta1.HelmRepository) (*repo.Entry, error) {
	entry := &repo.Entry{
		Name: helmRepo.GetName() + "-" + helmRepo.GetNamespace(),
		URL:  helmRepo.Spec.URL,
	}

	if helmRepo.Spec.SecretRef != nil {
		username, password, err := credsForRepository(ctx, h.Client, helmRepo)
		if err != nil {
			return nil, err
		}

		entry.Username = username
		entry.Password = password
	}

	return entry, nil
}

func credsForRepository(ctx context.Context, kc client.Client, hr *sourcev1beta1.HelmRepository) (string, string, error) {
	var secret corev1.Secret
	if err := kc.Get(ctx, types.NamespacedName{Name: hr.Spec.SecretRef.Name, Namespace: hr.Namespace}, &secret); err != nil {
		return "", "", fmt.Errorf("repository authentication: %w", err)
	}

	return string(secret.Data["username"]), string(secret.Data["password"]), nil
}

func fetchIndexFile(chartURL string) (*repo.IndexFile, error) {
	if hostname := os.Getenv("SOURCE_CONTROLLER_LOCALHOST"); hostname != "" {
		u, err := url.Parse(chartURL)
		if err != nil {
			return nil, err
		}
		u.Host = hostname
		chartURL = u.String()
	}

	u, err := url.Parse(chartURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL %q: %w", chartURL, err)
	}

	c, err := defaultChartGetters.ByScheme(u.Scheme)
	if err != nil {
		return nil, fmt.Errorf("no provider for scheme %q: %w", u.Scheme, err)
	}

	res, err := c.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("error fetching index file: %w", err)
	}

	b, err := ioutil.ReadAll(res)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	i := &repo.IndexFile{}
	if err := yaml.Unmarshal(b, i); err != nil {
		return nil, fmt.Errorf("error unmarshaling chart response: %w", err)
	}

	if i.APIVersion == "" {
		return nil, repo.ErrNoAPIVersion
	}

	i.SortEntries()

	return i, nil
}

func getLayer(annotations map[string]string) string {
	return annotations[LayerAnnotation]
}

func hasAnnotation(cm *chart.Metadata, name string) bool {
	for k := range cm.Annotations {
		if k == name {
			return true
		}
	}

	return false
}
