package add

// Provides support for adding a repository of manifests to a wego cluster. If the cluster does not have
// wego installed, the user will be prompted to install wego and then the repository will be added.

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	//	"time"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/status"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

type paramSet struct {
	dir    string
	name   string
	url    string
	branch string
}

var (
	params    paramSet
	repoOwner string
)

var Cmd = &cobra.Command{
	Use:   "add [--name <name>] [--url <url>] [--branch <branch>] <repository directory>",
	Short: "Add a workload repository to a wego cluster",
	Long: strings.TrimSpace(dedent.Dedent(`
        Associates an additional git repository with a wego cluster so that its contents may be managed via GitOps
    `)),
	Example: "wego add",
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
	Cmd.Flags().StringVar(&params.branch, "branch", "main", "Branch to watch within git repository")
}

func getClusterRepoName() string {
	clusterName, err := status.GetClusterName()
	checkAddError(err)
	return clusterName + "-wego"
}

func updateParametersIfNecessary() {
	d, _ := os.Getwd()
	fmt.Printf("UPIN -- URL: %s, NAME: %s, DIR: %s\n", params.url, params.name, d)

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
	fmt.Printf("UPIN2 -- URL: %s, NAME: %s\n", params.url, params.name)

	sshPrefix := "git@github.com:"
	if strings.HasPrefix(params.url, sshPrefix) {
		isPrivate, err := fluxops.IsPrivate(getOwner(), params.name)
		checkAddError(err)
		trimmed := strings.TrimPrefix(params.url, sshPrefix)
		if isPrivate {
			params.url = "ssh://git@github.com/" + trimmed
		} else {
			params.url = "https://github.com/" + trimmed
		}
	}

	fmt.Printf("using URL: '%s' of origin from git config...\n\n", params.url)

}

func generateSourceManifest() []byte {
	sourceManifest, err := fluxops.CallFlux(fmt.Sprintf(`create source git "%s" --url="%s" --branch="%s" --interval=30s --export`,
		params.name, params.url, params.branch))
	checkAddError(err)
	return sourceManifest
}

func generateKustomizeManifest() []byte {
	kustomizeManifest, err := fluxops.CallFlux(
		fmt.Sprintf(`create kustomization "%s" --path="./" --source="%s" --prune=true --validation=client --interval=5m --export`, params.name, params.name))
	checkAddError(err)
	return kustomizeManifest
}

// func bootstrapOrExit() {
//  if !askUser("The cluster does not have wego installed; install it now? (Y/n)") {
//      fmt.Fprintf(os.Stderr, "Wego not installed.")
//      os.Exit(1)
//  }
//  repoName, err := fluxops.GetRepoName()
//  checkAddError(err)
//  fluxops.Bootstrap(getOwner(), repoName)
// }

func askUser(question string) bool {
	fmt.Printf("%s ", question)
	return proceed()
}

func proceed() bool {
	answer := getAnswer()
	for !validAnswer(answer) {
		fmt.Println("Invalid answer, please choose 'Y' or 'n'")
		answer = getAnswer()
	}
	return strings.EqualFold(answer, "y")
}

func getAnswer() string {
	reader := bufio.NewReader(os.Stdin)
	str, err := reader.ReadString('\n')
	checkAddError(err)
	if str == "\n" {
		str = "Y\n"
	}
	return strings.Trim(str, "\n")
}

func validAnswer(answer string) bool {
	return strings.EqualFold(answer, "y") || strings.EqualFold(answer, "n")
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
	fmt.Printf("FILES: %#v\n", files)
	//	time.Sleep(60 * time.Minute)
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
	fmt.Printf("Updating parameters from environment... done\n\n")
	updateParametersIfNecessary()
	fmt.Printf("Checking cluster status... ")
	clusterStatus := status.GetClusterStatus()
	fmt.Printf("%s\n\n", clusterStatus)

	if clusterStatus == status.Unmodified {
		fmt.Printf("WeGO not installed... exiting\n")
		os.Exit(1)
	}

	source := generateSourceManifest()
	kust := generateKustomizeManifest()

	fluxRepoName, err := fluxops.GetRepoName()
	checkAddError(err)

	reposDir := filepath.Join(os.Getenv("HOME"), ".wego", "repositories")
	checkAddError(os.MkdirAll(reposDir, 0755))
	fluxRepo := filepath.Join(reposDir, fluxRepoName)
	owner := getOwner()

	if err := utils.CallCommandForEffect(fmt.Sprintf("git ls-remote %s/%s.git", owner, fluxRepoName)); err != nil {
		fmt.Printf("repo does not exist\n")
		checkAddError(os.Chdir(reposDir))
		checkAddError(os.Mkdir(fluxRepoName, 0755))
		checkAddError(os.Chdir(fluxRepo))
		checkAddError(utils.CallCommandForEffect("git init"))
		checkAddError(ioutil.WriteFile("README.md", []byte("# Repository containing references to applications"), 0644))
		checkAddError(utils.CallCommandForEffect("git add README.md && git commit -m'Initial commit'"))
		checkAddError(utils.CallCommandForEffect(fmt.Sprintf("hub create %s/%s", owner, fluxRepoName)))
		checkAddError(utils.CallCommandForEffect("git push -u origin main"))
	}

	// res, _ := utils.CallCommand(fmt.Sprintf("git ls-remote %s/%s.git", owner, fluxRepoName))
	// fmt.Printf("RES: %s\n", res)

	// if _, err := os.Stat(fluxRepo); os.IsNotExist(err) {
	//  err = utils.CallCommandForEffect(fmt.Sprintf("git clone https://github.com/%s/%s.git", owner, fluxRepoName))
	//  checkAddError(err)
	// }

	sourceName := filepath.Join(fluxRepo, fluxRepoName+"-source-"+params.name+".yaml")
	kustName := filepath.Join(fluxRepo, fluxRepoName+"-kustomize-"+params.name+".yaml")
	ioutil.WriteFile(sourceName, source, 0644)
	ioutil.WriteFile(kustName, kust, 0644)
	commitAndPush(sourceName, kustName)

	fmt.Printf("Successfully added repository: %s.\n", params.name)
}
