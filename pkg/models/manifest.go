package models

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"

	corev1 "k8s.io/api/core/v1"

	"github.com/fluxcd/go-git-providers/gitprovider"

	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type Manifest struct {
	Path    string
	Content []byte
}

const (
	MaxKubernetesResourceNameLength = 63

	AppCRDPath              = "wego-system.yaml"
	RuntimePath             = "gitops-runtime.yaml"
	SourcePath              = "flux-source-resource.yaml"
	SystemKustResourcePath  = "flux-system-kustomization-resource.yaml"
	UserKustResourcePath    = "flux-user-kustomization-resource.yaml"
	SystemKustomizationPath = "kustomization.yaml"
	WegoAppPath             = "wego-app.yaml"
	WegoConfigPath          = "wego-config.yaml"

	WegoConfigMapName = "weave-gitops-config"
)

type ManifestsParams struct {
	ClusterName   string
	WegoNamespace string
	ConfigRepo    gitproviders.RepoURL
}

// BootstrapManifests creates all yaml files that are going to be applied to the cluster
func BootstrapManifests(ctx context.Context, fluxClient flux.Flux, gitProvider gitproviders.GitProvider, kubeClient kube.Kube, params ManifestsParams) ([]Manifest, error) {
	runtimeManifests, err := fluxClient.Install(params.WegoNamespace, true)
	if err != nil {
		return nil, fmt.Errorf("failed getting runtime manifests: %w", err)
	}

	version := version.Version
	if os.Getenv("IS_TEST_ENV") != "" {
		version = "latest"
	}

	wegoAppManifests, err := manifests.GenerateWegoAppManifests(manifests.Params{AppVersion: version, Namespace: params.WegoNamespace})
	if err != nil {
		return nil, fmt.Errorf("error generating wego-app manifest: %w", err)
	}

	wegoAppManifest := bytes.Join(wegoAppManifests, []byte("---\n"))

	sourceName := CreateClusterSourceName(params.ConfigRepo)
	systemResourceName := ConstrainResourceName(fmt.Sprintf("%s-system", params.ClusterName))

	systemKustResourceManifest, err := fluxClient.CreateKustomization(systemResourceName, sourceName,
		workAroundFluxDroppingDot(git.GetSystemPath(params.ClusterName)), params.WegoNamespace)
	if err != nil {
		return nil, err
	}

	userResourceName := ConstrainResourceName(fmt.Sprintf("%s-user", params.ClusterName))

	userKustResourceManifest, err := fluxClient.CreateKustomization(userResourceName, sourceName,
		workAroundFluxDroppingDot(git.GetUserPath(params.ClusterName)), params.WegoNamespace)
	if err != nil {
		return nil, err
	}

	fluxNs := params.WegoNamespace

	fluxNamespace, err := kubeClient.FetchNamespaceWithLabel(context.Background(), flux.PartOfLabelKey, flux.PartOfLabelValue)
	if err != nil {
		if !errors.Is(err, kube.ErrNamespaceNotFound) {
			return nil, fmt.Errorf("failed fetching flux namespace: %w", err)
		}
	}

	if fluxNamespace != nil {
		fluxNs = fluxNamespace.Name
	}

	gitopsConfigMap, err := CreateGitopsConfigMap(fluxNs, params.WegoNamespace, params.ConfigRepo.String())
	if err != nil {
		return nil, err
	}

	wegoConfigManifest, err := yaml.Marshal(gitopsConfigMap)
	if err != nil {
		return nil, fmt.Errorf("failed marshalling wego config: %w", err)
	}

	configBranch, err := gitProvider.GetDefaultBranch(ctx, params.ConfigRepo)
	if err != nil {
		return nil, err
	}

	sourceManifest, err := GetSourceManifest(ctx, fluxClient, gitProvider, params.ClusterName, params.WegoNamespace, params.ConfigRepo, configBranch)
	if err != nil {
		return nil, err
	}

	return []Manifest{
		{
			Path:    git.GetSystemQualifiedPath(params.ClusterName, AppCRDPath),
			Content: manifests.AppCRD,
		},
		{
			Path:    git.GetSystemQualifiedPath(params.ClusterName, RuntimePath),
			Content: runtimeManifests,
		},
		{
			Path:    git.GetSystemQualifiedPath(params.ClusterName, SystemKustResourcePath),
			Content: systemKustResourceManifest,
		},
		{
			Path:    git.GetSystemQualifiedPath(params.ClusterName, UserKustResourcePath),
			Content: userKustResourceManifest,
		},
		{
			Path:    git.GetSystemQualifiedPath(params.ClusterName, WegoAppPath),
			Content: wegoAppManifest,
		},
		{
			Path:    git.GetSystemQualifiedPath(params.ClusterName, WegoConfigPath),
			Content: wegoConfigManifest,
		},
		sourceManifest,
	}, nil
}

