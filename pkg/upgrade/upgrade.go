package upgrade

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/helm/helm/pkg/strvals"
	"github.com/weaveworks/pctl/pkg/catalog"
	pctl_git "github.com/weaveworks/pctl/pkg/git"
	"github.com/weaveworks/pctl/pkg/install"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type UpgradeValues struct {
	Remote         string
	AppConfigURL   string
	HeadBranch     string
	BaseBranch     string
	CommitMessage  string
	Namespace      string
	ProfileBranch  string
	ProfileVersion string
	ConfigMap      string
	Out            string
	GitRepository  string
	Values         []string
	DryRun         bool
}

type UpgradeConfigs struct {
	CLIGitConfig  pctl_git.CLIGitConfig
	SCMConfig     pctl_git.SCMConfig
	InstallConfig install.Config
	Profile       catalog.Profile
}

type JSONMap map[string]interface{}

const EnterpriseChartURL string = "https://charts.dev.wkp.weave.works/charts-v3"
const CredentialsSecretName string = "weave-gitops-enterprise-credentials"

// Upgrade installs the private weave-gitops-enterprise profile into the working directory:
// 1. Private deploy key is read from a secret in the cluster
// 2. Private profiles repo is cloned locally using the deploy key
// 3. pctl is used to install the profile from the local clone into the current working directory
// 4. pctl is used to add, commit, push and create a PR.
//
func Upgrade(ctx context.Context, gitClient git.Git, gitProvider gitproviders.GitProvider, upgradeValues UpgradeValues, logger logger.Logger, w io.Writer) error {
	kube, kubeClient, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("error creating client for cluster %v", err)
	}

	return upgrade(ctx, upgradeValues, kube, gitClient, kubeClient, gitProvider, logger, w)
}

func upgrade(ctx context.Context, uv UpgradeValues, kube kube.Kube, gitClient git.Git, kubeClient client.Client, gitProvider gitproviders.GitProvider, logger logger.Logger, w io.Writer) error {
	cname, err := kube.GetClusterName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cluster name: %w", err)
	}

	resources, err := makeHelmResources(uv.Namespace, uv.ProfileVersion, cname, uv.AppConfigURL, uv.Values)
	if err != nil {
		return fmt.Errorf("error creating helm resources: %w", err)
	}

	out, err := marshalToYamlStream(resources)
	if err != nil {
		return fmt.Errorf("error marshalling helm resources: %w", err)
	}

	stringOut := string(out)

	if uv.DryRun {
		_, _ = w.Write([]byte(stringOut + "\n"))
		return nil
	}

	err = getBasicAuth(ctx, kubeClient, uv.Namespace)
	if err != nil {
		return fmt.Errorf("failed to load deploy key for profiles repos from cluster: %v", err)
	}

	normalizedURL, err := gitproviders.NewRepoURL(uv.AppConfigURL)
	if err != nil {
		return fmt.Errorf("failed to normalize URL %q: %w", uv.AppConfigURL, err)
	}

	// Create pull request
	path := filepath.Join(git.WegoRoot, git.WegoClusterDir, cname, git.WegoClusterOSWorkloadDir, git.WegoEnterpriseDir)

	pri := gitproviders.PullRequestInfo{
		Title:         "Gitops upgrade",
		Description:   "Pull request to upgrade to Weave GitOps Enterprise",
		CommitMessage: uv.CommitMessage,
		TargetBranch:  uv.BaseBranch,
		NewBranch:     uv.HeadBranch,
		Files: []gitprovider.CommitFile{
			{
				Path:    &path,
				Content: &stringOut,
			},
		},
	}

	_, err = gitProvider.CreatePullRequest(ctx, normalizedURL, pri)

	return err
}

func marshalToYamlStream(objects []runtime.Object) ([]byte, error) {
	out := [][]byte{}

	for _, obj := range objects {
		b, err := yaml.Marshal(obj)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal HelmRepository object to YAML: %w", err)
		}

		out = append(out, b)
	}

	return bytes.Join(out, []byte("---\n")), nil
}

func makeHelmResources(namespace, version, clusterName, repoURL string, values []string) ([]runtime.Object, error) {
	helmRepository := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-gitops-enterprise-charts",
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.String(),
		},
		Spec: sourcev1.HelmRepositorySpec{
			Interval: metav1.Duration{Duration: time.Minute},
			URL:      EnterpriseChartURL,
			SecretRef: &meta.LocalObjectReference{
				Name: CredentialsSecretName,
			},
		},
	}

	if version == "" {
		version = "latest"
	}

	base := JSONMap{
		"config": JSONMap{
			"cluster": JSONMap{
				"name": clusterName,
			},
			"capi": JSONMap{
				"repositoryURL": repoURL,
			},
		},
	}

	// User specified a value via --set
	for _, value := range values {
		if err := strvals.ParseInto(value, base); err != nil {
			return nil, fmt.Errorf("failed parsing --set data: %w", err)
		}
	}

	valuesJson, err := json.Marshal(base)
	if err != nil {
		return nil, fmt.Errorf("error marshalling values object: %w", err)
	}

	helmRelease := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-gitops-enterprise",
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		Spec: helmv2.HelmReleaseSpec{
			Interval: metav1.Duration{Duration: time.Minute},
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "mccp",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      helmRepository.GetName(),
						Namespace: helmRepository.GetNamespace(),
					},
					Version: version,
				},
			},
			Values: &v1.JSON{Raw: []byte(string(valuesJson))},
		},
	}

	return []runtime.Object{helmRepository, helmRelease}, nil
}

func getBasicAuth(ctx context.Context, kubeClient client.Client, ns string) error {
	deployKeySecret := &corev1.Secret{}

	err := kubeClient.Get(ctx, client.ObjectKey{
		Namespace: ns,
		Name:      CredentialsSecretName,
	}, deployKeySecret)
	if err != nil {
		return fmt.Errorf("failed to get entitlement: %v", err)
	}

	username := string(deployKeySecret.Data["username"])
	password := string(deployKeySecret.Data["password"])

	if username == "" || password == "" {
		return errors.New("username or password missing in entitlement secret, may be an old entitlement file")
	}

	return nil
}
