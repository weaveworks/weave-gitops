package dashboard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	// helmChartName = "weave-gitops"
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

		if flags.Export, err = cmd.Flags().GetBool("export"); err != nil {
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

		// var contextName string

		if !flags.Export {
			log.Actionf("Checking for a cluster in the kube config ...")

			_, _, err = kube.RestConfig()
			// _, contextName, err = kube.RestConfig()
			if err != nil {
				log.Failuref("Error getting a restconfig: %v", err.Error())
				return cmderrors.ErrNoCluster
			}
		}

		// cfg, err := kubeConfigArgs.ToRESTConfig()
		// if err != nil {
		// 	return fmt.Errorf("error getting a restconfig from kube config args: %w", err)
		// }

		// kubeClientOpts := run.GetKubeClientOptions()
		// kubeClientOpts.BindFlags(cmd.Flags())

		// kubeClient, err := run.GetKubeClient(log, contextName, cfg, kubeClientOpts)
		// if err != nil {
		// 	return cmderrors.ErrGetKubeClient
		// }

		// ctx := context.Background()

		// if createArgs.export {
		// 	return printExport(exportKs(&kustomization))
		// }

		// ctx, cancel := context.WithTimeout(context.Background(), rootArgs.timeout)
		// defer cancel()

		// kubeClient, err := utils.KubeClient(kubeconfigArgs, kubeclientOptions)
		// if err != nil {
		// 	return err
		// }

		// logger.Actionf("applying Kustomization")
		// namespacedName, err := upsertKustomization(ctx, kubeClient, &kustomization)
		// if err != nil {
		// 	return err
		// }

		// logger.Waitingf("waiting for Kustomization reconciliation")
		// if err := wait.PollImmediate(rootArgs.pollInterval, rootArgs.timeout,
		// 	isKustomizationReady(ctx, kubeClient, namespacedName, &kustomization)); err != nil {
		// 	return err
		// }
		// logger.Successf("Kustomization %s is ready", name)

		// logger.Successf("applied revision %s", kustomization.Status.LastAppliedRevision)
		// return nil

		//
		//
		//
		//
		//

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

		//
		//
		//

		// log.Actionf("Checking if Flux is already installed ...")

		// if fluxVersion, err := run.GetFluxVersion(log, ctx, kubeClient); err != nil {
		// 	log.Warningf("Flux is not found")
		// 	return err
		// } else {
		// 	log.Successf("Flux version %s is found", fluxVersion)
		// }

		// log.Actionf("Checking if GitOps Dashboard is already installed ...")

		// dashboardName := args[0]

		// dashboardInstalled := run.IsDashboardInstalled(log, ctx, kubeClient, dashboardName, flags.Namespace)

		// if dashboardInstalled {
		// 	log.Successf("GitOps Dashboard is found")
		// }

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

		// if dashboardInstalled {
		// 	timeout := time.Minute * 3

		// 	log.Actionf("Request reconciliation of dashboard (timeout %v) ...", timeout)

		// 	dashboardPort := "9001"
		// 	dashboardPodName := dashboardName + "-" + helmChartName

		// 	if err := run.ReconcileDashboard(kubeClient, dashboardName, flags.Namespace, dashboardPodName, timeout, dashboardPort); err != nil {
		// 		log.Failuref("Error requesting reconciliation of dashboard: %v", err.Error())
		// 	} else {
		// 		log.Successf("Dashboard reconciliation is done.")
		// 	}
		// }

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
