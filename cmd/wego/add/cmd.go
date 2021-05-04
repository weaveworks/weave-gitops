package add

// Provides support for adding a repository of manifests to a wego cluster. If the cluster does not have
// wego installed, the user will be prompted to install wego and then the repository will be added.

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

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
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

type paramSet struct {
	dir        string
	name       string
	url        string
	path       string
	branch     string
	privateKey string
}

var (
	params    paramSet
	repoOwner string
)

var Cmd = &cobra.Command{
	Use:   "add [--name <name>] [--url <url>] [--branch <branch>] [--path <path within repository>] [--private-key <keyfile>] <repository directory>",
	Short: "Add a workload repository to a wego cluster",
	Long: strings.TrimSpace(dedent.Dedent(`
        Associates an additional git repository with a wego cluster so that its contents may be managed via GitOps
    `)),
	Example: "wego add .",
	Run:     runCmd,
}

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

func init() {
	Cmd.Flags().StringVar(&params.name, "name", "", "Name of remote git repository")
	Cmd.Flags().StringVar(&params.url, "url", "", "URL of remote git repository")
	Cmd.Flags().StringVar(&params.path, "path", "./", "Path of files within git repository")
	Cmd.Flags().StringVar(&params.branch, "branch", "main", "Branch to watch within git repository")
	Cmd.Flags().StringVar(&params.privateKey, "private-key", filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"), "Private key that provides access to git repository")
}

func getClusterRepoName() string {
	clusterName, err := status.GetClusterName()
	checkAddError(err)
	return clusterName + "-wego"
}

func updateParametersIfNecessary() {
	if params.name == "" {
		repoPath, err := filepath.Abs(params.dir)
		checkAddError(err)
		repoName := strings.ReplaceAll(filepath.Base(repoPath), "_", "-")
		params.name = getClusterRepoName() + "-" + repoName
	}

	if params.url == "" {
		urlout, err := exec.Command("git", "remote", "get-url", "origin").CombinedOutput()
		checkError("Failed to discover URL of remote repository", err)
		url := strings.TrimRight(string(urlout), "\n")
		fmt.Printf("URL not specified; ")
		params.url = url
	}

	sshPrefix := "git@github.com:"
	if strings.HasPrefix(params.url, sshPrefix) {
		trimmed := strings.TrimPrefix(params.url, sshPrefix)
		params.url = "ssh://git@github.com/" + trimmed
	}

	fmt.Printf("using URL: '%s' of origin from git config...\n\n", params.url)

}

func generateWegoSourceManifest() []byte {
	fluxRepoName, err := fluxops.GetRepoName()
	checkAddError(err)
	_, err = fluxops.CallFlux(fmt.Sprintf(`create secret git "wego" --url="ssh://git@github.com/%s/%s" --private-key-file="%s"`, getOwner(), fluxRepoName, params.privateKey))
	checkAddError(err)
	sourceManifest, err := fluxops.CallFlux(fmt.Sprintf(`create source git "wego" --url="ssh://git@github.com/%s/%s" --branch="%s" --secret-ref="wego" --interval=30s --export`,
		getOwner(), fluxRepoName, params.branch))
	checkAddError(err)
	return sourceManifest
}

func generateWegoKustomizeManifest() []byte {
	kustomizeManifest, err := fluxops.CallFlux(
		fmt.Sprintf(`create kustomization "wego" --path="./" --source="wego" --prune=true --validation=client --interval=5m --export`))
	checkAddError(err)
	return kustomizeManifest
}

func generateSourceManifest() []byte {
	fluxRepoName, err := fluxops.GetRepoName()
	checkAddError(err)
	secretName := fluxRepoName + "-" + params.name
	_, err = fluxops.CallFlux(fmt.Sprintf(`create secret git "%s" --url="%s" --private-key-file="%s"`, secretName, params.url, params.privateKey))
	checkAddError(err)
	sourceManifest, err := fluxops.CallFlux(fmt.Sprintf(`create source git "%s" --url="%s" --branch="%s" --secret-ref="%s" --interval=30s --export`,
		params.name, params.url, params.branch, secretName))
	checkAddError(err)
	return sourceManifest
}

func generateKustomizeManifest() []byte {
	kustomizeManifest, err := fluxops.CallFlux(
		fmt.Sprintf(`create kustomization "%s" --path="./" --source="%s" --prune=true --validation=client --interval=5m --export`, params.name, params.name))
	checkAddError(err)
	return kustomizeManifest
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

func runCmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Printf("Location of application not specified.\n")
		os.Exit(1)
	}
	params.dir = args[0]
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
	appSubdir := filepath.Join(fluxRepo, params.name)
	checkAddError(os.MkdirAll(appSubdir, 0755))

	owner := getOwner()
	checkAddError(os.Chdir(fluxRepo))

	if err := utils.CallCommandForEffect(fmt.Sprintf("git ls-remote %s/%s.git", owner, fluxRepoName)); err != nil {
		fmt.Printf("repo does not exist\n")
		checkAddError(utils.CallCommandForEffectWithDebug("git init"))
		checkAddError(ioutil.WriteFile("README.md", []byte("# Repository containing references to applications"), 0644))
		checkAddError(utils.CallCommandForEffectWithDebug("git add README.md && git commit -m'Initial commit'"))
		checkAddError(utils.CallCommandForEffectWithDebug(fmt.Sprintf("hub create %s/%s", owner, fluxRepoName)))
		checkAddError(utils.CallCommandForEffectWithDebug("git push -u origin main"))
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
	}{params.name, params.path, params.url})
	checkAddError(err)

	checkAddError(os.MkdirAll(appSubdir, 0755))

	// Create controllers for new repo being added
	source := generateSourceManifest()
	kust := generateKustomizeManifest()
	sourceName := filepath.Join(appSubdir, "source-"+params.name+".yaml")
	kustName := filepath.Join(appSubdir, "kustomize-"+params.name+".yaml")
	appYamlName := filepath.Join(appSubdir, "app.yaml")

	ioutil.WriteFile(sourceName, source, 0644)
	ioutil.WriteFile(kustName, kust, 0644)
	ioutil.WriteFile(appYamlName, populated.Bytes(), 0644)

	commitAndPush(sourceName, kustName, appYamlName)

	fmt.Printf("Successfully added repository: %s.\n", params.name)
}
