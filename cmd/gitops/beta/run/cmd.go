package run

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fluxcd/flux2/pkg/manifestgen/install"
	"github.com/fsnotify/fsnotify"
	"github.com/manifoldco/promptui"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	clilogger "github.com/weaveworks/weave-gitops/cmd/gitops/logger"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/run"
	"github.com/weaveworks/weave-gitops/pkg/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	dashboardName    = "ww-gitops"
	dashboardPodName = "ww-gitops-weave-gitops"
	adminUsername    = "admin"
	helmChartVersion = "3.0.0"
)

type RunCommandFlags struct {
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

var flags RunCommandFlags

var kubeConfigArgs *genericclioptions.ConfigFlags

func RunCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Set up an interactive sync between your cluster and your local file system",
		Long:  "This will set up a sync between the cluster in your kubeconfig and the path that you specify on your local filesystem.  If you do not have Flux installed on the cluster then this will add it to the cluster automatically.  This is a requirement so we can sync the files successfully from your local system onto the cluster.  Flux will take care of producing the objects for you.",
		Example: `
# Run the sync on the current working directory
gitops beta run . [flags]

# Run the sync against the dev overlay path
gitops beta run ./deploy/overlays/dev

# Run the sync on the dev directory and forward the port.
# Listen on port 8080 on localhost, forwarding to 5000 in a pod of the service app.
gitops beta run ./dev --port-forward port=8080:5000,resource=svc/app

# Run the sync on the dev directory with a specified root dir.
gitops beta run ./clusters/default/dev --root-dir ./clusters/default

# Run the sync on the podinfo demo.
git clone https://github.com/stefanprodan/podinfo
cd podinfo
gitops beta run ./deploy/overlays/dev --timeout 3m --port-forward namespace=dev,resource=svc/backend,port=9898:9898`,
		SilenceUsage:      true,
		SilenceErrors:     true,
		PreRunE:           betaRunCommandPreRunE(&opts.Endpoint),
		RunE:              betaRunCommandRunE(opts),
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

func betaRunCommandPreRunE(endpoint *string) func(*cobra.Command, []string) error {
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

func betaRunCommandRunE(opts *config.Options) func(*cobra.Command, []string) error {
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

		relativePathForKs, err := run.GetRelativePathToRootDir(rootDir, targetPath)
		if err != nil { // if there is no git repo, we return an error
			return err
		}

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

			if err := run.ValidateComponents(components); err != nil {
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

		dashboardInstalled := run.IsDashboardInstalled(log, ctx, kubeClient, dashboardName, flags.Namespace)

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

				err = run.InstallDashboard(log, ctx, man, dashboardName, flags.Namespace, adminUsername, secret, helmChartVersion)
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

			if err := run.ReconcileDashboard(kubeClient, dashboardName, flags.Namespace, dashboardPodName, flags.Timeout, flags.DashboardPort); err != nil {
				log.Failuref("Error requesting reconciliation of dashboard: %v", err.Error())
			} else {
				log.Successf("Dashboard reconciliation is done.")
			}
		}

		cancelDevBucketPortForwarding, err := run.InstallDevBucketServer(log, kubeClient, cfg)
		if err != nil {
			return err
		}

		var cancelDashboardPortForwarding func() = nil

		if dashboardInstalled {
			cancelDashboardPortForwarding, err = run.EnablePortForwardingForDashboard(log, kubeClient, cfg, flags.Namespace, dashboardPodName, flags.DashboardPort)
			if err != nil {
				return err
			}
		}

		if err := run.SetupBucketSourceAndKS(log, kubeClient, flags.Namespace, relativePathForKs, flags.Timeout); err != nil {
			return err
		}

		minioClient, err := minio.New(
			"localhost:9000",
			&minio.Options{
				Creds:        credentials.NewStaticV4("user", "doesn't matter", ""),
				Secure:       false,
				BucketLookup: minio.BucketLookupPath,
			},
		)
		if err != nil {
			return err
		}

		// watch for file changes in dir gitRepoRoot
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}

		err = filepath.Walk(rootDir, run.WatchDirsForFileWalker(watcher))
		if err != nil {
			return err
		}

		// cancel function to stop forwarding port
		var (
			cancelPortFwd func()
			counter       uint64 = 1
			needToRescan  bool   = false
		)
		// atomic counter for the number of file change events that have changed

		go func() {
			for {
				select {
				case event := <-watcher.Events:
					if event.Op&fsnotify.Create == fsnotify.Create ||
						event.Op&fsnotify.Remove == fsnotify.Remove ||
						event.Op&fsnotify.Rename == fsnotify.Rename {
						// if it's a dir, we need to watch it
						if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
							needToRescan = true
						}
					}

					if cancelPortFwd != nil {
						cancelPortFwd()
					}

					atomic.AddUint64(&counter, 1)
				case err := <-watcher.Errors:
					if err != nil {
						log.Failuref("Error: %v", err)
					}
				}
			}
		}()

		// event aggregation loop
		ticker := time.NewTicker(680 * time.Millisecond)

		go func() {
			for { // nolint:gosimple
				select {
				case <-ticker.C:
					if counter > 0 {
						log.Actionf("%d change events detected", counter)

						// reset counter
						atomic.StoreUint64(&counter, 0)

						if err := run.SyncDir(log, rootDir, "dev-bucket", minioClient); err != nil {
							log.Failuref("Error syncing dir: %v", err)
						}

						if needToRescan {
							// close the old watcher
							if err := watcher.Close(); err != nil {
								log.Warningf("Error closing the old watcher: %v", err)
							}
							// create a new watcher
							watcher, err = fsnotify.NewWatcher()
							if err != nil {
								log.Failuref("Error creating new watcher: %v", err)
							}

							err = filepath.Walk(rootDir, run.WatchDirsForFileWalker(watcher))
							if err != nil {
								log.Failuref("Error re-walking dir: %v", err)
							}

							needToRescan = false
						}

						log.Actionf("Request reconciliation of dev-bucket, and dev-ks (timeout %v) ... ", flags.Timeout)

						if err := run.ReconcileDevBucketSourceAndKS(log, kubeClient, flags.Namespace, flags.Timeout); err != nil {
							log.Failuref("Error requesting reconciliation: %v", err)
						}

						log.Successf("Reconciliation is done.")

						if flags.PortForward != "" {
							specMap, err := run.ParsePortForwardSpec(flags.PortForward)
							if err != nil {
								log.Failuref("Error parsing port forward spec: %v", err)
							}

							// get pod from specMap
							pod, err := run.GetPodFromSpecMap(specMap, kubeClient)
							if err != nil {
								log.Failuref("Error getting pod from specMap: %v", err)
							}

							if pod != nil {
								waitFwd := make(chan struct{}, 1)
								readyChannel := make(chan struct{})
								cancelPortFwd = func() {
									close(waitFwd)

									cancelPortFwd = nil
								}

								log.Actionf("Port forwarding to pod %s/%s ...", pod.Namespace, pod.Name)

								// this function _BLOCKS_ until the stopChannel (waitPwd) is closed.
								if err := run.ForwardPort(pod, cfg, specMap, waitFwd, readyChannel); err != nil {
									log.Failuref("Error forwarding port: %v", err)
								}

								log.Successf("Port forwarding is stopped.")
							}
						}
					}
				}
			}
		}()

		// wait for interrupt or ctrl+C
		log.Waitingf("Press Ctrl+C to stop GitOps Run ...")

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs

		if err := watcher.Close(); err != nil {
			log.Warningf("Error closing watcher: %v", err.Error())
		}

		// print a blank line to make it easier to read the logs
		fmt.Println()
		cancelDevBucketPortForwarding()

		if cancelDashboardPortForwarding != nil {
			cancelDashboardPortForwarding()
		}

		ticker.Stop()

		if err := run.CleanupBucketSourceAndKS(log, kubeClient, flags.Namespace); err != nil {
			return err
		}

		// uninstall dev-bucket server
		if err := run.UninstallDevBucketServer(log, kubeClient); err != nil {
			return err
		}

		return nil
	}
}
