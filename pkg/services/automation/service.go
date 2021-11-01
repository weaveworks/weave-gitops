package automation

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/fluxcd/source-controller/pkg/sourceignore"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/models"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

const (
	maxKubernetesResourceNameLength = 63

	WeGOAppIdentifierLabelKey = "wego.weave.works/app-identifier"
)

// type AddParams struct {
//  Dir                        string
//  Name                       string
//  Url                        string
//  Path                       string
//  Branch                     string
//  DeploymentType             string
//  Chart                      string
//  SourceType                 wego.SourceType
//  AppConfigUrl               string
//  Namespace                  string
//  DryRun                     bool
//  AutoMerge                  bool
//  GitProviderToken           string
//  HelmReleaseTargetNamespace string
//  MigrateToNewDirStructure   func(string) string
// }

type ConfigMode string

type SourceType string

type ResourceKind string

type ResourceRef struct {
	Kind           ResourceKind
	Name           string
	RepositoryPath string
}

const (
	ConfigModeClusterOnly  ConfigMode = "clusterOnly"
	ConfigModeUserRepo     ConfigMode = "userRepo"
	ConfigModeExternalRepo ConfigMode = "externalRepo"

	ResourceKindApplication    ResourceKind = "Application"
	ResourceKindSecret         ResourceKind = "Secret"
	ResourceKindGitRepository  ResourceKind = "GitRepository"
	ResourceKindHelmRepository ResourceKind = "HelmRepository"
	ResourceKindKustomization  ResourceKind = "Kustomization"
	ResourceKindHelmRelease    ResourceKind = "HelmRelease"
	ResourceKindKustomize      ResourceKind = "Kustomize"
)

var defaultMigrateToNewDirStructure func(string) string = func(s string) string { return s }

type AutomationService interface {
	GenerateAutomation(ctx context.Context, app models.Application, clusterName string) ([]AutomationManifest, error)
}

type AutomationSvc struct {
	GitProvider gitproviders.GitProvider
	Flux        flux.Flux
	Logger      logger.Logger
}

var _ AutomationService = &AutomationSvc{}

type AutomationManifest struct {
	Path    string
	Content []byte
}

func NewAutomationService(gp gitproviders.GitProvider, flux flux.Flux, logger logger.Logger) AutomationService {
	return &AutomationSvc{
		GitProvider: gp,
		Flux:        flux,
		Logger:      logger,
	}
}

func (a *AutomationSvc) getAppSecretRef(ctx context.Context, app models.Application, clusterName string) (GeneratedSecretName, error) {
	if app.SourceType != models.SourceTypeHelm {
		return a.getSecretRef(ctx, app, app.GitSourceURL, clusterName)
	}

	return "", nil
}

func (a *AutomationSvc) getConfigSecretRef(ctx context.Context, app models.Application, clusterName string) (GeneratedSecretName, error) {
	return a.getSecretRef(ctx, app, app.ConfigURL, clusterName)
}

func (a *AutomationSvc) getSecretRef(ctx context.Context, app models.Application, url gitproviders.RepoURL, clusterName string) (GeneratedSecretName, error) {
	var secretRef GeneratedSecretName

	visibility, visibilityErr := a.GitProvider.GetRepoVisibility(ctx, url)
	if visibilityErr != nil {
		return "", visibilityErr
	}

	if *visibility != gitprovider.RepositoryVisibilityPublic {
		secretRef = RepoSecretName(app, url, clusterName)
	}

	return secretRef, nil
}

func (a *AutomationSvc) generateAppAutomation(ctx context.Context, app models.Application, clusterName string) ([]AutomationManifest, error) {
	a.Logger.Generatef("Generating GitOps automation manifests")

	appGoatManifests, err := a.generateApplicationGoat(app, clusterName)
	if err != nil {
		return nil, fmt.Errorf("could not create GitOps automation for '%s': %w", app.Name, err)
	}

	a.Logger.Generatef("Generating application spec manifest")

	appManifests, err := generateAppYaml(app)
	if err != nil {
		return nil, fmt.Errorf("could not create app.yaml for '%s': %w", app.Name, err)
	}

	return append(appGoatManifests, appManifests...), nil
}

