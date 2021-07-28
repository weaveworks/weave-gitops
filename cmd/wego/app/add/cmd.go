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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/wego/version"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"golang.org/x/term"
)

var params app.AddParams

var Cmd = &cobra.Command{
	Use:   "add [--name <name>] [--url <url>] [--branch <branch>] [--path <path within repository>] [--private-key <keyfile>] <repository directory>",
	Short: "Add a workload repository to a wego cluster",
	Long: strings.TrimSpace(dedent.Dedent(`
        Associates an additional application in a git repository with a wego cluster so that its contents may be managed via GitOps
    `)),
	Example: `
  # Add application to wego control from local git repository
  wego app add .

  # Add podinfo application to wego control from github repository
  wego app add --url git@github.com:myorg/podinfo

  # Get status of podinfo application
  wego app status podinfo
`,
	RunE:          runCmd,
	SilenceUsage:  true,
	SilenceErrors: true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

func init() {
	Cmd.Flags().StringVar(&params.Name, "name", "", "Name of application")
	Cmd.Flags().StringVar(&params.Url, "url", "", "URL of remote repository")
	Cmd.Flags().StringVar(&params.Path, "path", "./", "Path of files within git repository")
	Cmd.Flags().StringVar(&params.Branch, "branch", "main", "Branch to watch within git repository")
	Cmd.Flags().StringVar(&params.DeploymentType, "deployment-type", "kustomize", "deployment type [kustomize, helm]")
	Cmd.Flags().StringVar(&params.Chart, "chart", "", "Specify chart for helm source")
	Cmd.Flags().StringVar(&params.PrivateKey, "private-key", "", "Private key to access git repository over ssh")
	Cmd.Flags().StringVar(&params.AppConfigUrl, "app-config-url", "", "URL of external repository (if any) which will hold automation manifests; NONE to store only in the cluster")
	Cmd.Flags().BoolVar(&params.DryRun, "dry-run", false, "If set, 'wego add' will not make any changes to the system; it will just display the actions that would have been taken")
	Cmd.Flags().BoolVar(&params.AutoMerge, "auto-merge", false, "If set, 'wego add' will merge automatically into the set --branch")
}

func runCmd(cmd *cobra.Command, args []string) error {
	params.Namespace, _ = cmd.Parent().Flags().GetString("namespace")

	if params.Url != "" && len(args) > 0 {
		return fmt.Errorf("you should choose either --url or the app directory")
	}

	if len(args) > 0 {
		path, err := filepath.Abs(args[0])
		if err != nil {
			return errors.Wrap(err, "failed to get absolute path for the repo directory")
		}

		params.Dir = path
	}

	if strings.HasPrefix(params.PrivateKey, "~/") {
		dir, err := getHomeDir()
		if err != nil {
			return err
		}
		params.PrivateKey = filepath.Join(dir, params.PrivateKey[2:])
	} else if params.PrivateKey == "" {
		privateKey, err := findPrivateKeyFile()
		if err != nil {
			return err
		}
		params.PrivateKey = privateKey
	}

	authMethod, err := ssh.NewPublicKeysFromFile("git", params.PrivateKey, "")
	if err != nil {
		fmt.Print("Private Key Password: ")
		pw, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return errors.Wrap(err, "failed reading ssh key password")
		}

		authMethod, err = ssh.NewPublicKeysFromFile("git", params.PrivateKey, string(pw))
		if err != nil {
			return errors.Wrap(err, "failed reading ssh keys")
		}
	}

	params, err = setGitProviderToken(params)
	if err != nil {
		return err
	}

	cliRunner := &runner.CLIRunner{}
	osysClient := osys.New()
	fluxClient := flux.New(osysClient, cliRunner)
	kubeClient := kube.New(cliRunner)
	gitClient := git.New(authMethod)
	logger := logger.New(os.Stdout)

	appService := app.New(logger, gitClient, fluxClient, kubeClient, osysClient)

	utils.SetCommmitMessageFromArgs("wego app add", params.Url, params.Path, params.Name)

	if err := appService.Add(params); err != nil {
		return errors.Wrapf(err, "failed to add the app %s", params.Name)
	}

	return nil
}

func getHomeDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine user home directory")
	}
	return dir, nil
}

func findPrivateKeyFile() (string, error) {
	dir, err := getHomeDir()
	if err != nil {
		return "", err
	}

	modernFilePath := filepath.Join(dir, ".ssh", "id_ed25519")
	if utils.Exists(modernFilePath) {
		return modernFilePath, nil
	}

	legacyFilePath := filepath.Join(dir, ".ssh", "id_rsa")
	if utils.Exists(legacyFilePath) {
		return legacyFilePath, nil
	}

	return "", fmt.Errorf("could not locate ssh key file; please specify '--private-key'")
}

func setGitProviderToken(params app.AddParams) (app.AddParams, error) {
	providerToken, found := os.LookupEnv("GITHUB_TOKEN")
	if !found {
		return params, fmt.Errorf("GITHUB_TOKEN not set in environment")
	}

	params.GitProviderToken = providerToken

	return params, nil
}
