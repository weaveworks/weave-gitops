package dashboard

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	clilogger "github.com/weaveworks/weave-gitops/cmd/gitops/logger"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/run"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	helmChartName        = "weave-gitops"
	defaultAdminUsername = "admin"
)

type DashboardCommandFlags struct {
	Version string
	// Create command flags.
	Export  bool
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

		name := args[0]
		if !validateObjectName(name) {
			return fmt.Errorf("name '%s' is invalid, it should adhere to standard defined in RFC 1123, the name can only contain alphanumeric characters or '-'", name)
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

		if flags.Export, err = cmd.Flags().GetBool("export"); err != nil {
			return err
		}

		if flags.Timeout, err = cmd.Flags().GetDuration("timeout"); err != nil {
			return err
		}

		var output io.Writer

		if flags.Export {
			output = &bytes.Buffer{}
		} else {
			output = os.Stdout
		}

		log := clilogger.NewCLILogger(output)

		log.Generatef("Generating GitOps Dashboard manifests ...")

		var secret string

		if flags.Password != "" {
			secret, err = run.GenerateSecret(log, flags.Password)
			if err != nil {
				return err
			}
		}

		dashboardName := args[0]

		adminUsername := flags.Username

		if adminUsername == "" && flags.Password != "" {
			adminUsername = defaultAdminUsername
		}

		manifests, err := run.CreateDashboardObjects(log, dashboardName, flags.Namespace, adminUsername, secret, flags.Version)
		if err != nil {
			return fmt.Errorf("error creating dashboard objects: %w", err)
		} else {
			log.Successf("Generated GitOps Dashboard manifests")
			fmt.Println("---")
			fmt.Println(resourceToString(manifests))
		}

		if flags.Export {
			return nil
		}

		log.Actionf("Checking for a cluster in the kube config ...")

		_, contextName, err := kube.RestConfig()
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

		log.Actionf("Checking if Flux is already installed ...")

		ctx, cancel := context.WithTimeout(context.Background(), flags.Timeout)
		defer cancel()

		if fluxVersion, err := run.GetFluxVersion(log, ctx, kubeClient); err != nil {
			log.Failuref("Flux is not found")
			return err
		} else {
			log.Successf("Flux version %s is found", fluxVersion)
		}

		log.Actionf("Applying GitOps Dashboard manifests")

		man, err := run.NewManager(log, ctx, kubeClient, kubeConfigArgs)
		if err != nil {
			log.Failuref("Error creating resource manager")
			return err
		}

		err = run.InstallDashboard(log, ctx, man, manifests)
		if err != nil {
			return fmt.Errorf("gitops dashboard installation failed: %w", err)
		} else {
			log.Successf("GitOps Dashboard has been installed")
		}

		log.Actionf("Request reconciliation of dashboard (timeout %v) ...", flags.Timeout)

		log.Waitingf("Waiting for GitOps Dashboard reconciliation")

		// dashboardPort := "9001"
		// dashboardPodName := dashboardName + "-" + helmChartName

		// if err := run.ReconcileDashboard(kubeClient, dashboardName, flags.Namespace, dashboardPodName, flags.Timeout, dashboardPort); err != nil {
		// 	log.Failuref("Error requesting reconciliation of dashboard: %v", err.Error())
		// } else {
		// 	log.Successf("GitOps Dashboard %s is ready", dashboardName)
		// }

		log.Successf("Installed GitOps Dashboard")
		// log.Successf("Installed GitOps Dashboard version %s", kustomization.Status.LastAppliedRevision)

		return nil
	}
}

func validateObjectName(name string) bool {
	r := regexp.MustCompile(`^[a-z0-9]([a-z0-9\\-]){0,61}[a-z0-9]$`)
	return r.MatchString(name)
}

func resourceToString(data []byte) string {
	data = bytes.ReplaceAll(data, []byte("  creationTimestamp: null\n"), []byte(""))
	data = bytes.ReplaceAll(data, []byte("status: {}\n"), []byte(""))

	return string(data)
}
