package root

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run"
	"github.com/weaveworks/weave-gitops/pkg/run/install"
)

const (
	helmChartName        = "weave-gitops"
	defaultAdminUsername = "admin"
)

var createDashboardCommandFlags struct {
	// Override global flags.
	Username string
	Password string
}

// var kubeConfigArgs *genericclioptions.ConfigFlags

func createDashboardCommand(opts *config.Options) *cobra.Command {
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

	cmdFlags.StringVar(&createDashboardCommandFlags.Username, "username", "admin", "The username of the dashboard admin user.")
	cmdFlags.StringVar(&createDashboardCommandFlags.Password, "password", "", "The password of the dashboard admin user.")

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

		name := args[0]
		if !validateObjectName(name) {
			return fmt.Errorf("name '%s' is invalid, it should adhere to standard defined in RFC 1123, the name can only contain alphanumeric characters or '-'", name)
		}

		return nil
	}
}

func createDashboardCommandRunE(opts *config.Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var (
			err    error
			output io.Writer
		)

		if createCommandFlags.Export {
			output = io.Discard
		} else {
			output = os.Stdout
		}

		log := logger.NewCLILogger(output)

		log.Generatef("Generating GitOps Dashboard manifests ...")

		var passwordHash string

		if createDashboardCommandFlags.Password != "" {
			passwordHash, err = install.GeneratePasswordHash(log, createDashboardCommandFlags.Password)
			if err != nil {
				return err
			}
		}

		dashboardName := args[0]

		adminUsername := createDashboardCommandFlags.Username

		if adminUsername == "" && createDashboardCommandFlags.Password != "" {
			adminUsername = defaultAdminUsername
		}

		manifests, err := install.CreateDashboardObjects(log, dashboardName, *kubeconfigArgs.Namespace, adminUsername, passwordHash, "")
		if err != nil {
			return fmt.Errorf("error creating dashboard objects: %w", err)
		}

		log.Successf("Generated GitOps Dashboard manifests")

		if createCommandFlags.Export {
			fmt.Println("---")
			fmt.Println(string(manifests))

			return nil
		}

		// Installing the dashboard
		log.Actionf("Checking for a cluster in the kube config ...")

		var contextName string

		if kubeconfigArgs.Context != nil {
			_, contextName, err = kube.RestConfig()
			if err != nil {
				log.Failuref("Error getting a restconfig: %v", err.Error())
				return cmderrors.ErrNoCluster
			}
		} else {
			contextName = *kubeconfigArgs.Context
		}

		cfg, err := kubeconfigArgs.ToRESTConfig()
		if err != nil {
			return fmt.Errorf("error getting a restconfig from kube config args: %w", err)
		}

		kubeClient, err := run.GetKubeClient(log, contextName, cfg, kubeclientOptions)
		if err != nil {
			return cmderrors.ErrGetKubeClient
		}

		log.Actionf("Checking if Flux is already installed ...")

		ctx, cancel := context.WithTimeout(context.Background(), createCommandFlags.Timeout)
		defer cancel()

		if fluxVersion, err := install.GetFluxVersion(log, ctx, kubeClient); err != nil {
			log.Failuref("Flux is not found")
			return err
		} else {
			log.Successf("Flux version %s is found", fluxVersion)
		}

		log.Actionf("Applying GitOps Dashboard manifests")

		man, err := install.NewManager(log, ctx, kubeClient, kubeconfigArgs)
		if err != nil {
			log.Failuref("Error creating resource manager")
			return err
		}

		err = install.InstallDashboard(log, ctx, man, manifests)
		if err != nil {
			return fmt.Errorf("gitops dashboard installation failed: %w", err)
		} else {
			log.Successf("GitOps Dashboard has been installed")
		}

		log.Actionf("Request reconciliation of dashboard (timeout %v) ...", createCommandFlags.Timeout)

		log.Waitingf("Waiting for GitOps Dashboard reconciliation")

		dashboardPodName := dashboardName + "-" + helmChartName

		if err := install.ReconcileDashboard(ctx, kubeClient, dashboardName, *kubeconfigArgs.Namespace, dashboardPodName, createCommandFlags.Timeout); err != nil {
			log.Failuref("Error requesting reconciliation of dashboard: %v", err.Error())
		} else {
			log.Successf("GitOps Dashboard %s is ready", dashboardName)
		}

		log.Successf("Installed GitOps Dashboard")

		return nil
	}
}

func validateObjectName(name string) bool {
	r := regexp.MustCompile(`^[a-z0-9]([a-z0-9\\-]){0,61}[a-z0-9]$`)
	return r.MatchString(name)
}
