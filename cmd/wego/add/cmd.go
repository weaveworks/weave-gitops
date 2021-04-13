package add

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/status"
)

type paramSet struct {
	name   string
	url    string
	branch string
}

var params paramSet

var Cmd = &cobra.Command{
	Use:   "add [--name <name>] [--url <url>] [--branch <branch>]",
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

func updateParametersIfNecessary() {
	if params.url == "" {
		url, err := exec.Command("git", "remote", "get-url", "origin").CombinedOutput()
		checkError("Failed to discover URL of remote repository", err)
		fmt.Printf("URL not specified; using URL of 'origin' from git config...")
		params.url = string(url)
	}

	if params.name == "" {
		clusterName, err := status.GetClusterName()
		checkAddError(err)
		params.name = clusterName + "-wego"
	}
}

func generateSourceManifest() []byte {
	sourceManifest, err := fluxops.CallFlux(fmt.Sprintf(`create source git "%s" --url="%s" --branch="%s" --interval=30s --export`,
		params.name, params.url, params.branch))
	checkAddError(err)
	return sourceManifest
}

func generateKustomizeManifest() []byte {
	kustomizeManifest, err := fluxops.CallFlux(
		fmt.Sprintf(`create kustomization %s --path="./kustomize" --prune=true --validation=client --interval=5m --export`, params.name))
	checkAddError(err)
	return kustomizeManifest
}

func bootstrapOrExit() {
	fmt.Printf("The cluster has not had wego installed; install it now? (Y/n)")
	if !proceed() {
		fmt.Fprintf(os.Stderr, "Wego not installed.")
		os.Exit(1)
	}

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
		str = "N\n"
	}
	return strings.Trim(str, "\n")
}

func validAnswer(answer string) bool {
	return strings.EqualFold(answer, "y") || strings.EqualFold(answer, "n")
}

func runCmd(cmd *cobra.Command, args []string) {
	updateParametersIfNecessary()
	fmt.Printf("UPDATED...\n")
	if clusterStatus := status.GetClusterStatus(); clusterStatus == status.Unmodified {
		bootstrapOrExit()
	}
	fmt.Printf("BOOTSTRAPPED...\n")
	time.Sleep(30 * time.Minute)
	applyCmd := exec.Command("kubectl", "apply", "-f", "-")
	stdin, err := applyCmd.StdinPipe()
	checkAddError(err)
	defer stdin.Close()
	io.WriteString(stdin, fmt.Sprintf("%s\n---\n%s", generateSourceManifest(), generateKustomizeManifest()))
	fmt.Printf("FED DATA...\n")
	applyCmd.Stdout = os.Stdout
	applyCmd.Stderr = os.Stderr
	checkAddError(applyCmd.Run())
}
