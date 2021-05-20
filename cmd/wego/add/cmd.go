package add

// Provides support for adding a repository of manifests to a wego cluster. If the cluster does not have
// wego installed, the user will be prompted to install wego and then the repository will be added.

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/cmdimpl"
)

var params cmdimpl.AddParamSet

var Cmd = &cobra.Command{
	Use:   "add [--name <name>] [--url <url>] [--branch <branch>] [--path <path within repository>] [--private-key <keyfile>] <repository directory>",
	Short: "Add a workload repository to a wego cluster",
	Long: strings.TrimSpace(dedent.Dedent(`
        Associates an additional git repository with a wego cluster so that its contents may be managed via GitOps
    `)),
	Example: "wego add .",
	Run:     runCmd,
}

func init() {
	Cmd.Flags().StringVar(&params.Owner, "owner", "", "Owner of remote git repository")
	Cmd.Flags().StringVar(&params.Name, "name", "", "Name of remote git repository")
	Cmd.Flags().StringVar(&params.Url, "url", "", "URL of remote git repository")
	Cmd.Flags().StringVar(&params.Path, "path", "./", "Path of files within git repository")
	Cmd.Flags().StringVar(&params.Branch, "branch", "main", "Branch to watch within git repository")
	Cmd.Flags().StringVar(&params.DeploymentType, "deployment-type", "kustomize", "deployment type [kustomize, helm]")
	Cmd.Flags().StringVar(&params.PrivateKey, "private-key", filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"), "Private key that provides access to git repository")
}

func runCmd(cmd *cobra.Command, args []string) {
	params.Namespace, _ = cmd.Parent().Flags().GetString("namespace")
	cmdimpl.Add(args, params)
}
