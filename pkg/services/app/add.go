package app

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
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

type AddParams struct {
	Dir                  string
	Name                 string
	Url                  string
	Path                 string
	Branch               string
	PrivateKey           string
	PrivateKeyPass       string
	DeploymentType       string
	Chart                string
	SourceType           string
	CommitManifests      bool
	AutomationRepo       string
	AutomationRepoPath   string
	AutomationRepoBranch string
	Namespace            string
	DryRun               bool
}

func (a *App) Add(params AddParams) error {
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

	clusterName, err := status.GetClusterName()
	if err != nil {
		return err
	}

	if params.AutomationRepo != "" {
		return a.addAppWithConfigInExternalRepo(params, clusterName)
	}

	if params.CommitManifests {
		return a.addAppWithConfigInAppRepo(params, clusterName)
	}

	return a.addAppWithNoConfigRepo(params, clusterName)

}

func (a *App) updateParametersIfNecessary(params AddParams) (AddParams, error) {
	params.SourceType = string(SourceTypeGit)
	if params.Chart != "" {
		params.SourceType = string(SourceTypeHelm)
		params.DeploymentType = string(DeployTypeHelm)
	}

	// Identifying repo url if not set by the user
	if params.Url == "" {
		url, err := a.getGitRemoteUrl(params)
		if err != nil {
			return params, err
		}

		params.Url = url
	}

	fmt.Printf("using URL: '%s' of origin from git config...\n\n", params.Url)

	if params.Name == "" {
		params.Name = generateResourceName(params.Url)
	}

	return params, nil
}

func (a *App) cloneRepo(url string, branch string) error {
	url = sanitizeRepoUrl(url)

	repoDir, err := ioutil.TempDir("", "user-repo-")
	if err != nil {
		return errors.Wrap(err, "failed creating temp. directory to clone repo")
	}

	_, err = a.git.Clone(context.Background(), repoDir, url, branch)
	if err != nil {
		return errors.Wrapf(err, "failed cloning user repo: %s", url)
	}

	return nil
}

func (a *App) getGitRemoteUrl(params AddParams) (string, error) {
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

	return sanitizeRepoUrl(urls[0]), nil
}

