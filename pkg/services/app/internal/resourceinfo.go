package internal

import (
	"crypto/md5"
	"fmt"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"path/filepath"
	"strings"
)

const maxKubernetesResourceNameLength = 63

func NewResourceInfo(app wego.Application, clusterName string) *AppResourceInfo {
	return &AppResourceInfo{
		Application: app,
		clusterName: clusterName,
		targetName:  clusterName,
	}
}

type AppResourceInfo struct {
	wego.Application
	clusterName string
	targetName  string
}

func (a *AppResourceInfo) ConfigMode() ConfigMode {
	if strings.ToUpper(a.Spec.ConfigURL) == string(ConfigTypeNone) {
		return ConfigModeClusterOnly
	}

	if a.Spec.ConfigURL == string(ConfigTypeUserRepo) || a.Spec.ConfigURL == a.Spec.URL {
		return ConfigModeUserRepo
	}

	return ConfigModeExternalRepo
}

func (a *AppResourceInfo) automationRoot() string {
	root := "."

	if a.ConfigMode() == ConfigModeUserRepo {
		root = ".wego"
	}

	return root
}

func (a *AppResourceInfo) AppYamlPath() string {
	return filepath.Join(a.AppYamlDir(), "app.yaml")
}

func (a *AppResourceInfo) AppYamlDir() string {
	return filepath.Join(a.automationRoot(), "apps", a.Name)
}

func (a *AppResourceInfo) AppAutomationSourcePath() string {
	return filepath.Join(a.AppAutomationDir(), fmt.Sprintf("%s-gitops-source.yaml", a.Name))
}

func (a *AppResourceInfo) AppAutomationDeployPath() string {
	return filepath.Join(a.AppAutomationDir(), fmt.Sprintf("%s-gitops-deploy.yaml", a.Name))
}

func (a *AppResourceInfo) AppAutomationDir() string {
	return filepath.Join(a.automationRoot(), "targets", a.clusterName, a.Name)
}

func (a *AppResourceInfo) AppSourceName() string {
	return a.Name
}

func (a *AppResourceInfo) AppDeployName() string {
	return a.Name
}

func (a *AppResourceInfo) AppResourceName() string {
	return a.Name
}

func (a *AppResourceInfo) AutomationAppsDirKustomizationName() string {
	return HashNameIfTooLong(fmt.Sprintf("%s-apps-dir", a.Name))
}

func (a *AppResourceInfo) AutomationTargetDirKustomizationName() string {
	return HashNameIfTooLong(fmt.Sprintf("%s-%s", a.targetName, a.Name))
}

func (a *AppResourceInfo) SourceKind() ResourceKind {
	result := ResourceKindGitRepository

	if a.Spec.SourceType == "helm" {
		result = ResourceKindHelmRepository
	}

	return result
}

func (a *AppResourceInfo) DeployKind() ResourceKind {
	result := ResourceKindKustomization

	if a.Spec.DeploymentType == "helm" {
		result = ResourceKindHelmRelease
	}

	return result
}

func (a *AppResourceInfo) ClusterResources() []ResourceRef {
	resources := []ResourceRef{}

	// Application GOAT, common to all three modes
	appPath := a.AppYamlPath()
	automationSourcePath := a.AppAutomationSourcePath()
	automationDeployPath := a.AppAutomationDeployPath()

	if a.ConfigMode() == ConfigModeClusterOnly {
		appPath = ""
		automationSourcePath = ""
		automationDeployPath = ""
	}

	resources = append(
		resources,
		ResourceRef{
			Kind:           a.SourceKind(),
			Name:           a.AppSourceName(),
			RepositoryPath: automationSourcePath},
		ResourceRef{
			Kind:           a.DeployKind(),
			Name:           a.AppDeployName(),
			RepositoryPath: automationDeployPath},
		ResourceRef{
			Kind:           ResourceKindApplication,
			Name:           a.AppResourceName(),
			RepositoryPath: appPath})

	// Secret for deploy key associated with app repository;
	// common to all three modes when not using upstream Helm repository
	if a.SourceKind() == ResourceKindGitRepository {
		resources = append(
			resources,
			ResourceRef{
				Kind: ResourceKindSecret,
				Name: a.RepoSecretName(a.Spec.URL).String()})
	}

	if strings.ToUpper(a.Spec.ConfigURL) == string(ConfigTypeNone) {
		// Only app resources present in cluster; no resources to manage config
		return resources
	}

	// App dir and target dir resources are common to app and external repo modes
	resources = append(
		resources,
		// Kustomization for .wego/apps/<app-name> directory
		ResourceRef{
			Kind: ResourceKindKustomization,
			Name: a.AutomationAppsDirKustomizationName()},
		// Kustomization for .wego/targets/<cluster-name>/<app-name> directory
		ResourceRef{
			Kind: ResourceKindKustomization,
			Name: a.AutomationTargetDirKustomizationName()})

	// External repo adds a secret and source for the external repo
	if a.Spec.ConfigURL != string(ConfigTypeUserRepo) && a.Spec.ConfigURL != a.Spec.URL {
		// Config stored in external repo
		resources = append(
			resources,
			// Secret for deploy key associated with config repository
			ResourceRef{
				Kind: ResourceKindSecret,
				Name: a.RepoSecretName(a.Spec.ConfigURL).String()},
			// Source for config repository
			ResourceRef{
				Kind: ResourceKindGitRepository,
				Name: GenerateResourceName(a.Spec.ConfigURL)})
	}

	return resources
}

func (a *AppResourceInfo) ClusterResourcePaths() []string {
	if a.ConfigMode() == ConfigModeClusterOnly {
		return []string{}
	}

	return []string{a.AppYamlPath(), a.AppAutomationSourcePath(), a.AppAutomationDeployPath()}
}

func (a *AppResourceInfo) GetAppHash() string {
	var getHash = func(inputs ...string) string {
		final := []byte(strings.Join(inputs, ""))
		return fmt.Sprintf("%x", md5.Sum(final))
	}

	if a.Spec.DeploymentType == wego.DeploymentTypeHelm {
		return "wego-" + getHash(a.Spec.URL, a.Name, a.Spec.Branch)
	} else {
		return "wego-" + getHash(a.Spec.URL, a.Spec.Path, a.Spec.Branch)
	}
}

type GeneratedSecretName string

func (s GeneratedSecretName) String() string {
	return string(s)
}

func (a *AppResourceInfo) RepoSecretName(repoURL string) GeneratedSecretName {
	return CreateRepoSecretName(a.clusterName, repoURL)
}

func CreateRepoSecretName(targetName string, repoURL string) GeneratedSecretName {
	return GeneratedSecretName(HashNameIfTooLong(fmt.Sprintf("wego-%s-%s", targetName, GenerateResourceName(repoURL))))
}

func GenerateResourceName(url string) string {
	return HashNameIfTooLong(strings.ReplaceAll(utils.UrlToRepoName(url), "_", "-"))
}

func HashNameIfTooLong(name string) string {
	if !NameTooLong(name) {
		return name
	}

	return fmt.Sprintf("wego-%x", md5.Sum([]byte(name)))
}

func NameTooLong(name string) bool {
	return len(name) > maxKubernetesResourceNameLength
}
