package app

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/status"
)

type DeploymentType string
type SourceType string
type ConfigType string

const (
	DeployTypeKustomize DeploymentType = "kustomize"
	DeployTypeHelm      DeploymentType = "helm"

	SourceTypeGit  SourceType = "git"
	SourceTypeHelm SourceType = "helm"

	ConfigTypeUserRepo ConfigType = ""
	ConfigTypeNone     ConfigType = "NONE"
)

type Params struct {
	Dir            string
	Name           string
	Url            string
	Path           string
	Branch         string
	PrivateKey     string
	PrivateKeyPass string
	DeploymentType string
	Chart          string
	SourceType     string
	AppConfigUrl   string
	Namespace      string
	DryRun         bool
}

type Dependencies struct {
	Git  git.Git
	Flux flux.Flux
	Kube kube.Kube
}

type AppService interface {
	Add()
}

type App struct {
	git  git.Git
	flux flux.Flux
	kube kube.Kube
}

func New(deps *Dependencies) *App {
	return &App{
		git:  deps.Git,
		flux: deps.Flux,
		kube: deps.Kube,
	}
}

func (a *App) Add(params Params) error {
	fmt.Print("Updating parameters from environment... ")
	params, err := a.updateParametersIfNecessary(params)
	if err != nil {
		return errors.Wrap(err, "could not update parameters")
	}

	fmt.Print("done\n\n")
	fmt.Print("Checking cluster status... ")

	clusterStatus := status.GetClusterStatus()
	fmt.Printf("%s\n\n", clusterStatus)

	switch clusterStatus {
	case status.Unmodified:
		return errors.New("WeGO not installed... exiting")
	case status.Unknown:
		return errors.New("WeGO can not determine cluster status... exiting")
	}

	switch strings.ToUpper(params.AppConfigUrl) {
	case string(ConfigTypeNone):
		return a.addAppWithNoConfigRepo(params)
		// case string(ConfigTypeUserRepo):
		// 	return addAppWithConfigInUserRepo(ctx, deps.GitClient)
		// default:
		// 	return addAppWithConfigInExternalRepo(ctx, deps.GitClient)
	}

	return nil
}

func (a *App) updateParametersIfNecessary(params Params) (Params, error) {
	params.SourceType = string(SourceTypeGit)
	if params.Chart != "" {
		params.SourceType = string(SourceTypeHelm)
		params.DeploymentType = string(DeployTypeHelm)
	}

	if params.AppConfigUrl == string(ConfigTypeUserRepo) && params.SourceType != string(SourceTypeGit) {
		return params, fmt.Errorf("cannot create .wego directory in helm repository:\n" +
			"  you must either use --app-config-url=none or --appconfig-url=<url of external git repo>")
	}

	// Identifying repo url if not set by the user
	if params.Url == "" {
		url, err := a.getGitRemoteUrl(params)
		if err != nil {
			return params, err
		}

		params.Url = url
	}

	sshPrefix := "git@github.com:"
	if strings.HasPrefix(params.Url, sshPrefix) {
		trimmed := strings.TrimPrefix(params.Url, sshPrefix)
		params.Url = "ssh://git@github.com/" + trimmed
	}
	fmt.Printf("using URL: '%s' of origin from git config...\n\n", params.Url)

	if params.Name == "" {
		params.Name = generateAppName(params.Url)
	}

	return params, nil
}

func (a *App) getGitRemoteUrl(params Params) (string, error) {
	repo, err := a.git.Open(params.Dir)
	if err != nil {
		return "", errors.Wrapf(err, "failed to open repository: %s", params.Dir)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return "", errors.Wrapf(err, "failed to find the origin remote in the repository")
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		return "", errors.Errorf("remote config in %s does not have an url", params.Dir)
	}

	return urls[0], nil
}

