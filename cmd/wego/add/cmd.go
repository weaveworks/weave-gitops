package add

// Provides support for adding a repository of manifests to a wego cluster. If the cluster does not have
// wego installed, the user will be prompted to install wego and then the repository will be added.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/shims"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

var params app.AddParams

var Cmd = &cobra.Command{
	Use:   "add [--name <name>] [--url <url>] [--branch <branch>] [--path <path within repository>] [--private-key <keyfile>] <repository directory>",
	Short: "Add a workload repository to a wego cluster",
	Long: strings.TrimSpace(dedent.Dedent(`
        Associates an additional application in a git repository with a wego cluster so that its contents may be managed via GitOps
    `)),
	Example: "wego add .",
	Run:     runCmd,
}

func init() {
	Cmd.Flags().StringVar(&params.Name, "name", "", "Name of remote git repository")
	Cmd.Flags().StringVar(&params.Url, "url", "", "URL of remote repository")
	Cmd.Flags().StringVar(&params.Path, "path", "./", "Path of files within git repository")
	Cmd.Flags().StringVar(&params.Branch, "branch", "main", "Branch to watch within git repository")
	Cmd.Flags().StringVar(&params.DeploymentType, "deployment-type", "kustomize", "deployment type [kustomize, helm]")
	Cmd.Flags().StringVar(&params.Chart, "chart", "", "Specify chart for helm source")
	Cmd.Flags().StringVar(&params.PrivateKey, "private-key", "", "Private key to access git repository over ssh")
	Cmd.Flags().BoolVar(&params.CommitManifests, "commit-manifests", true, "commits gitops automation manifests into the application repository. When using using '--automation-repo' this flag is ignored and it always commits.")
	Cmd.Flags().BoolVar(&params.DryRun, "dry-run", false, "If set, 'wego add' will not make any changes to the system; it will just display the actions that would have been taken")
	Cmd.Flags().StringVar(&params.AutomationRepo, "automation-repo", "", "Repository that will hold the Gitops Automation manifests")
	Cmd.Flags().StringVar(&params.AutomationRepoBranch, "automation-repo-branch", "main", "Repository branch that will hold the Gitops Automation manifests")
	Cmd.Flags().StringVar(&params.AutomationRepoPath, "automation-repo-path", "./wego", "Repository path that will hold the Gitops Automation manifests")
}

func runCmd(cmd *cobra.Command, args []string) {
	params.Namespace, _ = cmd.Parent().Flags().GetString("namespace")

	if strings.HasPrefix(params.PrivateKey, "~/") {
		dir := getHomeDir()
		params.PrivateKey = filepath.Join(dir, params.PrivateKey[2:])
	} else if params.PrivateKey == "" {
		params.PrivateKey = findPrivateKeyFile()
	}

	authMethod, err := ssh.NewPublicKeysFromFile("git", params.PrivateKey, params.PrivateKeyPass)
	if err != nil {
		fmt.Fprintf(shims.Stderr(), "failed reading ssh keys: %s\n", err)
		shims.Exit(1)
	}

	if params.Url == "" {
		if len(args) == 0 {
			fmt.Fprint(shims.Stderr(), "no app --url or app location specified")
			shims.Exit(1)
		} else {
			params.Dir = args[0]
		}
	}

	gitClient := git.New(authMethod)
	fluxClient := flux.New()
	kubeClient := kube.New()

	deps := &app.Dependencies{
		Git:  gitClient,
		Flux: fluxClient,
		Kube: kubeClient,
	}
	appService := app.New(deps)

	if err := appService.Add(params); err != nil {
		fmt.Fprintf(shims.Stderr(), "%v\n", err)
		shims.Exit(1)
	}
}

func getHomeDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(shims.Stderr(), "could not determine user home directory\n")
		shims.Exit(1)
	}
	return dir
}

func findPrivateKeyFile() string {
	dir := getHomeDir()
	modernFilePath := filepath.Join(dir, ".ssh", "id_ed25519")
	if utils.Exists(modernFilePath) {
		return modernFilePath
	}
	legacyFilePath := filepath.Join(dir, ".ssh", "id_rsa")
	if utils.Exists(legacyFilePath) {
		return legacyFilePath
	}
	fmt.Fprintf(shims.Stderr(), "could not locate ssh key file; please specify '--private-key'\n")
	shims.Exit(1)
	return ""
}
