package dashboard

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fluxcd/flux2/pkg/manifestgen/install"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	clilogger "github.com/weaveworks/weave-gitops/cmd/gitops/logger"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/run"
	"github.com/weaveworks/weave-gitops/pkg/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/utils/strings/slices"
)

type DashboardCommandFlags struct {
	FluxVersion     string
	AllowK8sContext string
	Components      []string
	ComponentsExtra []string
	Timeout         time.Duration
	PortForward     string // port forward specifier, e.g. "port=8080:8080,resource=svc/app"
	DashboardPort   string
	RootDir         string
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

	cmdFlags.StringVar(&flags.FluxVersion, "flux-version", version.FluxVersion, "The version of Flux to install.")
	cmdFlags.StringVar(&flags.AllowK8sContext, "allow-k8s-context", "", "The name of the KubeConfig context to explicitly allow.")
	cmdFlags.StringSliceVar(&flags.Components, "components", []string{"source-controller", "kustomize-controller", "helm-controller", "notification-controller"}, "The Flux components to install.")
	cmdFlags.StringSliceVar(&flags.ComponentsExtra, "components-extra", []string{}, "Additional Flux components to install, allowed values are image-reflector-controller,image-automation-controller.")
	cmdFlags.DurationVar(&flags.Timeout, "timeout", 30*time.Second, "The timeout for operations during GitOps Run.")
	cmdFlags.StringVar(&flags.PortForward, "port-forward", "", "Forward the port from a cluster's resource to your local machine i.e. 'port=8080:8080,resource=svc/app'.")
	cmdFlags.StringVar(&flags.DashboardPort, "dashboard-port", "9001", "GitOps Dashboard port")
	cmdFlags.StringVar(&flags.RootDir, "root-dir", "", "Specify the root directory to watch for changes. If not specified, the root of Git repository will be used.")

	kubeConfigArgs = run.GetKubeConfigArgs()

	kubeConfigArgs.AddFlags(cmd.Flags())

	return cmd
}

func createDashboardCommandPreRunE(endpoint *string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		numArgs := len(args)

		if numArgs == 0 {
			return cmderrors.ErrNoFilePath
		}

		if numArgs > 1 {
			return cmderrors.ErrMultipleFilePaths
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

		rootDir := flags.RootDir
		if rootDir == "" {
			rootDir = gitRepoRoot
		}

		// check if rootDir is valid
		if _, err := os.Stat(rootDir); err != nil {
			return fmt.Errorf("root directory %s does not exist", rootDir)
		}

		// // find absolute path of the root Dir
		// rootDir, err = filepath.Abs(rootDir)
		// if err != nil {
		// 	return err
		// }

		// currentDir, err := os.Getwd()
		// if err != nil {
		// 	return err
		// }

		// targetPath, err := filepath.Abs(filepath.Join(currentDir, args[0]))
		// if err != nil {
		// 	return err
		// }

		// relativePathForKs, err := run.GetRelativePathToRootDir(rootDir, targetPath)
		// if err != nil { // if there is no git repo, we return an error
		// 	return err
		// }

		log := clilogger.NewCLILogger(os.Stdout)

		if flags.KubeConfig != "" {
			kubeConfigArgs.KubeConfig = &flags.KubeConfig

			if flags.Context == "" {
				log.Failuref("A context should be provided if a kubeconfig is provided")
				return cmderrors.ErrNoContextForKubeConfig
			}
		}

		log.Actionf("Checking for a cluster in the kube config ...")

		var contextName string

		if flags.Context != "" {
			contextName = flags.Context
		} else {
			_, contextName, err = kube.RestConfig()
			if err != nil {
				log.Failuref("Error getting a restconfig: %v", err.Error())
				return cmderrors.ErrNoCluster
			}
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

		contextName = kubeClient.ClusterName
		if flags.AllowK8sContext == contextName {
			log.Actionf("Explicitly allow GitOps Run on %s context", contextName)
		} else if !run.IsLocalCluster(kubeClient) {
			return fmt.Errorf("to run against a remote cluster, use --allow-k8s-context=%s", contextName)
		}

		ctx := context.Background()

		log.Actionf("Checking if Flux is already installed ...")

		if fluxVersion, err := run.GetFluxVersion(log, ctx, kubeClient); err != nil {
			log.Warningf("Flux is not found: %v", err.Error())

			components := flags.Components
			components = append(components, flags.ComponentsExtra...)

			if err := ValidateComponents(components); err != nil {
				return fmt.Errorf("can't install flux: %w", err)
			}

			installOpts := install.MakeDefaultOptions()
			installOpts.Version = flags.FluxVersion
			installOpts.Namespace = flags.Namespace
			installOpts.Components = components
			installOpts.ManifestFile = "flux-system.yaml"
			installOpts.Timeout = flags.Timeout

			man, err := run.NewManager(log, ctx, kubeClient, kubeConfigArgs)
			if err != nil {
				log.Failuref("Error creating resource manager")
				return err
			}

			if err := run.InstallFlux(log, ctx, installOpts, man); err != nil {
				return fmt.Errorf("flux installation failed: %w", err)
			} else {
				log.Successf("Flux has been installed")
			}

			for _, controllerName := range components {
				log.Actionf("Waiting for %s/%s to be ready ...", flags.Namespace, controllerName)

				if err := run.WaitForDeploymentToBeReady(log, kubeClient, controllerName, flags.Namespace); err != nil {
					return err
				}

				log.Successf("%s/%s is now ready ...", flags.Namespace, controllerName)
			}
		} else {
			log.Successf("Flux version %s is found", fluxVersion)
		}

		log.Actionf("Checking if GitOps Dashboard is already installed ...")

		dashboardInstalled := run.IsDashboardInstalled(log, ctx, kubeClient, flags.Namespace)

		if dashboardInstalled {
			log.Successf("GitOps Dashboard is found")
		} else {
			prompt := promptui.Prompt{
				Label:     "Would you like to install the GitOps Dashboard",
				IsConfirm: true,
				Default:   "Y",
			}
			_, err = prompt.Run()
			if err == nil {
				secret, err := run.GenerateSecret(log)
				if err != nil {
					return err
				}

				man, err := run.NewManager(log, ctx, kubeClient, kubeConfigArgs)
				if err != nil {
					log.Failuref("Error creating resource manager")
					return err
				}

				err = run.InstallDashboard(log, ctx, man, flags.Namespace, secret)
				if err != nil {
					return fmt.Errorf("gitops dashboard installation failed: %w", err)
				} else {
					dashboardInstalled = true

					log.Successf("GitOps Dashboard has been installed")
				}
			}
		}

		if dashboardInstalled {
			log.Actionf("Request reconciliation of dashboard (timeout %v) ...", flags.Timeout)

			if err := run.ReconcileDashboard(kubeClient, flags.Namespace, flags.Timeout, flags.DashboardPort); err != nil {
				log.Failuref("Error requesting reconciliation of dashboard: %v", err.Error())
			} else {
				log.Successf("Dashboard reconciliation is done.")
			}
		}

		return nil
	}
}

func ValidateComponents(components []string) error {
	defaults := install.MakeDefaultOptions()
	bootstrapAllComponents := append(defaults.Components, defaults.ComponentsExtra...)

	for _, component := range components {
		if !slices.Contains(bootstrapAllComponents, component) {
			return fmt.Errorf("component %s is not available", component)
		}
	}

	return nil
}