func (a *AutomationSvc) generateAppSource(ctx context.Context, app models.Application, clusterName string) (AutomationManifest, error) {
	var source []byte
	var err error

	appSecretRef, err := a.getAppSecretRef(ctx, app, clusterName)
	if err != nil {
		return AutomationManifest{}, err
	}

	switch app.SourceType {
	case models.SourceTypeGit:
		source, err = a.Flux.CreateSourceGit(app.Name, app.GitSourceURL.String(), app.Branch, appSecretRef.String(), app.Namespace)
		if err == nil {
			source, err = addWegoIgnore(source)
		}
	case models.SourceTypeHelm:
		source, err = a.Flux.CreateSourceHelm(app.Name, app.HelmSourceURL, app.Namespace)
	default:
		return AutomationManifest{}, fmt.Errorf("unknown source type: %v", app.SourceType)
	}

	if err != nil {
		return AutomationManifest{}, err
	}

	return AutomationManifest{Path: AppAutomationSourcePath(app), Content: source}, nil
}

func GetOrCreateKustomize(filename, name, namespace string) (types.Kustomization, error) {
	var k types.Kustomization

	contents, err := os.ReadFile(filename)
	if err == nil {
		if err := yaml.Unmarshal(contents, &k); err != nil {
			return k, fmt.Errorf("failed to read existing kustomize file %s: %w", filename, err)
		}

		return k, nil
	}

	return CreateKustomize(name, namespace), nil
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

func createAppKustomize(app models.Application, automation []AutomationManifest) (AutomationManifest, error) {
	resources := []string{}

	for _, a := range automation {
		resources = append(resources, filepath.Base(a.Path))
	}

	k := CreateKustomize(AppDeployName(app), app.Namespace, resources...)

	bytes, err := yaml.Marshal(k)
	if err != nil {
		return AutomationManifest{}, fmt.Errorf("failed to marshal kustomization for app: %w", err)
	}

	return AutomationManifest{Path: AppAutomationKustomizePath(app), Content: bytes}, nil
}

func addWegoIgnore(sourceManifest []byte) ([]byte, error) {
	var gitRepository sourcev1.GitRepository

	if err := yaml.Unmarshal(sourceManifest, &gitRepository); err != nil {
		return nil, err
	}

	ignores := []string{".weave-gitops/"}

	for _, ignore := range []string{sourceignore.ExcludeVCS, sourceignore.ExcludeExt, sourceignore.ExcludeCI, sourceignore.ExcludeExtra} {
		ignores = append(ignores, strings.Split(ignore, ",")...)
	}

	ignoreSpec := strings.Join(ignores, "\n")
	gitRepository.Spec.Ignore = &ignoreSpec

	updatedManifest, err := yaml.Marshal(gitRepository)
	if err != nil {
		return nil, err
	}

	return updatedManifest, nil
}

// func (a *AutomationSvc) generateAutomationAutomation(ctx context.Context, app models.Application, clusterName string) ([]AutomationManifest, error) {
//  secretName, err := a.getConfigSecretRef(ctx, app, clusterName)
//  if err != nil {
//      return nil, err
//  }

//  appsDirAutomationManifest, err := a.Flux.CreateKustomization(
//      AutomationAppsDirKustomizationName(app),
//      secretName.String(),
//      AppYamlDir(app),
//      app.Namespace)
//  if err != nil {
//      return nil, fmt.Errorf("could not generate app dir kustomization for '%s': %w", app.Name, err)
//  }

//  targetDirAutomationManifest, err := a.Flux.CreateKustomization(
//      AutomationTargetDirKustomizationName(app, clusterName),
//      secretName.String(),
//      AppAutomationDir(app, clusterName),
//      app.Namespace)
//  if err != nil {
//      return nil, fmt.Errorf("could not generate target dir kustomization for '%s': %w", app.Name, err)
//  }

//  return []AutomationManifest{
//      AutomationManifest{Path: "", Content: appsDirAutomationManifest},
//      AutomationManifest{Path: "", Content: targetDirAutomationManifest},
//  }, nil
// }

func (a *AutomationSvc) GenerateAutomation(ctx context.Context, app models.Application, clusterName string) ([]AutomationManifest, error) {
	appDeployManifests, err := a.generateAppAutomation(ctx, app, clusterName)
	if err != nil {
		return nil, err
	}

	source, err := a.generateAppSource(ctx, app, clusterName)
	if err != nil {
		return nil, err
	}

	automationManifests := append(appDeployManifests, source)

	appKustomize, err := createAppKustomize(app, automationManifests)
	if err != nil {
		return nil, err
	}

	return append(automationManifests, appKustomize), nil
}

func (a *AutomationSvc) generateApplicationGoat(app models.Application, clusterName string) ([]AutomationManifest, error) {
	var (
		b   []byte
		err error
	)

	switch app.AutomationType {
	case models.AutomationTypeKustomize:
		b, err = a.Flux.CreateKustomization(app.Name, app.Name, app.Path, app.Namespace)
	case models.AutomationTypeHelm:
		switch app.SourceType {
		case models.SourceTypeHelm:
			b, err = a.Flux.CreateHelmReleaseHelmRepository(app.Name, app.Path, app.Namespace, app.HelmTargetNamespace)
		case models.SourceTypeGit:
			b, err = a.Flux.CreateHelmReleaseGitRepository(app.Name, app.Name, app.Path, app.Namespace, app.HelmTargetNamespace)
		default:
			return nil, fmt.Errorf("invalid source type: %v", app.SourceType)
		}
	default:
		return nil, fmt.Errorf("invalid automation type: %v", app.AutomationType)
	}

	return []AutomationManifest{AutomationManifest{Path: AppAutomationDeployPath(app), Content: sanitizeWegoDirectory(b)}}, err
}

func generateAppYaml(app models.Application) ([]AutomationManifest, error) {
	wegoapp := AppToWegoApp(app)

	wegoapp.ObjectMeta.Labels = map[string]string{
		WeGOAppIdentifierLabelKey: GetAppHash(app),
	}

	b, err := yaml.Marshal(&wegoapp)
	if err != nil {
		return nil, fmt.Errorf("could not marshal yaml: %w", err)
	}

	return []AutomationManifest{AutomationManifest{Path: AppYamlPath(app), Content: sanitizeK8sYaml(b)}}, nil
}

func WegoAppToApp(app wego.Application) (models.Application, error) {
	var (
		helmRepoUrl   string
		appRepoUrl    gitproviders.RepoURL
		configRepoUrl gitproviders.RepoURL
		err           error
	)

	if wego.SourceType(app.Spec.SourceType) == wego.SourceType(wego.SourceTypeGit) {
		appRepoUrl, err = gitproviders.NewRepoURL(app.Spec.URL)
		if err != nil {
			return models.Application{}, err
		}
	} else {
		helmRepoUrl = app.Spec.URL
	}

	if models.IsExternalConfigUrl(app.Spec.ConfigURL) {
		configRepoUrl, err = gitproviders.NewRepoURL(app.Spec.ConfigURL)
		if err != nil {
			return models.Application{}, err
		}
	}

	return models.Application{
		Name:                app.Name,
		Namespace:           app.Namespace,
		GitSourceURL:        appRepoUrl,
		HelmSourceURL:       helmRepoUrl,
		ConfigURL:           configRepoUrl,
		Branch:              app.Spec.Branch,
		Path:                app.Spec.Path,
		HelmTargetNamespace: app.Spec.HelmTargetNamespace,
		SourceType:          models.SourceType(app.Spec.SourceType),
		AutomationType:      models.AutomationType(app.Spec.DeploymentType),
	}, nil
}

func AppToWegoApp(app models.Application) wego.Application {
	sourceUrl := app.GitSourceURL.String()

	if app.SourceType == models.SourceTypeHelm {
		sourceUrl = app.HelmSourceURL
	}

	gvk := wego.GroupVersion.WithKind(wego.ApplicationKind)
	wegoApp := wego.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
		Spec: wego.ApplicationSpec{
			ConfigURL:           app.ConfigURL.String(),
			Branch:              app.Branch,
			URL:                 sourceUrl,
			Path:                app.Path,
			DeploymentType:      wego.DeploymentType(string(app.AutomationType)),
			SourceType:          wego.SourceType(string(app.SourceType)),
			HelmTargetNamespace: app.HelmTargetNamespace,
		},
	}

	return wegoApp
}

