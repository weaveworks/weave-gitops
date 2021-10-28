package models

import (
	"crypto/md5"
	"fmt"
	"path/filepath"
	"strings"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ConfigType string

type ConfigMode string

type ResourceKind string

type ResourceRef struct {
	Kind           ResourceKind
	Name           string
	RepositoryPath string
}

const maxKubernetesResourceNameLength = 63

const (
	ConfigTypeUserRepo ConfigType = ""
	ConfigTypeNone     ConfigType = "NONE"

	ConfigModeClusterOnly  ConfigMode = "clusterOnly"
	ConfigModeUserRepo     ConfigMode = "userRepo"
	ConfigModeExternalRepo ConfigMode = "externalRepo"

	ResourceKindApplication    ResourceKind = "Application"
	ResourceKindSecret         ResourceKind = "Secret"
	ResourceKindGitRepository  ResourceKind = "GitRepository"
	ResourceKindHelmRepository ResourceKind = "HelmRepository"
	ResourceKindKustomization  ResourceKind = "Kustomization"
	ResourceKindHelmRelease    ResourceKind = "HelmRelease"
)

type Application struct {
	wego.Application
	targetName    string
	AppRepoUrl    gitproviders.RepoURL
	ConfigRepoUrl gitproviders.RepoURL
}

func NewApplication(clusterApplication wego.Application) (Application, error) {
	var (
		appRepoUrl    gitproviders.RepoURL
		configRepoUrl gitproviders.RepoURL
		err           error
	)

	if wego.DeploymentType(clusterApplication.Spec.SourceType) == wego.DeploymentType(wego.SourceTypeGit) {
		appRepoUrl, err = gitproviders.NewRepoURL(clusterApplication.Spec.URL)
		if err != nil {
			return Application{}, err
		}
	}

	if IsExternalConfigUrl(clusterApplication.Spec.ConfigURL) {
		configRepoUrl, err = gitproviders.NewRepoURL(clusterApplication.Spec.ConfigURL)
		if err != nil {
			return Application{}, err
		}
	}

	return Application{
		Application:   clusterApplication,
		AppRepoUrl:    appRepoUrl,
		ConfigRepoUrl: configRepoUrl,
	}, nil
}

func (a Application) ConfigMode() ConfigMode {
	if strings.ToUpper(a.Spec.ConfigURL) == string(ConfigTypeNone) {
		return ConfigModeClusterOnly
	}

	if a.Spec.ConfigURL == string(ConfigTypeUserRepo) || a.Spec.ConfigURL == a.Spec.URL {
		return ConfigModeUserRepo
	}

	return ConfigModeExternalRepo
}

func (a Application) automationRoot() string {
	root := "."

	if a.ConfigMode() == ConfigModeUserRepo {
		root = ".wego"
	}

	return root
}

func (a Application) AppYamlPath() string {
	return filepath.Join(a.AppYamlDir(), "app.yaml")
}

func (a Application) AppYamlDir() string {
	return filepath.Join(a.automationRoot(), "apps", a.Name)
}

func (a Application) AppAutomationSourcePath(clusterName string) string {
	return filepath.Join(a.AppAutomationDir(clusterName), fmt.Sprintf("%s-gitops-source.yaml", a.Name))
}

func (a Application) AppAutomationDeployPath(clusterName string) string {
	return filepath.Join(a.AppAutomationDir(clusterName), fmt.Sprintf("%s-gitops-deploy.yaml", a.Name))
}

func (a Application) AppAutomationDir(clusterName string) string {
	return filepath.Join(a.automationRoot(), "targets", clusterName, a.Name)
}

func (a Application) AppSourceName() string {
	return a.Name
}

func (a Application) AppDeployName() string {
	return a.Name
}

func (a Application) AppResourceName() string {
	return a.Name
}

type GeneratedSecretName string

func (s GeneratedSecretName) String() string {
	return string(s)
}

func (a Application) RepoSecretName(repoURL gitproviders.RepoURL, clusterName string) GeneratedSecretName {
	return CreateRepoSecretName(clusterName, repoURL)
}

func CreateRepoSecretName(targetName string, repoURL gitproviders.RepoURL) GeneratedSecretName {
	return GeneratedSecretName(hashNameIfTooLong(fmt.Sprintf("wego-%s-%s", targetName, GenerateResourceName(repoURL))))
}

func (a Application) AutomationAppsDirKustomizationName() string {
	return hashNameIfTooLong(fmt.Sprintf("%s-apps-dir", a.Name))
}

func (a Application) AutomationTargetDirKustomizationName(clusterName string) string {
	return hashNameIfTooLong(fmt.Sprintf("%s-%s", clusterName, a.Name))
}

func (a Application) SourceKind() ResourceKind {
	result := ResourceKindGitRepository

	if a.Spec.SourceType == "helm" {
		result = ResourceKindHelmRepository
	}

	return result
}

func (a Application) DeployKind() ResourceKind {
	result := ResourceKindKustomization

	if a.Spec.DeploymentType == "helm" {
		result = ResourceKindHelmRelease
	}

	return result
}

func (a Application) ClusterResources(clusterName string) []ResourceRef {
	resources := []ResourceRef{}

	// Application GOAT, common to all three modes
	appPath := a.AppYamlPath()
	automationSourcePath := a.AppAutomationSourcePath(clusterName)
	automationDeployPath := a.AppAutomationDeployPath(clusterName)

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
				Name: a.RepoSecretName(a.AppRepoUrl, clusterName).String()})
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
			Name: a.AutomationTargetDirKustomizationName(clusterName)})

	// External repo adds a secret and source for the external repo
	if a.ConfigMode() != ConfigModeUserRepo {
		// Config stored in external repo
		resources = append(
			resources,
			// Secret for deploy key associated with config repository
			ResourceRef{
				Kind: ResourceKindSecret,
				Name: a.RepoSecretName(a.ConfigRepoUrl, clusterName).String()},
			// Source for config repository
			ResourceRef{
				Kind: ResourceKindGitRepository,
				Name: GenerateResourceName(a.ConfigRepoUrl)})
	}

	return resources
}

