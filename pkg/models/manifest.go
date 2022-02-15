package models

import (
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
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
	WegoProfilesPath        = "profiles.yaml"

	WegoConfigMapName = "weave-gitops-config"
	WegoImage         = "ghcr.io/weaveworks/wego-app"
)

type ManifestsParams struct {
	ClusterName   string
	WegoNamespace string
	ConfigRepo    gitproviders.RepoURL
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