func (a *App) addAppWithNoConfigRepo(params AddParams, clusterName string) error {
	var secretRef string
	var err error
	if SourceType(params.SourceType) == SourceTypeGit {
		fmt.Println("Generating deploy key...")
		if !params.DryRun {
			secretRef, err = a.createAndUploadDeployKey(params.Url, clusterName, params.Namespace)
			if err != nil {
				return errors.Wrap(err, "could not generate deploy key")
			}
		}
	}

	var sourceManifest, appManifest, appGoatManifest []byte
	// Source covers entire user repo
	fmt.Println("Generating source manifest...")
	sourceManifest, err = a.generateSource(params, secretRef)
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

func (a *App) addAppWithConfigInAppRepo(params AddParams, clusterName string) error {
	var secretRef string
	var err error
	if SourceType(params.SourceType) == SourceTypeGit {
		fmt.Println("Generating deploy key...")
		if !params.DryRun {
			secretRef, err = a.createAndUploadDeployKey(params.Url, clusterName, params.Namespace)
			if err != nil {
				return errors.Wrap(err, "could not generate deploy key")
			}
		}
	}

	var sourceManifest, appManifest, appGoatManifest []byte
	// Source covers entire user repo
	fmt.Println("Generating source manifest...")
	sourceManifest, err = a.generateSource(params, secretRef)
	if err != nil {
		return errors.Wrap(err, "could not set up GitOps for user repository")
	}

	fmt.Println("Generating app manifest...")
	appManifest, err = generateAppYaml(params)
	if err != nil {
		return errors.Wrapf(err, "could not create app.yaml for '%s'", params.Name)
	}

	// kustomize or helm referencing single user repo source
	fmt.Println("Generating GitOps automation manifests...")
	appGoatManifest, err = a.generateApplicationGoat(params)
	if err != nil {
		return errors.Wrapf(err, "could not create GitOps automation for '%s'", params.Name)
	}

	fmt.Println("Applying manifests to the cluster...")
	if err := a.applyToCluster(params, sourceManifest, appManifest, appGoatManifest); err != nil {
		return errors.Wrap(err, "could not apply manifests to the cluster")
	}

	// a local directory has not been passed, so we clone the repo passed in the --url
	if params.Dir == "" {
		if err := a.cloneRepo(params.Url, params.Branch); err != nil {
			return errors.Wrap(err, "failed to clone application repo")
		}
	}

	fmt.Println("Writing manifests to disk...")
	if err := a.writeAppYaml(".wego", params.Name, appManifest); err != nil {
		return errors.Wrap(err, "failed writing app.yaml to disk")
	}

	if err := a.writeAppGoats(".wego", params.Name, clusterName, sourceManifest, appGoatManifest); err != nil {
		return errors.Wrap(err, "failed writing app.yaml to disk")
	}

	return a.commitAndPush(params, func(fname string) bool {
		return strings.HasPrefix(fname, ".wego")
	})
}

func (a *App) addAppWithConfigInExternalRepo(params AddParams, clusterName string) error {
	// making sure the url is in good format
	params.AutomationRepo = sanitizeRepoUrl(params.AutomationRepo)

	fmt.Println("Generating deploy key...")
	var secretRef string
	var err error
	if !params.DryRun {
		secretRef, err = a.createAndUploadDeployKey(params.AutomationRepo, clusterName, params.Namespace)
		if err != nil {
			return errors.Wrap(err, "could not generate deploy key")
		}
	}

	var appSourceManifest, appManifest, appGoatManifest []byte
	// Source covers entire user repo
	fmt.Println("Generating source manifest...")
	appSourceManifest, err = a.generateSource(params, secretRef)
	if err != nil {
		return errors.Wrap(err, "could not set up GitOps for user repository")
	}

	fmt.Println("Generating app manifest...")
	appManifest, err = generateAppYaml(params)
	if err != nil {
		return errors.Wrapf(err, "could not create app.yaml for '%s'", params.Name)
	}

	// kustomize or helm referencing single user repo source
	fmt.Println("Generating GitOps automation manifests...")
	appGoatManifest, err = a.generateApplicationGoat(params)
	if err != nil {
		return errors.Wrapf(err, "could not create GitOps automation for '%s'", params.Name)
	}

	targetSourceManifest, err := a.generateTargetSource(params, secretRef)
	if err != nil {
		return errors.Wrap(err, "could not generate target source manifests")
	}

	targetGoatManifest, err := a.generateTargetGoat(params, clusterName)
	if err != nil {
		return errors.Wrap(err, "could not generate target goat manifests")
	}

	fmt.Println("Applying manifests to the cluster...")
	if err := a.applyToCluster(params, appSourceManifest, appManifest, appGoatManifest, targetSourceManifest, targetGoatManifest); err != nil {
		return errors.Wrapf(err, "could not apply manifests to the cluster")
	}

	if err := a.cloneRepo(params.AutomationRepo, params.AutomationRepoBranch); err != nil {
		return errors.Wrap(err, "failed to clone application repo")
	}

	fmt.Println("Writing manifests to disk...")
	if err := a.writeAppYaml(params.AutomationRepoPath, params.Name, appManifest); err != nil {
		return errors.Wrap(err, "failed writing app.yaml to disk")
	}

	if err := a.writeAppGoats(params.AutomationRepoPath, params.Name, clusterName, appSourceManifest, appGoatManifest); err != nil {
		return errors.Wrap(err, "failed writing app.yaml to disk")
	}

	return a.commitAndPush(params, func(fname string) bool {
		return strings.Contains(fname, strings.TrimPrefix(params.AutomationRepoPath, "./"))
	})
}

func (a *App) writeAppYaml(basePath string, name string, manifest []byte) error {
	manifestPath := filepath.Join(basePath, "apps", name, "app.yaml")

	return a.git.Write(manifestPath, manifest)
}

func (a *App) writeAppGoats(basePath string, name string, clusterName string, manifests ...[]byte) error {
	goatPath := filepath.Join(basePath, "targets", clusterName, name, fmt.Sprintf("%s-gitops-runtime.yaml", name))

	goat := bytes.Join(manifests, []byte(""))
	return a.git.Write(goatPath, goat)
}

// NOTE: ready to save the targets automation in phase 2
// func (a *App) writeTargetGoats(basePath string, name string, manifests ...[]byte) error {
// 	goatPath := filepath.Join(basePath, "targets", fmt.Sprintf("%s-gitops-runtime.yaml", name))

// 	goat := bytes.Join(manifests, []byte(""))
// 	return a.git.Write(goatPath, goat)
// }

func (a *App) generateTargetSource(params AddParams, secretRef string) ([]byte, error) {
	repoName := generateResourceName(params.AutomationRepo)

	return a.flux.CreateSourceGit(repoName, params.AutomationRepo, params.AutomationRepoBranch, secretRef, params.Namespace)
}

func (a *App) generateTargetGoat(params AddParams, clusterName string) ([]byte, error) {
	repoName := urlToRepoName(params.AutomationRepo)

	targetPath := filepath.Join(params.AutomationRepoPath, "targets", clusterName)

	return a.flux.CreateKustomization(fmt.Sprintf("weave-gitops-%s", clusterName), repoName, targetPath, params.Namespace)
}

func (a *App) commitAndPush(params AddParams, filters ...func(string) bool) error {
	fmt.Println("Commiting and pushing wego resources for application...")
	if params.DryRun {
		return nil
	}

	_, err := a.git.Commit(git.Commit{
		Author:  git.Author{Name: "Weave Gitops", Email: "weave-gitops@weave.works"},
		Message: "Add App manifests",
	}, filters...)
	if err != nil && err != git.ErrNoStagedFiles {
		return fmt.Errorf("failed to commit sync manifests: %w", err)
	}

	if err == nil {
		fmt.Println("Pushing app manifests to repository...")
		if err = a.git.Push(context.Background()); err != nil {
			return fmt.Errorf("failed to push manifests: %w", err)
		}
	} else {
		fmt.Println("App manifests are up to date")
	}

	return nil
}

func (a *App) createAndUploadDeployKey(repoUrl string, clusterName string, namespace string) (string, error) {
	repoUrl = sanitizeRepoUrl(repoUrl)

	secretRef := fmt.Sprintf("weave-gitops-%s", clusterName)

	deployKey, err := a.flux.CreateSecretGit(secretRef, repoUrl, namespace)
	if err != nil {
		return "", errors.Wrap(err, "could not create git secret")
	}

	owner, err := getOwnerFromUrl(repoUrl)
	if err != nil {
		return "", err
	}

	repoName := urlToRepoName(repoUrl)
	if err := gitproviders.UploadDeployKey(owner, repoName, deployKey); err != nil {
		return "", errors.Wrap(err, "error uploading deploy key")
	}

	return secretRef, nil
}

func (a *App) generateSource(params AddParams, secretRef string) ([]byte, error) {
	switch SourceType(params.SourceType) {
	case SourceTypeGit:
		sourceManifest, err := a.flux.CreateSourceGit(params.Name, params.Url, params.Branch, secretRef, params.Namespace)
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

func (a *App) generateApplicationGoat(params AddParams) ([]byte, error) {
	switch params.DeploymentType {
	case string(DeployTypeKustomize):
		return a.flux.CreateKustomization(params.Name, params.Name, params.Path, params.Namespace)
	case string(DeployTypeHelm):
		switch params.SourceType {
		case string(SourceTypeHelm):
			return a.flux.CreateHelmReleaseHelmRepository(params.Name, params.Chart, params.Namespace)
		case string(SourceTypeGit):
			return a.flux.CreateHelmReleaseGitRepository(params.Name, params.Name, params.Path, params.Namespace)
		default:
			return nil, fmt.Errorf("invalid source type: %v", params.SourceType)
		}
	default:
		return nil, fmt.Errorf("invalid deployment type: %v", params.DeploymentType)
	}
}

func (a *App) applyToCluster(params AddParams, manifests ...[]byte) error {
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

func generateAppYaml(params AddParams) ([]byte, error) {
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

func generateResourceName(url string) string {
	return strings.ReplaceAll(urlToRepoName(url), "_", "-")
}

func getOwnerFromUrl(url string) (string, error) {
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("could not get owner from url %s", url)
	}
	return parts[len(parts)-2], nil
}

func urlToRepoName(url string) string {
	return strings.TrimSuffix(filepath.Base(url), ".git")
}

func sanitizeRepoUrl(url string) string {
	trimmed := ""

	sshPrefix := "git@github.com:"
	if strings.HasPrefix(url, sshPrefix) {
		trimmed = strings.TrimPrefix(url, sshPrefix)
	}

	httpsPrefix := "https://github.com/"
	if strings.HasPrefix(url, httpsPrefix) {
		trimmed = strings.TrimPrefix(url, httpsPrefix)
	}

	if trimmed != "" {
		return "ssh://git@github.com/" + trimmed
	}

	return url
}
