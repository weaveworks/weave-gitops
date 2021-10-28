package automation

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/fluxcd/source-controller/pkg/sourceignore"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/models"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"sigs.k8s.io/yaml"
)

const (
	WeGOAppIdentifierLabelKey = "wego.weave.works/app-identifier"
)

type AddParams struct {
	Dir                        string
	Name                       string
	Url                        string
	Path                       string
	Branch                     string
	DeploymentType             string
	Chart                      string
	SourceType                 wego.SourceType
	AppConfigUrl               string
	Namespace                  string
	DryRun                     bool
	AutoMerge                  bool
	GitProviderToken           string
	HelmReleaseTargetNamespace string
	MigrateToNewDirStructure   func(string) string
}

var defaultMigrateToNewDirStructure func(string) string = func(s string) string { return s }

type AutomationService interface {
	GenerateManifests(ctx context.Context, app models.Application, clusterName string) ([]AutomationManifest, error)
}

type AutomationSvc struct {
	GitProvider gitproviders.GitProvider
	Flux        flux.Flux
	Logger      logger.Logger
}

type AutomationManifest struct {
	Path     string
	Manifest []byte
}

func NewAutomationService(gp gitproviders.GitProvider, flux flux.Flux, logger logger.Logger) AutomationService {
	return &AutomationSvc{
		GitProvider: gp,
		Logger:      logger,
	}
}

func (a *AutomationSvc) getAppSecretRef(ctx context.Context, app models.Application, clusterName string) (models.GeneratedSecretName, error) {
	if app.Spec.SourceType != wego.SourceTypeHelm {
		return a.getSecretRef(ctx, app, app.AppRepoUrl, clusterName)
	}

	return "", nil
}

func (a *AutomationSvc) getConfigSecretRef(ctx context.Context, app models.Application, clusterName string) (models.GeneratedSecretName, error) {
	configUrl, err := app.GetConfigUrl()
	if err != nil {
		return "", err
	}

	return a.getSecretRef(ctx, app, configUrl, clusterName)
}

