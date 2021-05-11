package cmdimpl

// Implementation of the 'wego add' command

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	cgitprovider "github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/status"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

const appYamlTemplate = `apiVersion: wego.weave.works/v1alpha1
kind: Application
metadata:
  name: {{ .AppName }}
spec:
  path: {{ .AppPath }}
  url: {{ .AppURL }}
`

// Will move into filesystem when we store wego infrastructure in git
const appCRD = `apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: apps.wego.weave.works
spec:
  group: wego.weave.works
  scope: Cluster
  names:
    kind: Application
    listKind: ApplicationList
    plural: apps
    singular: app
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      required: ["spec"]
      properties:
        spec:
          required: ["url", "path"]
          properties:
            url:
              type: "string"
              minimum: 1
              maximum: 1
            path:
              type: "string"
              minimum: 1
              maximum: 1
  version: v1alpha1
  versions:
    - name: v1alpha1
      served: true
      storage: true
`

const (
	DeployTypeKustomize = "kustomize"
	DeployTypeHelm      = "helm"
)

type AddParamSet struct {
	Dir            string
	Name           string
	Url            string
	Path           string
	Branch         string
	PrivateKey     string
	DeploymentType string
}

var (
	params    AddParamSet
	repoOwner string
)

// checkError will print a message to stderr and exit
func checkError(msg string, err interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		os.Exit(1)
	}
}

func checkAddError(err interface{}) {
	checkError("Failed to add workload repository", err)
}

func getClusterRepoName() string {
	clusterName, err := status.GetClusterName()
	checkAddError(err)
	return clusterName + "-wego"
}

func updateParametersIfNecessary() {
	if params.Name == "" {
		repoPath, err := filepath.Abs(params.Dir)
		checkAddError(err)
		repoName := strings.ReplaceAll(filepath.Base(repoPath), "_", "-")
		params.Name = getClusterRepoName() + "-" + repoName
	}

	if params.Url == "" {
		urlout, err := exec.Command("git", "remote", "get-url", "origin").CombinedOutput()
		checkError("Failed to discover URL of remote repository", err)
		url := strings.TrimRight(string(urlout), "\n")
		fmt.Printf("URL not specified; ")
		params.Url = url
	}

	sshPrefix := "git@github.com:"
	if strings.HasPrefix(params.Url, sshPrefix) {
		trimmed := strings.TrimPrefix(params.Url, sshPrefix)
		params.Url = "ssh://git@github.com/" + trimmed
	}

	fmt.Printf("using URL: '%s' of origin from git config...\n\n", params.Url)

}

func generateWegoSourceManifest() []byte {
	fluxRepoName, err := fluxops.GetRepoName()
	checkAddError(err)
	_, err = fluxops.CallFlux(fmt.Sprintf(`create secret git "wego" --url="ssh://git@github.com/%s/%s" --private-key-file="%s"`, getOwner(), fluxRepoName, params.PrivateKey))
	checkAddError(err)
	sourceManifest, err := fluxops.CallFlux(fmt.Sprintf(`create source git "wego" --url="ssh://git@github.com/%s/%s" --branch="%s" --secret-ref="wego" --interval=30s --export`,
		getOwner(), fluxRepoName, params.Branch))
	checkAddError(err)
	return sourceManifest
}

func generateWegoKustomizeManifest() []byte {
	kustomizeManifest, err := fluxops.CallFlux(`create kustomization "wego" --path="./" --source="wego" --prune=true --validation=client --interval=5m --export`)
	checkAddError(err)
	return kustomizeManifest
}

func generateSourceManifest() []byte {
	secretName := params.Name
	_, err := fluxops.CallFlux(fmt.Sprintf(`create secret git "%s" --url="%s" --private-key-file="%s"`, secretName, params.Url, params.PrivateKey))
	checkAddError(err)
	sourceManifest, err := fluxops.CallFlux(fmt.Sprintf(`create source git "%s" --url="%s" --branch="%s" --secret-ref="%s" --interval=30s --export`,
		params.Name, params.Url, params.Branch, secretName))
	checkAddError(err)
	return sourceManifest
}

func generateKustomizeManifest() []byte {
	kustomizeManifest, err := fluxops.CallFlux(
		fmt.Sprintf(`create kustomization "%s" --path="./" --source="%s" --prune=true --validation=client --interval=5m --export`, params.Name, params.Name))
	checkAddError(err)
	return kustomizeManifest
}

func generateHelmManifest() []byte {
	helmManifest, err := fluxops.CallFlux(
		fmt.Sprintf(`create helmrelease %s --source="%s" --chart="%s" --interval=5m --export`, params.Name, params.Name, params.Path))

	checkAddError(err)
	return helmManifest
}

func getOwner() string {
	if repoOwner != "" {
		return repoOwner
	}
	owner, err := fluxops.GetOwnerFromEnv()
	if err != nil || owner == "" {
		repoOwner = getOwnerInteractively()
		return repoOwner
	}
	repoOwner = owner
	return owner
}