// Operations to extract useful information from an Application

func GetConfigMode(a models.Application) ConfigMode {
	if a.GitSourceURL.String() == a.ConfigURL.String() {
		return ConfigModeUserRepo
	}

	return ConfigModeExternalRepo
}

func automationRoot(a models.Application) string {
	return ".weave-gitops"
}

func AppYamlPath(a models.Application) string {
	return filepath.Join(AppYamlDir(a), "app.yaml")
}

func AppYamlDir(a models.Application) string {
	return filepath.Join(automationRoot(a), "apps", a.Name)
}

func AppAutomationSourcePath(a models.Application) string {
	return filepath.Join(AppYamlDir(a), fmt.Sprintf("%s-gitops-source.yaml", a.Name))
}

func AppAutomationDeployPath(a models.Application) string {
	return filepath.Join(AppYamlDir(a), fmt.Sprintf("%s-gitops-deploy.yaml", a.Name))
}

func AppAutomationKustomizePath(a models.Application) string {
	return filepath.Join(AppYamlDir(a), "kustomization.yaml")
}

func AppSourceName(a models.Application) string {
	return a.Name
}

func AppDeployName(a models.Application) string {
	return a.Name
}

func AppResourceName(a models.Application) string {
	return a.Name
}

type GeneratedSecretName string