func (a *AutomationSvc) getSecretRef(ctx context.Context, app models.Application, url gitproviders.RepoURL, clusterName string) (models.GeneratedSecretName, error) {
	var secretRef models.GeneratedSecretName

	visibility, visibilityErr := a.GitProvider.GetRepoVisibility(ctx, url)
	if visibilityErr != nil {
		return "", visibilityErr
	}

	if *visibility != gitprovider.RepositoryVisibilityPublic {
		secretRef = app.RepoSecretName(url, clusterName)
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

func (a *AutomationSvc) generateSources(ctx context.Context, app models.Application, clusterName string) ([]AutomationManifest, error) {
	appSources, err := a.generateAppSources(ctx, app, clusterName)
	if err != nil {
		return nil, err
	}

	automationSources, err := a.generateAutomationSources(ctx, app, clusterName)
	if err != nil {
		return nil, err
	}

	return append(appSources, automationSources...), nil
}

func (a *AutomationSvc) generateAppSources(ctx context.Context, app models.Application, clusterName string) ([]AutomationManifest, error) {
	var source []byte
	var err error

	appSecretRef, err := a.getAppSecretRef(ctx, app, clusterName)
	if err != nil {
		return nil, err
	}

	switch app.Spec.SourceType {
	case wego.SourceTypeGit:
		source, err = a.Flux.CreateSourceGit(app.Name, app.Spec.URL, app.Spec.Branch, appSecretRef.String(), app.Namespace)
		if err == nil {
			source, err = addWegoIgnore(source)
		}
	case wego.SourceTypeHelm:
		source, err = a.Flux.CreateSourceHelm(app.Name, app.Spec.URL, app.Namespace)
	default:
		return nil, fmt.Errorf("unknown source type: %v", app.Spec.SourceType)
	}

	if err != nil {
		return nil, err
	}

	return []AutomationManifest{AutomationManifest{Path: app.AppAutomationSourcePath(clusterName), Manifest: source}}, nil
}

func (a *AutomationSvc) generateAutomationSources(ctx context.Context, app models.Application, clusterName string) ([]AutomationManifest, error) {
	configSecretRef, err := a.getConfigSecretRef(ctx, app, clusterName)
	if err != nil {
		return nil, err
	}

	if app.ConfigMode() == models.ConfigModeClusterOnly {
		return []AutomationManifest{}, nil
	}

	configUrl, err := app.GetConfigUrl()
	if err != nil {
		return nil, err
	}

	configBranch, err := a.GitProvider.GetDefaultBranch(ctx, configUrl)
	if err != nil {
		return nil, fmt.Errorf("could not determine default branch for config repository: %w", err)
	}

	automationSource, err := a.Flux.CreateSourceGit(models.GenerateResourceName(configUrl), configUrl.String(), configBranch, configSecretRef.String(), app.Namespace)
	if err != nil {
		return nil, fmt.Errorf("could not generate automation source manifest: %w", err)
	}

	return []AutomationManifest{
		AutomationManifest{
			Path:     "", // For now, the automation managing manifests aren't stored in the repository
			Manifest: automationSource,
		},
	}, nil
}

func addWegoIgnore(sourceManifest []byte) ([]byte, error) {
	var gitRepositorySpec sourcev1.GitRepositorySpec

	if err := yaml.Unmarshal(sourceManifest, &gitRepositorySpec); err != nil {
		return nil, err
	}

	ignoreSpec := strings.Join([]string{sourceignore.ExcludeVCS, sourceignore.ExcludeExt, sourceignore.ExcludeCI, sourceignore.ExcludeExtra, "/.wego/"}, ",")
	gitRepositorySpec.Ignore = &ignoreSpec

	updatedManifest, err := yaml.Marshal(gitRepositorySpec)
	if err != nil {
		return nil, err
	}

	return updatedManifest, nil
}

func (a *AutomationSvc) generateAutomationAutomation(ctx context.Context, app models.Application, clusterName string) ([]AutomationManifest, error) {
	secretName, err := a.getConfigSecretRef(ctx, app, clusterName)
	if err != nil {
		return nil, err
	}

	appsDirAutomationManifest, err := a.Flux.CreateKustomization(
		app.AutomationAppsDirKustomizationName(),
		secretName.String(),
		app.AppYamlDir(),
		app.Namespace)
	if err != nil {
		return nil, fmt.Errorf("could not generate app dir kustomization for '%s': %w", app.Name, err)
	}

	targetDirAutomationManifest, err := a.Flux.CreateKustomization(
		app.AutomationTargetDirKustomizationName(clusterName),
		secretName.String(),
		app.AppAutomationDir(clusterName),
		app.Namespace)
	if err != nil {
		return nil, fmt.Errorf("could not generate target dir kustomization for '%s': %w", app.Name, err)
	}

	return []AutomationManifest{
		AutomationManifest{Path: app.AppAutomationSourcePath(clusterName), Manifest: appsDirAutomationManifest},
		AutomationManifest{Path: app.AppAutomationDeployPath(clusterName), Manifest: targetDirAutomationManifest},
	}, nil
}

func (a *AutomationSvc) GenerateManifests(ctx context.Context, app models.Application, clusterName string) ([]AutomationManifest, error) {
	sources, err := a.generateSources(ctx, app, clusterName)
	if err != nil {
		return nil, err
	}

	appAutomationManifests, err := a.generateAppAutomation(ctx, app, clusterName)
	if err != nil {
		return nil, err
	}

	automationAutomationManifests, err := a.generateAutomationAutomation(ctx, app, clusterName)
	if err != nil {
		return nil, err
	}

	return append(append(appAutomationManifests, sources...), automationAutomationManifests...), nil
}

func (a *AutomationSvc) generateApplicationGoat(app models.Application, clusterName string) ([]AutomationManifest, error) {
	var (
		b   []byte
		err error
	)

	switch app.Spec.DeploymentType {
	case wego.DeploymentTypeKustomize:
		b, err = a.Flux.CreateKustomization(app.Name, app.Name, app.Spec.Path, app.Namespace)
	case wego.DeploymentTypeHelm:
		switch app.Spec.SourceType {
		case wego.SourceTypeHelm:
			b, err = a.Flux.CreateHelmReleaseHelmRepository(app.Name, app.Spec.Path, app.Namespace, app.Spec.HelmTargetNamespace)
		case wego.SourceTypeGit:
			b, err = a.Flux.CreateHelmReleaseGitRepository(app.Name, app.Name, app.Spec.Path, app.Namespace, app.Spec.HelmTargetNamespace)
		default:
			return nil, fmt.Errorf("invalid source type: %v", app.Spec.SourceType)
		}
	default:
		return nil, fmt.Errorf("invalid deployment type: %v", app.Spec.DeploymentType)
	}

	return []AutomationManifest{AutomationManifest{Path: app.AppAutomationDeployPath(clusterName), Manifest: sanitizeWegoDirectory(b)}}, err
}

func generateAppYaml(app models.Application) ([]AutomationManifest, error) {
	wegoapp := appToWegoApp(app)

	wegoapp.ObjectMeta.Labels = map[string]string{
		WeGOAppIdentifierLabelKey: app.GetAppHash(),
	}

	b, err := yaml.Marshal(&wegoapp)
	if err != nil {
		return nil, fmt.Errorf("could not marshal yaml: %w", err)
	}

	return []AutomationManifest{AutomationManifest{Path: app.AppYamlPath(), Manifest: sanitizeK8sYaml(b)}}, nil
}

func appToWegoApp(app models.Application) wego.Application {
	return app.Application
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
	return bytes.ReplaceAll(manifest, []byte("path: ./wego"), []byte("path: .wego"))
}
