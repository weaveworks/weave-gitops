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
	MaxKubernetesResourceNameLength = 63

	WeGOAppIdentifierLabelKey = "wego.weave.works/app-identifier"
)

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
)

type AutomationGenerator interface {
	GenerateApplicationAutomation(ctx context.Context, app models.Application, clusterName string) (ApplicationAutomation, error)
	GenerateClusterAutomation(ctx context.Context, cluster models.Cluster, configURL gitproviders.RepoURL, namespace string) (ClusterAutomation, error)
	GetSecretRefForPrivateGitSources(ctx context.Context, url gitproviders.RepoURL) (GeneratedSecretName, error)
}

type AutomationGen struct {
	GitProvider gitproviders.GitProvider
	Flux        flux.Flux
	Logger      logger.Logger
}

var _ AutomationGenerator = &AutomationGen{}

type ApplicationAutomation struct {
	AppYaml       Manifest
	AppAutomation Manifest
	AppSource     Manifest
	AppKustomize  Manifest
}

type Manifest struct {
	Path    string
	Content []byte
}

type GeneratedSecretName string

func (s GeneratedSecretName) String() string {
	return string(s)
}

func NewAutomationGenerator(gp gitproviders.GitProvider, flux flux.Flux, logger logger.Logger) AutomationGenerator {
	return &AutomationGen{
		GitProvider: gp,
		Flux:        flux,
		Logger:      logger,
	}
}

func (a *AutomationGen) getAppSecretRef(ctx context.Context, app models.Application) (GeneratedSecretName, error) {
	if app.SourceType != models.SourceTypeHelm {
		return a.GetSecretRefForPrivateGitSources(ctx, app.GitSourceURL)
	}

	return "", nil
}

func (a *AutomationGen) GetSecretRefForPrivateGitSources(ctx context.Context, url gitproviders.RepoURL) (GeneratedSecretName, error) {
	var secretRef GeneratedSecretName

	visibility, err := a.GitProvider.GetRepoVisibility(ctx, url)
	if err != nil {
		return "", err
	}

	if *visibility != gitprovider.RepositoryVisibilityPublic {
		secretRef = CreateRepoSecretName(url)
	}

	return secretRef, nil
}

func (a *AutomationGen) generateAppSource(ctx context.Context, app models.Application) (Manifest, error) {
	var (
		source []byte
		err    error
	)

	appSecretRef, err := a.getAppSecretRef(ctx, app)
	if err != nil {
		return Manifest{}, err
	}

	switch app.SourceType {
	case models.SourceTypeGit:
		source, err = a.Flux.CreateSourceGit(app.Name, app.GitSourceURL, app.Branch, appSecretRef.String(), app.Namespace)
		if err == nil {
			source, err = AddWegoIgnore(source)
		}
	case models.SourceTypeHelm:
		source, err = a.Flux.CreateSourceHelm(app.Name, app.HelmSourceURL, app.Namespace)
	default:
		return Manifest{}, fmt.Errorf("unknown source type: %v", app.SourceType)
	}

	if err != nil {
		return Manifest{}, err
	}

	return Manifest{Path: AppAutomationSourcePath(app), Content: source}, nil
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

func createAppKustomize(app models.Application, automation ...Manifest) (Manifest, error) {
	resources := []string{}

	for _, a := range automation {
		resources = append(resources, filepath.Base(a.Path))
	}

	k := CreateKustomize(AppDeployName(app), app.Namespace, resources...)

	bytes, err := yaml.Marshal(k)
	if err != nil {
		return Manifest{}, fmt.Errorf("failed to marshal kustomization for app: %w", err)
	}

	return Manifest{Path: AppAutomationKustomizePath(app), Content: bytes}, nil
}

func AddWegoIgnore(sourceManifest []byte) ([]byte, error) {
	var gitRepository sourcev1.GitRepository

	if err := yaml.Unmarshal(sourceManifest, &gitRepository); err != nil {
		return nil, err
	}

	ignores := []string{automationRoot + "/"}

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

func (a *AutomationGen) GenerateApplicationAutomation(ctx context.Context, app models.Application, clusterName string) (ApplicationAutomation, error) {
	a.Logger.Generatef("Generating application spec manifest")

	appYamlManifest, err := generateAppYaml(app)
	if err != nil {
		return ApplicationAutomation{}, err
	}

	a.Logger.Generatef("Generating GitOps automation manifest")

	appDeployManifest, err := a.generateAppAutomation(ctx, app, clusterName)
	if err != nil {
		return ApplicationAutomation{}, err
	}

	a.Logger.Generatef("Generating GitOps source manifest")

	source, err := a.generateAppSource(ctx, app)
	if err != nil {
		return ApplicationAutomation{}, err
	}

	a.Logger.Generatef("Generating GitOps Kustomization manifest")

	appKustomize, err := createAppKustomize(app, appYamlManifest, appDeployManifest, source)
	if err != nil {
		return ApplicationAutomation{}, err
	}

	return ApplicationAutomation{
		AppYaml:       appYamlManifest,
		AppAutomation: appDeployManifest,
		AppSource:     source,
		AppKustomize:  appKustomize,
	}, nil
}

func (aa ApplicationAutomation) Manifests() []Manifest {
	return append([]Manifest{aa.AppYaml}, aa.AppAutomation, aa.AppSource, aa.AppKustomize)
}

func (a *AutomationGen) generateAppAutomation(ctx context.Context, app models.Application, clusterName string) (Manifest, error) {
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
			return Manifest{}, fmt.Errorf("invalid source type: %v", app.SourceType)
		}
	default:
		return Manifest{}, fmt.Errorf("invalid automation type: %v", app.AutomationType)
	}

	return Manifest{Path: AppAutomationDeployPath(app), Content: sanitizeWegoDirectory(b)}, err
}