// NoClusterApplicableManifests generates all yaml files that are going to be written in the config repo and cannot be applied to the cluster directly
func NoClusterApplicableManifests(params ManifestsParams) ([]Manifest, error) {
	systemKustomization := CreateKustomization(params.ClusterName, params.WegoNamespace, RuntimePath, SourcePath, SystemKustResourcePath, UserKustResourcePath, WegoAppPath)

	systemKustomizationManifest, err := yaml.Marshal(systemKustomization)
	if err != nil {
		return nil, err
	}

	return []Manifest{
		{
			Path:    git.GetSystemQualifiedPath(params.ClusterName, SystemKustomizationPath),
			Content: systemKustomizationManifest,
		},
		{
			Path:    filepath.Join(git.GetUserPath(params.ClusterName), ".keep"),
			Content: strconv.AppendQuote(nil, "# keep"),
		},
	}, nil
}

func CreateKustomization(name, namespace string, resources ...string) types.Kustomization {
	var k types.Kustomization

	k.MetaData = &types.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}
	k.APIVersion = types.KustomizationVersion
	k.Kind = types.KustomizationKind
	k.Resources = resources

	return k
}

func GetSecretRefForPrivateGitSources(ctx context.Context, gitProvider gitproviders.GitProvider, url gitproviders.RepoURL) (GeneratedSecretName, error) {
	var secretRef GeneratedSecretName

	visibility, err := gitProvider.GetRepoVisibility(ctx, url)
	if err != nil {
		return "", err
	}

	if *visibility != gitprovider.RepositoryVisibilityPublic {
		secretRef = CreateRepoSecretName(url)
	}

	return secretRef, nil
}

func CreateClusterSourceName(gitSourceURL gitproviders.RepoURL) string {
	provider := string(gitSourceURL.Provider())
	cleanRepoName := replaceUnderscores(gitSourceURL.RepositoryName())
	qualifiedName := fmt.Sprintf("wego-auto-%s-%s", provider, cleanRepoName)
	lengthConstrainedName := hashNameIfTooLong(qualifiedName)

	return lengthConstrainedName
}

func replaceUnderscores(str string) string {
	return strings.ReplaceAll(str, "_", "-")
}

type GeneratedSecretName string

func (s GeneratedSecretName) String() string {
	return string(s)
}

func CreateRepoSecretName(gitSourceURL gitproviders.RepoURL) GeneratedSecretName {
	provider := string(gitSourceURL.Provider())
	cleanRepoName := replaceUnderscores(gitSourceURL.RepositoryName())
	qualifiedName := fmt.Sprintf("wego-%s-%s", provider, cleanRepoName)
	lengthConstrainedName := hashNameIfTooLong(qualifiedName)

	return GeneratedSecretName(lengthConstrainedName)
}

func hashNameIfTooLong(name string) string {
	if !ApplicationNameTooLong(name) {
		return name
	}

	return fmt.Sprintf("wego-%x", md5.Sum([]byte(name)))
}

func ApplicationNameTooLong(name string) bool {
	return len(name) > MaxKubernetesResourceNameLength
}

func ValidateApplicationName(name string) error {
	errs := validation.IsDNS1123Label(name)
	if len(errs) > 0 {
		var s strings.Builder
		for _, e := range errs {
			if s.Len() > 0 {
				s.WriteString("; ")
			}

			s.WriteString(e)
		}

		return fmt.Errorf("invalid application name %q :%s", name, s.String())
	}

	return nil
}

func ConstrainResourceName(str string) string {
	return hashNameIfTooLong(replaceUnderscores(str))
}

func workAroundFluxDroppingDot(str string) string {
	return "." + str
}

func CreateGitopsConfigMap(fluxNamespace string, wegoNamespace string, configRepo string) (corev1.ConfigMap, error) {
	config := kube.WegoConfig{
		FluxNamespace: fluxNamespace,
		WegoNamespace: wegoNamespace,
		ConfigRepo:    configRepo,
	}

	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return corev1.ConfigMap{}, fmt.Errorf("failed marshalling wego config: %w", err)
	}

	return corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      WegoConfigMapName,
			Namespace: wegoNamespace,
		},
		Data: map[string]string{
			"config": string(configBytes),
		},
	}, nil
}

func ConvertManifestsToCommitFiles(manifests []Manifest) []gitprovider.CommitFile {
	files := make([]gitprovider.CommitFile, 0)

	for _, manifest := range manifests {
		path := manifest.Path
		content := string(manifest.Content)

		files = append(files, gitprovider.CommitFile{
			Path:    &path,
			Content: &content,
		})
	}

	return files
}

func GetClusterHash(clusterName string) string {
	return fmt.Sprintf("wego-%x", md5.Sum([]byte(clusterName)))
}

func GetSourceManifest(ctx context.Context, fluxClient flux.Flux, gitProviderClient gitproviders.GitProvider, clusterName string, namespace string, configURL gitproviders.RepoURL, branch string) (Manifest, error) {
	secretRef, err := GetSecretRefForPrivateGitSources(ctx, gitProviderClient, configURL)
	if err != nil {
		return Manifest{}, fmt.Errorf("failed getting ref secret: %w", err)
	}

	secretStr := secretRef.String()
	sourceName := CreateClusterSourceName(configURL)

	sourceManifest, err := fluxClient.CreateSourceGit(sourceName, configURL, branch, secretStr, namespace)
	if err != nil {
		return Manifest{}, fmt.Errorf("failed creating config repo git source: %w", err)
	}

	return Manifest{
		Path:    git.GetSystemQualifiedPath(clusterName, SourcePath),
		Content: sourceManifest,
	}, nil
}
