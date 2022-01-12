package models

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"

	"github.com/fluxcd/go-git-providers/gitprovider"

	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/kustomize/api/types"
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

// BootstrapManifests creates all yaml files that are going to be applied to the cluster
func BootstrapManifests(fluxClient flux.Flux, clusterName string, namespace string, configURL gitproviders.RepoURL) ([]Manifest, error) {
	runtimeManifests, err := fluxClient.Install(namespace, true)
	if err != nil {
		return nil, fmt.Errorf("failed getting runtime manifests: %w", err)
	}

	version := version.Version
	if os.Getenv("IS_TEST_ENV") != "" {
		version = "latest"
	}

	wegoAppManifests, err := manifests.GenerateWegoAppManifests(manifests.Params{AppVersion: version, Namespace: namespace})
	if err != nil {
		return nil, fmt.Errorf("error generating wego-app manifest: %w", err)
	}

	wegoAppManifest := bytes.Join(wegoAppManifests, []byte("---\n"))

	sourceName := createClusterSourceName(configURL)

	systemKustResourceManifest, err := fluxClient.CreateKustomization(ConstrainResourceName(fmt.Sprintf("%s-system", clusterName)), sourceName,
		workAroundFluxDroppingDot(git.GetSystemPath(clusterName)), namespace)
	if err != nil {
		return nil, err
	}

	userKustResourceManifest, err := fluxClient.CreateKustomization(ConstrainResourceName(fmt.Sprintf("%s-user", clusterName)), sourceName,
		workAroundFluxDroppingDot(git.GetUserPath(clusterName)), namespace)
	if err != nil {
		return nil, err
	}

	gitopsConfigMap, err := gitopsConfigMap(namespace, namespace)
	if err != nil {
		return nil, err
	}

	wegoConfigManifest, err := yaml.Marshal(gitopsConfigMap)
	if err != nil {
		return nil, fmt.Errorf("failed marshalling wego config: %w", err)
	}

	return []Manifest{
		{
			Path:    git.GetSystemQualifiedPath(clusterName, AppCRDPath),
			Content: manifests.AppCRD,
		},
		{
			Path:    git.GetSystemQualifiedPath(clusterName, RuntimePath),
			Content: runtimeManifests,
		},
		{
			Path:    git.GetSystemQualifiedPath(clusterName, SystemKustResourcePath),
			Content: systemKustResourceManifest,
		},
		{
			Path:    git.GetSystemQualifiedPath(clusterName, UserKustResourcePath),
			Content: userKustResourceManifest,
		},
		{
			Path:    git.GetSystemQualifiedPath(clusterName, WegoAppPath),
			Content: wegoAppManifest,
		},
		{
			Path:    git.GetSystemQualifiedPath(clusterName, WegoConfigPath),
			Content: wegoConfigManifest,
		},
	}, nil
}

// GitopsManifests generates all yaml files that are going to be written in the config repo
func GitopsManifests(ctx context.Context, fluxClient flux.Flux, gitProvider gitproviders.GitProvider, clusterName string, namespace string, configURL gitproviders.RepoURL) ([]Manifest, error) {
	bootstrapManifest, err := BootstrapManifests(fluxClient, clusterName, namespace, configURL)
	if err != nil {
		return nil, err
	}

	systemKustomization := CreateKustomize(clusterName, namespace, RuntimePath, SourcePath, SystemKustResourcePath, UserKustResourcePath)

	systemKustomizationManifest, err := yaml.Marshal(systemKustomization)
	if err != nil {
		return nil, err
	}

	secretRef, err := GetSecretRefForPrivateGitSources(ctx, gitProvider, configURL)
	if err != nil {
		return nil, err
	}

	configBranch, err := gitProvider.GetDefaultBranch(ctx, configURL)
	if err != nil {
		return nil, err
	}

	secretStr := secretRef.String()
	sourceName := createClusterSourceName(configURL)

	sourceManifest, err := fluxClient.CreateSourceGit(sourceName, configURL, configBranch, secretStr, namespace)
	if err != nil {
		return nil, err
	}

	return append(bootstrapManifest, Manifest{
		Path:    git.GetSystemQualifiedPath(clusterName, SystemKustomizationPath),
		Content: systemKustomizationManifest,
	}, Manifest{ // TODO: Move this to boostrap manifests until getGitClients is refactored
		Path:    git.GetSystemQualifiedPath(clusterName, SourcePath),
		Content: sourceManifest,
	}), nil
}

func CreateKustomize(name, namespace string, resources ...string) types.Kustomization {
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

func createClusterSourceName(gitSourceURL gitproviders.RepoURL) string {
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

func ConstrainResourceName(str string) string {
	return hashNameIfTooLong(replaceUnderscores(str))
}

func workAroundFluxDroppingDot(str string) string {
	return "." + str
}

func gitopsConfigMap(fluxNamespace string, wegoNamespace string) (corev1.ConfigMap, error) {
	config := kube.WegoConfig{
		FluxNamespace: fluxNamespace,
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