func getOwnerInteractively() string {
	fmt.Printf("Who is the owner of the repository? ")
	reader := bufio.NewReader(os.Stdin)
	str, err := reader.ReadString('\n')
	checkAddError(err)

	if str == "\n" {
		return getOwnerInteractively()
	}

	return strings.Trim(str, "\n")
}

func commitAndPush(files ...string) {
	_, err := utils.CallCommand(
		fmt.Sprintf("git pull --rebase && git add %s && git commit -m'Save %s' && git push", strings.Join(files, " "), strings.Join(files, ", ")))
	checkAddError(err)
}

// Add provides the implementation for the wego add command
func Add(args []string, allParams AddParamSet) {
	if len(args) < 1 {
		fmt.Printf("Location of application not specified.\n")
		os.Exit(1)
	}
	params = allParams
	params.Dir = args[0]
	fmt.Printf("Updating parameters from environment... ")
	updateParametersIfNecessary()
	fmt.Printf("done\n\n")
	fmt.Printf("Checking cluster status... ")
	clusterStatus := status.GetClusterStatus()
	fmt.Printf("%s\n\n", clusterStatus)

	if clusterStatus == status.Unmodified {
		fmt.Printf("WeGO not installed... exiting\n")
		os.Exit(1)
	}

	// Set up wego repository if required
	fluxRepoName, err := fluxops.GetRepoName()
	checkAddError(err)

	reposDir := filepath.Join(os.Getenv("HOME"), ".wego", "repositories")
	fluxRepo := filepath.Join(reposDir, fluxRepoName)
	appSubdir := filepath.Join(fluxRepo, "apps", params.Name)
	checkAddError(os.MkdirAll(appSubdir, 0755))

	owner := getOwner()
	checkAddError(os.Chdir(fluxRepo))

	if err := utils.CallCommandForEffect(fmt.Sprintf("git ls-remote ssh://git@github.com/%s/%s.git", owner, fluxRepoName)); err != nil {
		fmt.Printf("repo does not exist\n")
		checkAddError(utils.CallCommandForEffectWithDebug("git init"))

		c, err := cgitprovider.GithubProvider()
		checkAddError(err)

		orgRef := cgitprovider.NewOrgRepositoryRef(cgitprovider.GITHUB_DOMAIN, owner, fluxRepoName)

		repoInfo := cgitprovider.NewRepositoryInfo("wego repo", gitprovider.RepositoryVisibilityPrivate)

		repoCreateOpts := &gitprovider.RepositoryCreateOptions{
			AutoInit:        gitprovider.BoolVar(true),
			LicenseTemplate: gitprovider.LicenseTemplateVar(gitprovider.LicenseTemplateApache2),
		}

		err = cgitprovider.CreateOrgRepository(c, orgRef, repoInfo, repoCreateOpts)
		checkAddError(err)

		checkAddError(utils.CallCommandForEffectWithDebug(
			fmt.Sprintf("git remote add origin %s && git pull --rebase origin main && git push --set-upstream origin main", orgRef.String())))
	} else {
		checkAddError(utils.CallCommandForEffectWithDebug("git branch --set-upstream-to=origin/main main"))
	}

	// Install Source and Kustomize controllers, and CRD for application (may already be present)
	wegoSource := generateWegoSourceManifest()
	wegoKust := generateWegoKustomizeManifest()
	checkAddError(utils.CallCommandForEffectWithInputPipe("kubectl apply -f -", appCRD))
	checkAddError(utils.CallCommandForEffectWithInputPipe("kubectl apply -f -", string(wegoSource)))
	checkAddError(utils.CallCommandForEffectWithInputPipe("kubectl apply -f -", string(wegoKust)))

	// Create app.yaml
	t, err := template.New("appYaml").Parse(appYamlTemplate)
	checkAddError(err)

	var populated bytes.Buffer
	err = t.Execute(&populated, struct {
		AppName string
		AppPath string
		AppURL  string
	}{params.Name, params.Path, params.Url})
	checkAddError(err)

	// Create controllers for new repo being added
	source := generateSourceManifest()

	var appManifests []byte
	if params.DeploymentType == DeployTypeHelm {
		appManifests = generateHelmManifest()
	} else {
		appManifests = generateKustomizeManifest()
	}

	sourceName := filepath.Join(appSubdir, "source-"+params.Name+".yaml")
	manifestsName := filepath.Join(appSubdir, fmt.Sprintf("%s-%s.yaml", params.DeploymentType, params.Name))
	appYamlName := filepath.Join(appSubdir, "app.yaml")

	checkAddError(ioutil.WriteFile(sourceName, source, 0644))
	checkAddError(ioutil.WriteFile(manifestsName, appManifests, 0644))
	checkAddError(ioutil.WriteFile(appYamlName, populated.Bytes(), 0644))

	commitAndPush(sourceName, manifestsName, appYamlName)

	fmt.Printf("Successfully added repository: %s.\n", params.Name)
}
