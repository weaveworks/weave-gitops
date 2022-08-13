package dashboard

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	clilogger "github.com/weaveworks/weave-gitops/cmd/gitops/logger"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/run"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type DashboardCommandFlags struct {
	Version string
	Export  string
	Timeout time.Duration
	// Overriden global flags.
	Username string
	Password string
	// Global flags.
	Namespace  string
	KubeConfig string
	// Flags, created by genericclioptions.
	Context string
}

var flags DashboardCommandFlags

var kubeConfigArgs *genericclioptions.ConfigFlags

func DashboardCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Create a HelmRepository and HelmRelease to deploy Weave GitOps",
		Long:  "Create a HelmRepository and HelmRelease to deploy Weave GitOps",
		Example: `
# Create a HelmRepository and HelmRelease to deploy Weave GitOps
gitops create dashboard ww-gitops \
  --password=$PASSWORD \
  --export > ./clusters/my-cluster/weave-gitops-dashboard.yaml
		`,
		SilenceUsage:      true,
		SilenceErrors:     true,
		PreRunE:           createDashboardCommandPreRunE(&opts.Endpoint),
		RunE:              createDashboardCommandRunE(opts),
		DisableAutoGenTag: true,
	}

	cmdFlags := cmd.Flags()

	cmdFlags.StringVar(&flags.Username, "username", "admin", "The username of the admin user.")
	cmdFlags.StringVar(&flags.Password, "password", "", "The password of the admin user.")
	cmdFlags.StringVar(&flags.Version, "version", "", "The version of the dashboard that should be installed.")
	cmdFlags.StringVar(&flags.Export, "export", "", "The path to export manifests to.")
	cmdFlags.DurationVar(&flags.Timeout, "timeout", 30*time.Second, "The timeout for operations during GitOps Run.")

	kubeConfigArgs = run.GetKubeConfigArgs()

	kubeConfigArgs.AddFlags(cmd.Flags())

	return cmd
}

func createDashboardCommandPreRunE(endpoint *string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		numArgs := len(args)

		if numArgs == 0 {
			return cmderrors.ErrNoName
		}

		if numArgs > 1 {
			return cmderrors.ErrMultipleNames
		}

		return nil
	}
}

func createDashboardCommandRunE(opts *config.Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		if flags.Namespace, err = cmd.Flags().GetString("namespace"); err != nil {
			return err
		}

		kubeConfigArgs.Namespace = &flags.Namespace

		if flags.KubeConfig, err = cmd.Flags().GetString("kubeconfig"); err != nil {
			return err
		}

		if flags.Context, err = cmd.Flags().GetString("context"); err != nil {
			return err
		}

		gitRepoRoot, err := run.FindGitRepoDir()
		if err != nil {
			return err
		}

		rootDir := gitRepoRoot

		// check if rootDir is valid
		if _, err := os.Stat(rootDir); err != nil {
			return fmt.Errorf("root directory %s does not exist", rootDir)
		}

		// find absolute path of the root Dir
		rootDir, err = filepath.Abs(rootDir)
		if err != nil {
			return err
		}

		currentDir, err := os.Getwd()
		if err != nil {
			return err
		}

		targetPath, err := filepath.Abs(filepath.Join(currentDir, args[0]))
		if err != nil {
			return err
		}

		_, err = run.GetRelativePathToRootDir(rootDir, targetPath)
		// relativePathForKs, err := run.GetRelativePathToRootDir(rootDir, targetPath)
		if err != nil { // if there is no git repo, we return an error
			return err
		}

		log := clilogger.NewCLILogger(os.Stdout)

		log.Actionf("Checking for a cluster in the kube config ...")

		var contextName string

		_, contextName, err = kube.RestConfig()
		if err != nil {
			log.Failuref("Error getting a restconfig: %v", err.Error())
			return cmderrors.ErrNoCluster
		}

		cfg, err := kubeConfigArgs.ToRESTConfig()
		if err != nil {
			return fmt.Errorf("error getting a restconfig from kube config args: %w", err)
		}

		kubeClientOpts := run.GetKubeClientOptions()
		kubeClientOpts.BindFlags(cmd.Flags())

		kubeClient, err := run.GetKubeClient(log, contextName, cfg, kubeClientOpts)
		if err != nil {
			return cmderrors.ErrGetKubeClient
		}

		ctx := context.Background()

		log.Actionf("Checking if Flux is already installed ...")

		if fluxVersion, err := run.GetFluxVersion(log, ctx, kubeClient); err != nil {
			log.Warningf("Flux is not found")
			return err
		} else {
			log.Successf("Flux version %s is found", fluxVersion)
		}

		log.Actionf("Checking if GitOps Dashboard is already installed ...")

		dashboardInstalled := run.IsDashboardInstalled(log, ctx, kubeClient, flags.Namespace)

		if dashboardInstalled {
			log.Successf("GitOps Dashboard is found")
		}

		// else {
		// 	prompt := promptui.Prompt{
		// 		Label:     "Would you like to install the GitOps Dashboard",
		// 		IsConfirm: true,
		// 		Default:   "Y",
		// 	}
		// 	_, err = prompt.Run()
		// 	if err == nil {
		// 		secret, err := run.GenerateSecret(log)
		// 		if err != nil {
		// 			return err
		// 		}

		// 		man, err := run.NewManager(log, ctx, kubeClient, kubeConfigArgs)
		// 		if err != nil {
		// 			log.Failuref("Error creating resource manager")
		// 			return err
		// 		}

		// 		err = run.InstallDashboard(log, ctx, man, flags.Namespace, secret)
		// 		if err != nil {
		// 			return fmt.Errorf("gitops dashboard installation failed: %w", err)
		// 		} else {
		// 			dashboardInstalled = true

		// 			log.Successf("GitOps Dashboard has been installed")
		// 		}
		// 	}
		// }

		if dashboardInstalled {
			timeout := time.Minute * 3

			log.Actionf("Request reconciliation of dashboard (timeout %v) ...", timeout)

			dashboardPort := "9001"

			if err := run.ReconcileDashboard(kubeClient, flags.Namespace, timeout, dashboardPort); err != nil {
				log.Failuref("Error requesting reconciliation of dashboard: %v", err.Error())
			} else {
				log.Successf("Dashboard reconciliation is done.")
			}
		}

		return nil
	}
}
