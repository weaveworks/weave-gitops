package add

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
)

type paramSet struct {
	name   string
	url    string
	branch string
}

var params paramSet
var fluxHandler = defaultFluxHandler

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

func updateURLIfNecessary() {
	if params.url == "" {
		url, err := exec.Command("git", "remote", "get-url", "origin").CombinedOutput()
		checkError("Failed to discover URL of remote repository", err)
		fmt.Printf("URL not specified; using URL of 'origin' from git config...")
		params.url = string(url)
	}
}

func callFlux(arglist string) ([]byte, error) {
	return fluxHandler(arglist)
}

func defaultFluxHandler(arglist string) ([]byte, error) {
	homedir := os.Getenv("HOME")
	return callCommand(fmt.Sprintf("%s/.wego/bin/flux %s", homedir, arglist))
}

func callCommand(cmdstr string) ([]byte, error) {
	cmd := exec.Command(fmt.Sprintf("sh -c '%s'", escape(cmdstr)))
	return cmd.CombinedOutput()
}

func escape(cmd string) string {
	return "'" + strings.ReplaceAll(cmd, "'", "'\"'\"'") + "'"
}

func generateSourceManifest() []byte {
	sourceManifest, err := callFlux(fmt.Sprintf(`create source git --name="%s" --url="%s" --branch="%s" --interval=30s --export`,
		params.name, params.url, params.branch))
	checkAddError(err)
	return sourceManifest
}

func generateKustomizeManifest() []byte {
	kustomizeManifest, err := callFlux(
		fmt.Sprintf(`create kustomization %s --path="./kustomize" --prune=true --validation=client --interval=5m --export`, params.name))
	checkAddError(err)
	return kustomizeManifest
}

func runCmd(cmd *cobra.Command, args []string) {
	updateURLIfNecessary()
	applyCmd := exec.Command("kubectl apply -f -")
	stdin, err := applyCmd.StdinPipe()
	checkAddError(err)
	defer stdin.Close()
	io.WriteString(stdin, fmt.Sprintf("%s\n---\n%s", generateSourceManifest(), generateKustomizeManifest()))
	checkAddError(applyCmd.Run())
}