func (s GeneratedSecretName) String() string {
	return string(s)
}

func RepoSecretName(a models.Application, gitSourceURL gitproviders.RepoURL, clusterName string) GeneratedSecretName {
	return CreateRepoSecretName(clusterName, gitSourceURL)
}

func CreateRepoSecretName(targetName string, gitSourceURL gitproviders.RepoURL) GeneratedSecretName {
	return GeneratedSecretName(hashNameIfTooLong(fmt.Sprintf("wego-%s-%s", targetName, GenerateResourceName(gitSourceURL))))
}

func AutomationAppsDirKustomizationName(a models.Application) string {
	return hashNameIfTooLong(fmt.Sprintf("%s-apps-dir", a.Name))
}

// func AutomationTargetDirKustomizationName(a models.Application, clusterName string) string {
//  return hashNameIfTooLong(fmt.Sprintf("%s-%s", clusterName, a.Name))
// }

func SourceKind(a models.Application) ResourceKind {
	result := ResourceKindGitRepository

	if a.SourceType == models.SourceTypeHelm {
		result = ResourceKindHelmRepository
	}

	return result
}

func DeployKind(a models.Application) ResourceKind {
	result := ResourceKindKustomization

	if a.AutomationType == models.AutomationTypeHelm {
		result = ResourceKindHelmRelease
	}

	return result
}

func ClusterResources(a models.Application, clusterName string) []ResourceRef {
	resources := []ResourceRef{}

	// Application GOAT, common to all three modes
	appPath := AppYamlPath(a)
	automationSourcePath := AppAutomationSourcePath(a)
	automationDeployPath := AppAutomationDeployPath(a)
	automationKustomizePath := AppAutomationKustomizePath(a)

	resources = append(
		resources,
		ResourceRef{
			Kind:           SourceKind(a),
			Name:           AppSourceName(a),
			RepositoryPath: automationSourcePath},
		ResourceRef{
			Kind:           DeployKind(a),
			Name:           AppDeployName(a),
			RepositoryPath: automationDeployPath},
		ResourceRef{
			Kind:           ResourceKindKustomize,
			Name:           "kustomization.yaml",
			RepositoryPath: automationKustomizePath},
		ResourceRef{
			Kind:           ResourceKindApplication,
			Name:           AppResourceName(a),
			RepositoryPath: appPath})

	if a.ConfigURL.URL() == nil {
		// Only app resources present in cluster; no resources to manage config
		return resources
	}

	return resources
}

func ClusterResourcePaths(a models.Application, clusterName string) []string {
	return []string{AppYamlPath(a), AppAutomationSourcePath(a), AppAutomationDeployPath(a)}
}

func GetAppHash(a models.Application) string {
	var getHash = func(inputs ...string) string {
		final := []byte(strings.Join(inputs, ""))
		return fmt.Sprintf("%x", md5.Sum(final))
	}

	if a.AutomationType == models.AutomationTypeHelm {
		return "wego-" + getHash(a.GitSourceURL.String(), a.Name, a.Branch)
	} else {
		return "wego-" + getHash(a.GitSourceURL.String(), a.Path, a.Branch)
	}
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
	case ResourceKindKustomize:
		return schema.GroupVersionResource{}, nil
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

// Remove some problematic fields before saving the yaml files.
// K8s/reconcilers will populate these fields after creation.
// https://github.com/fluxcd/flux2/blob/0ae39d5a0a5220c177b29e71fc8824babd1e0d7c/cmd/flux/export.go#L111
func sanitizeK8sYaml(data []byte) []byte {
	out := []byte("---\n")
	data = bytes.Replace(data, []byte("  creationTimestamp: null\n"), []byte(""), 1)
	data = bytes.Replace(data, []byte("status: {}\n"), []byte(""), 1)

	return append(out, data...)
}

func sanitizeWegoDirectory(manifest []byte) []byte {
	return bytes.ReplaceAll(manifest, []byte("path: ./weave-gitops"), []byte("path: .weave-gitops"))
}
