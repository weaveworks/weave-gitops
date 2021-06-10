package app

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/status"
	"github.com/weaveworks/weave-gitops/pkg/utils"
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
	GitClient git.Git
}

type AppService interface {
	Add()
}

type App struct {
	gitClient git.Git
}

func New(deps *Dependencies) *App {
	return &App{
		gitClient: deps.GitClient,
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
	repo, err := a.gitClient.Open(params.Dir)
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
	// Source covers entire user repo
	userRepoSource, err := a.generateSource(params)
	if err != nil {
		return errors.Wrap(err, "could not set up GitOps for user repository")
	}
	appYaml, err := generateAppYaml(params)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not create app.yaml for '%s'", params.Name))
	}
	// kustomize or helm referencing single user repo source
	applicationGOAT, err := a.generateApplicationGoat(params)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not create GitOps automation for '%s'", params.Name))
	}
	return a.applyToCluster(params, userRepoSource, appYaml, applicationGOAT)
}

func (a *App) generateSource(params Params) ([]byte, error) {
	secretName := params.Name

	switch SourceType(params.SourceType) {
	case SourceTypeGit:
		cmd := fmt.Sprintf(`create secret git "%s" \
            --url="%s" \
            --namespace="%s"`,
			secretName,
			params.Url,
			params.Namespace)
		if params.DryRun {
			fmt.Printf(cmd + "\n")
		} else {
			// TODO create a function for this in fluxops pkg
			output, err := fluxops.WithFluxHandler(fluxops.QuietFluxHandler{}, func() ([]byte, error) {
				return fluxops.CallFlux(cmd)
			})
			if err != nil {
				return nil, errors.Wrap(err, "could not create git secret")
			}
			owner, err := getOwnerFromUrl(params.Url)
			if err != nil {
				return nil, err
			}
			deployKeyBody := bytes.TrimPrefix(output, []byte("✚ deploy key: "))
			deployKeyLines := bytes.Split(deployKeyBody, []byte("\n"))
			if len(deployKeyBody) == 0 {
				return nil, fmt.Errorf("no deploy key found [%s]", string(output))
			}
			if err := gitproviders.UploadDeployKey(owner, params.Name, deployKeyLines[0]); err != nil {
				return nil, errors.Wrap(err, "error uploading deploy key")
			}
		}
		cmd = fmt.Sprintf(`create source git "%s" \
            --url="%s" \
            --branch="%s" \
            --secret-ref="%s" \
            --interval=30s \
            --export \
            --namespace="%s"`,
			params.Name,
			params.Url,
			params.Branch,
			secretName,
			params.Namespace)
		sourceManifest, err := fluxops.CallFlux(cmd)
		if err != nil {
			return nil, errors.Wrap(err, "could not create git source")
		}
		return sourceManifest, nil
	case SourceTypeHelm:
		return a.generateSourceHelmManifest(params)
	default:
		return nil, fmt.Errorf("unknown source type: %v", params.SourceType)
	}
}

func (a *App) generateSourceHelmManifest(params Params) ([]byte, error) {
	cmd := fmt.Sprintf(`create source helm %s \
            --url="%s" \
            --interval=30s \
            --export \
            --namespace=%s `,
		params.Name,
		params.Url,
		params.Namespace)

	sourceManifest, err := fluxops.CallFlux(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "could not create git source")
	}
	return sourceManifest, nil
}

func generateAppYaml(params Params) ([]byte, error) {
	const appYamlTemplate = `apiVersion: wego.weave.works/v1alpha1
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

func (a *App) generateApplicationGoat(params Params) ([]byte, error) {
	switch params.DeploymentType {
	case string(DeployTypeKustomize):
		return a.generateKustomizeManifest(params)
	case string(DeployTypeHelm):
		switch params.SourceType {
		case string(SourceTypeHelm):
			return a.generateHelmManifestHelmRepository(params)
		case string(SourceTypeGit):
			return a.generateHelmManifestGitRepository(params)
		default:
			return nil, fmt.Errorf("Invalid source type: %v", params.SourceType)
		}
	default:
		return nil, fmt.Errorf("Invalid deployment type: %v", params.DeploymentType)
	}
}

func (a *App) generateKustomizeManifest(params Params) ([]byte, error) {
	cmd := fmt.Sprintf(`create kustomization "%s" \
                --path="%s" \
                --source="%s" \
                --prune=true \
                --validation=client \
                --interval=1m \
                --export \
                --namespace=%s`,
		params.Name,
		params.Path,
		params.Name,
		params.Namespace)
	kustomizeManifest, err := fluxops.CallFlux(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "could not create kustomization manifest")
	}

	return bytes.ReplaceAll(kustomizeManifest, []byte("path: ./wego"), []byte("path: .wego")), nil
}

func (a *App) generateHelmManifestHelmRepository(params Params) ([]byte, error) {
	cmd := fmt.Sprintf(`create helmrelease %s \
            --source="HelmRepository/%s" \
            --chart="%s" \
            --interval=5m \
            --export \
            --namespace=%s`,
		params.Name,
		params.Name,
		params.Chart,
		params.Namespace,
	)

	return fluxops.CallFlux(cmd)
}

func (a *App) generateHelmManifestGitRepository(params Params) ([]byte, error) {
	cmd := fmt.Sprintf(`create helmrelease %s \
            --source="GitRepository/%s" \
            --chart="%s" \
            --interval=1m \
            --export \
            --namespace=%s`,
		params.Name,
		params.Name,
		params.Name,
		params.Namespace,
	)
	helmManifest, err := fluxops.CallFlux(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "could not create helm manifest")
	}

	return bytes.ReplaceAll(helmManifest, []byte("path: ./wego"), []byte("path: .wego")), nil
}

func (a *App) applyToCluster(params Params, manifests ...[]byte) error {
	if params.DryRun {
		fmt.Printf("Applying:\n\n")
		for _, manifest := range manifests {
			fmt.Printf("%s\n", manifest)
		}
		return nil
	}

	kubectlApply := fmt.Sprintf("kubectl apply --namespace=%s -f -", params.Namespace)

	for _, manifest := range manifests {
		if err := utils.CallCommandForEffectWithInputPipe(kubectlApply, string(manifest)); err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not apply manifest: %s", manifest))
		}
	}

	return nil
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