func generateAppYaml(app models.Application) (Manifest, error) {
	wegoapp := AppToWegoApp(app)

	wegoapp.ObjectMeta.Labels = map[string]string{
		WeGOAppIdentifierLabelKey: GetAppHash(app),
	}

	b, err := yaml.Marshal(&wegoapp)
	if err != nil {
		return Manifest{}, fmt.Errorf("could not marshal yaml: %w", err)
	}

	return Manifest{Path: AppYamlPath(app), Content: sanitizeK8sYaml(b)}, nil
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

	if models.IsExternalConfigRepo(app.Spec.ConfigRepo) {
		configRepoUrl, err = gitproviders.NewRepoURL(app.Spec.ConfigRepo)
		if err != nil {
			return models.Application{}, err
		}
	}

	return models.Application{
		Name:                app.Name,
		Namespace:           app.Namespace,
		GitSourceURL:        appRepoUrl,
		HelmSourceURL:       helmRepoUrl,
		ConfigRepo:          configRepoUrl,
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
			ConfigRepo:          app.ConfigRepo.String(),
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

const automationRoot = ".weave-gitops"

func AppYamlDir(a models.Application) string {
	return filepath.Join(automationRoot, "apps", a.Name)
}

func AppYamlPath(a models.Application) string {
	return filepath.Join(AppYamlDir(a), "app.yaml")
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

func AutomationUserKustomizePath(clusterName string) string {
	return filepath.Join(automationRoot, "clusters", clusterName, "user", "kustomization.yaml")
}

func AppDeployName(a models.Application) string {
	return a.Name
}

func CreateRepoSecretName(gitSourceURL gitproviders.RepoURL) GeneratedSecretName {
	provider := string(gitSourceURL.Provider())
	cleanRepoName := replaceUnderscores(gitSourceURL.RepositoryName())
	qualifiedName := fmt.Sprintf("wego-%s-%s", provider, cleanRepoName)
	lengthConstrainedName := hashNameIfTooLong(qualifiedName)

	return GeneratedSecretName(lengthConstrainedName)
}

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

func GetAppHash(a models.Application) string {
	var getHash = func(inputs ...string) string {
		final := []byte(strings.Join(inputs, ""))
		return fmt.Sprintf("%x", md5.Sum(final))
	}

	if a.AutomationType == models.AutomationTypeHelm {
		if a.SourceType == models.SourceTypeHelm {
			return "wego-" + getHash(a.HelmSourceURL, a.Name, a.Branch)
		} else {
			return "wego-" + getHash(a.GitSourceURL.String(), a.Name, a.Branch)
		}
	} else {
		return "wego-" + getHash(a.GitSourceURL.String(), a.Path, a.Branch)
	}
}

func GenerateResourceName(url gitproviders.RepoURL) string {
	return ConstrainResourceName(url.RepositoryName())
}

func ConstrainResourceName(str string) string {
	return hashNameIfTooLong(replaceUnderscores(str))
}

func replaceUnderscores(str string) string {
	return strings.ReplaceAll(str, "_", "-")
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
	return len(name) > MaxKubernetesResourceNameLength
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