func (a Application) ClusterResourcePaths(clusterName string) []string {
	if a.ConfigMode() == ConfigModeClusterOnly {
		return []string{}
	}

	return []string{a.AppYamlPath(), a.AppAutomationSourcePath(clusterName), a.AppAutomationDeployPath(clusterName)}
}

func (info Application) GetAppHash() string {
	var getHash = func(inputs ...string) string {
		final := []byte(strings.Join(inputs, ""))
		return fmt.Sprintf("%x", md5.Sum(final))
	}

	if info.Spec.DeploymentType == wego.DeploymentTypeHelm {
		return "wego-" + getHash(info.Spec.URL, info.Name, info.Spec.Branch)
	} else {
		return "wego-" + getHash(info.Spec.URL, info.Spec.Path, info.Spec.Branch)
	}
}

func (app Application) GetConfigUrl() (gitproviders.RepoURL, error) {
	switch app.ConfigMode() {
	case ConfigModeExternalRepo:
		return app.ConfigRepoUrl, nil
	case ConfigModeUserRepo:
		return app.AppRepoUrl, nil
	default:
		return gitproviders.RepoURL{}, fmt.Errorf("Application %q has no configuration repository", app.Name)
	}
}

func IsExternalConfigUrl(url string) bool {
	return strings.ToUpper(url) != string(ConfigTypeNone) &&
		strings.ToUpper(url) != string(ConfigTypeUserRepo)
}

func GenerateResourceName(url gitproviders.RepoURL) string {
	return hashNameIfTooLong(strings.ReplaceAll(url.RepositoryName(), "_", "-"))
}

func (rk ResourceKind) ToGVR() (schema.GroupVersionResource, error) {
	switch rk {
	case ResourceKindApplication:
		return kube.GVRApp, nil
	case ResourceKindSecret:
		return kube.GVRSecret, nil
	case ResourceKindGitRepository:
		return kube.GVRGitRepository, nil
	case ResourceKindHelmRepository:
		return kube.GVRHelmRepository, nil
	case ResourceKindHelmRelease:
		return kube.GVRHelmRelease, nil
	case ResourceKindKustomization:
		return kube.GVRKustomization, nil
	default:
		return schema.GroupVersionResource{}, fmt.Errorf("no matching schema.GroupVersionResource to the ResourceKind: %s", string(rk))
	}
}

func ApplicationNameTooLong(name string) bool {
	return len(name) > maxKubernetesResourceNameLength
}

func hashNameIfTooLong(name string) string {
	if !ApplicationNameTooLong(name) {
		return name
	}

	return fmt.Sprintf("wego-%x", md5.Sum([]byte(name)))
}