func (a *App) addAppWithNoConfigRepo(params Params) error {
	if SourceType(params.SourceType) == SourceTypeGit {
		fmt.Println("Generating deploy key...")
		if !params.DryRun {
			if err := a.createAndUploadDeployKey(params); err != nil {
				return errors.Wrap(err, "could not generate deploy key")
			}
		}
	}

	var err error
	var sourceManifest, appManifest, appGoatManifest []byte
	// Source covers entire user repo
	fmt.Println("Generating source manifest...")
	sourceManifest, err = a.generateSource(params)
	if err != nil {
		return errors.Wrap(err, "could not set up GitOps for user repository")
	}

	fmt.Println("Generating app manifest...")
	appManifest, err = generateAppYaml(params)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not create app.yaml for '%s'", params.Name))
	}

	// kustomize or helm referencing single user repo source
	fmt.Println("Generating GitOps automation manifests...")
	appGoatManifest, err = a.generateApplicationGoat(params)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not create GitOps automation for '%s'", params.Name))
	}

	fmt.Println("Applying manifests to the cluster...")
	return a.applyToCluster(params, sourceManifest, appManifest, appGoatManifest)
}

func (a *App) createAndUploadDeployKey(params Params) error {
	deployKey, err := a.flux.CreateSecretGit(params.Name, params.Url, params.Namespace)
	if err != nil {
		return errors.Wrap(err, "could not create git secret")
	}

	owner, err := getOwnerFromUrl(params.Url)
	if err != nil {
		return err
	}

	if err := gitproviders.UploadDeployKey(owner, params.Name, deployKey); err != nil {
		return errors.Wrap(err, "error uploading deploy key")
	}

	return nil
}

func (a *App) generateSource(params Params) ([]byte, error) {
	switch SourceType(params.SourceType) {
	case SourceTypeGit:
		sourceManifest, err := a.flux.CreateSourceGit(params.Name, params.Url, params.Branch, params.Name, params.Namespace)
		if err != nil {
			return nil, errors.Wrap(err, "could not create git source")
		}

		return sourceManifest, nil
	case SourceTypeHelm:
		return a.flux.CreateHelmReleaseHelmRepository(params.Name, params.Url, params.Namespace)
	default:
		return nil, fmt.Errorf("unknown source type: %v", params.SourceType)
	}
}

func (a *App) generateApplicationGoat(params Params) ([]byte, error) {
	switch params.DeploymentType {
	case string(DeployTypeKustomize):
		return a.flux.CreateKustomization(params.Name, params.Path, params.Namespace)
	case string(DeployTypeHelm):
		switch params.SourceType {
		case string(SourceTypeHelm):
			return a.flux.CreateHelmReleaseHelmRepository(params.Name, params.Chart, params.Namespace)
		case string(SourceTypeGit):
			return a.flux.CreateHelmReleaseGitRepository(params.Name, params.Path, params.Namespace)
		default:
			return nil, fmt.Errorf("invalid source type: %v", params.SourceType)
		}
	default:
		return nil, fmt.Errorf("invalid deployment type: %v", params.DeploymentType)
	}
}

func (a *App) applyToCluster(params Params, manifests ...[]byte) error {
	if params.DryRun {
		for _, manifest := range manifests {
			fmt.Printf("%s\n", manifest)
		}
		return nil
	}

	for _, manifest := range manifests {
		if out, err := a.kube.Apply(manifest, params.Namespace); err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not apply manifest: %s", string(out)))
		}
	}

	return nil
}

func generateAppYaml(params Params) ([]byte, error) {
	const appYamlTemplate = `---
apiVersion: wego.weave.works/v1alpha1
kind: Application
metadata:
  name: {{ .AppName }}
spec:
  path: {{ .AppPath }}
  url: {{ .AppURL }}
`
	// Create app.yaml
	t, err := template.New("appYaml").Parse(appYamlTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse app yaml template")
	}

	var populated bytes.Buffer
	err = t.Execute(&populated, struct {
		AppName string
		AppPath string
		AppURL  string
	}{params.Name, params.Path, params.Url})
	if err != nil {
		return nil, errors.Wrap(err, "could not execute populated template")
	}
	return populated.Bytes(), nil
}

func generateAppName(url string) string {
	return strings.TrimSuffix(strings.ReplaceAll(filepath.Base(url), "_", "-"), ".git")
}

func getOwnerFromUrl(url string) (string, error) {
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("could not get owner from url %s", url)
	}
	return parts[len(parts)-2], nil
}
