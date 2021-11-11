package upgrade

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/weaveworks/pctl/pkg/catalog"
	pctl_git "github.com/weaveworks/pctl/pkg/git"
	"github.com/weaveworks/pctl/pkg/install"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type UpgradeValues struct {
	Remote         string
	RepoURL        string
	RepoOrgAndName string
	HeadBranch     string
	BaseBranch     string
	CommitMessage  string
	Namespace      string
	ProfileBranch  string
	ConfigMap      string
	Out            string
	GitRepository  string
	DryRun         bool
}

type UpgradeConfigs struct {
	CLIGitConfig  pctl_git.CLIGitConfig
	SCMConfig     pctl_git.SCMConfig
	InstallConfig install.Config
	Profile       catalog.Profile
}

const EnterpriseChartURL string = "https://charts.dev.wkp.weave.works/charts-v3"
const CredentialsSecretName string = "weave-gitops-enterprise-credentials"

// Upgrade installs the private weave-gitops-enterprise profile into the working directory:
// 1. Private deploy key is read from a secret in the cluster
// 2. Private profiles repo is cloned locally using the deploy key
// 3. pctl is used to install the profile from the local clone into the current working directory
// 4. pctl is used to add, commit, push and create a PR.
//
func Upgrade(ctx context.Context, upgradeValues UpgradeValues, w io.Writer) error {
	kubeClient, err := makeKubeClient()
	if err != nil {
		return fmt.Errorf("error creating client for cluster %v", err)
	}

	gitClient := git.New(nil, wrapper.NewGoGit())

	return upgrade(ctx, upgradeValues, gitClient, kubeClient, auth.InitGitProvider, w)
}

type InitGitProvider func(
	repoUrl gitproviders.RepoURL,
	osysClient osys.Osys,
	logger logger.Logger,
	cliAuthHandler auth.BlockingCLIAuthHandler,
	getAccountType gitproviders.AccountTypeGetter,
) (gitproviders.GitProvider, error)

func upgrade(ctx context.Context, upgradeValues UpgradeValues, gitClient git.Git, kubeClient client.Client, initGitProvider InitGitProvider, w io.Writer) error {
	uv, err := buildUpgradeConfigs(ctx, upgradeValues, kubeClient, gitClient, w)
	if err != nil {
		return fmt.Errorf("failed to build upgrade configs: %v", err)
	}

	out, err := marshalToYamlStream(makeHelmResources(uv.Namespace))
	if err != nil {
		return fmt.Errorf("error marshalling helm resources: %w", err)
	}

	stringOut := string(out)

	if uv.DryRun {
		w.Write([]byte(stringOut + "\n"))
		return nil
	}

	err = getBasicAuth(ctx, kubeClient, uv.Namespace)
	if err != nil {
		return fmt.Errorf("failed to load deploy key for profiles repos from cluster: %v", err)
	}

	normalizedURL, err := gitproviders.NewRepoURL(uv.RepoURL)
	if err != nil {
		return fmt.Errorf("failed to normalize URL %q: %w", uv.RepoURL, err)
	}

	authHandler, err := auth.NewAuthCLIHandler(normalizedURL.Provider())
	if err != nil {
		return fmt.Errorf("error initializing cli auth handler: %w", err)
	}

	osysClient := osys.New()
	logger := logger.NewCLILogger(osysClient.Stdout())

	gitProvider, err := initGitProvider(normalizedURL, osysClient, logger, authHandler, gitproviders.GetAccountType)
	if err != nil {
		return fmt.Errorf("error obtaining git provider token: %w", err)
	}

	path := "helmrelease.yaml"

	pri := gitproviders.PullRequestInfo{
		Title:         "upgrade",
		Description:   "upgrade",
		CommitMessage: "upgrade",
		TargetBranch:  "main",
		NewBranch:     "upgrade",
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

func makeHelmResources(namespace string) []runtime.Object {
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
					// Version: latest,
				},
			},
			Values: &v1.JSON{Raw: []byte(`"hi"`)},
		},
	}

	return []runtime.Object{helmRepository, helmRelease}
}

func makeKubeClient() (client.Client, error) {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		sourcev1.AddToScheme,
		corev1.AddToScheme,
	}

	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		return nil, fmt.Errorf("error adding sourcev1 to kube client scheme %v", err)
	}

	kubeClientConfig := config.GetConfigOrDie()

	return client.New(kubeClientConfig, client.Options{Scheme: scheme})
}

// buildUpgradeConfigs sets some flags default values from the env
func buildUpgradeConfigs(ctx context.Context, uv UpgradeValues, kubeClient client.Client, gitClient git.Git, w io.Writer) (*UpgradeValues, error) {
	repoUrlString, err := gitClient.GetRemoteUrl(".", uv.Remote)
	if err != nil {
		return nil, err
	}

	uv.RepoURL = repoUrlString

	// Calculate defaults from current working directory
	if uv.RepoOrgAndName == "" {
		githubRepoPath, err := getRepoOrgAndName(repoUrlString)
		if err != nil {
			return nil, err
		}

		fmt.Fprintf(w, "Deriving org/repo for PR as %v\n", githubRepoPath)
		uv.RepoOrgAndName = githubRepoPath
	}

	if uv.GitRepository == "" {
		gitRepositoryNameNamespace := fmt.Sprintf("%s/%s", uv.Namespace, strings.TrimSuffix(filepath.Base(repoUrlString), ".git"))
		fmt.Fprintf(w, "Deriving name of GitRepository Resource as %v\n", gitRepositoryNameNamespace)
		uv.GitRepository = gitRepositoryNameNamespace
	}

	err = ensureGitRepositoryResource(context.Background(), kubeClient, uv.GitRepository)
	if err != nil {
		return nil, err
	}

	return &uv, nil
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

func ensureGitRepositoryResource(ctx context.Context, kubeClient client.Client, gitRepository string) error {
	gitRepoNamespace, gitRepoName, err := getGitRepositoryNamespaceAndName(gitRepository)
	if err != nil {
		return err
	}

	gitRepo := &sourcev1.GitRepository{}

	err = kubeClient.Get(ctx, client.ObjectKey{
		Namespace: gitRepoNamespace,
		Name:      gitRepoName,
	}, gitRepo)
	if apiErrors.IsNotFound(err) {
		return fmt.Errorf("couldn't find GitRepository resource \"%v/%v\" in the cluster, please specify", gitRepoNamespace, gitRepoName)
	}

	if err != nil {
		return fmt.Errorf("failed to look up GitRepository %v/%v to install into: %v", gitRepoNamespace, gitRepoName, err)
	}

	return nil
}

func getGitRepositoryNamespaceAndName(gitRepository string) (string, string, error) {
	split := strings.Split(gitRepository, "/")
	if len(split) != 2 {
		return "", "", fmt.Errorf("git-repository must in format <namespace>/<name>; was: %s", gitRepository)
	}

	return split[0], split[1], nil
}

// getRepoOrgAndName transforms git remotes to github "paths"
// - git@github.com:org/repo.git -> org/repo
// - https://github.com/org/repo.git -> org/repo
//
func getRepoOrgAndName(url string) (string, error) {
	repoEndpoint, err := transport.NewEndpoint(url)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(strings.TrimSuffix(repoEndpoint.Path, ".git"), "/"), nil
}
