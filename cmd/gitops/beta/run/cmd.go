package run

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/weaveworks/weave-gitops/pkg/fluxexec"
	"github.com/weaveworks/weave-gitops/pkg/fluxinstall"
	"github.com/weaveworks/weave-gitops/pkg/run/install"
	"github.com/weaveworks/weave-gitops/pkg/run/watch"

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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/weaveworks/weave-gitops/pkg/run/ui"
)

const (
	dashboardName    = "ww-gitops"
	dashboardPodName = "ww-gitops-weave-gitops"
	adminUsername    = "admin"
)

var HelmChartVersion = "3.0.0"

var LogOutput bytes.Buffer

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
	cmdFlags.DurationVar(&flags.Timeout, "timeout", 5*time.Minute, "The timeout for operations during GitOps Run.")
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

func logMsg(msg string) {
	ui.Program.Send(ui.LogMsg{Msg: msg})
}

func logAdditionalMsg(msg string) {
	ui.Program.Send(ui.AdditionalLogMsg{Msg: msg})
}

func logErr(err error, shouldExit bool) tea.Msg {
	return ui.LogErrMsg{Err: err, ShouldExit: shouldExit}
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

		go func() error {
			gitRepoRoot, err := install.FindGitRepoDir()
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

			relativePath, err := install.GetRelativePathToRootDir(rootDir, targetPath)
			if err != nil { // if there is no git repo, we return an error
				return err
			}

			log := clilogger.NewCLILogger(&LogOutput)
			// log := clilogger.NewCLILogger(os.Stdout)

			if flags.KubeConfig != "" {
				kubeConfigArgs.KubeConfig = &flags.KubeConfig

				if flags.Context == "" {
					logMsg("A context should be provided if a kubeconfig is provided")
					// log.Failuref("A context should be provided if a kubeconfig is provided")
					return cmderrors.ErrNoContextForKubeConfig
				}
			}

			logMsg("Checking for a cluster in the kube config ...")
			// log.Actionf("Checking for a cluster in the kube config ...")

			var contextName string

			if flags.Context != "" {
				contextName = flags.Context
			} else {
				_, contextName, err = kube.RestConfig()
				if err != nil {
					logMsg(fmt.Sprintf("Error getting a restconfig: %v", err.Error()))
					// log.Failuref("Error getting a restconfig: %v", err.Error())
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
				logMsg(fmt.Sprintf("Explicitly allow GitOps Run on %s context", contextName))
				// log.Actionf("Explicitly allow GitOps Run on %s context", contextName)
			} else if !run.IsLocalCluster(kubeClient) {
				return fmt.Errorf("to run against a remote cluster, use --allow-k8s-context=%s", contextName)
			}

			ctx := context.Background()

			logMsg("Checking if Flux is already installed ...")
			// log.Actionf("Checking if Flux is already installed ...")

			fluxVersion := ""

			if fluxVersion, err = install.GetFluxVersion(log, ctx, kubeClient); err != nil {
				logMsg(fmt.Sprintf("Flux is not found: %v", err.Error()))
				// log.Warningf("Flux is not found: %v", err.Error())

				product := fluxinstall.NewProduct(flags.FluxVersion)

				installer := fluxinstall.NewInstaller()

				execPath, err := installer.Ensure(ctx, product)
				if err != nil {
					execPath, err = installer.Install(ctx, product)
					if err != nil {
						return err
					}
				}

				wd, err := os.Getwd()
				if err != nil {
					return err
				}

				flux, err := fluxexec.NewFlux(wd, execPath)
				if err != nil {
					return err
				}

				flux.SetStdout(&LogOutput)
				flux.SetStderr(&LogOutput)

				// flux.SetStdout(os.Stdout)
				// flux.SetStderr(os.Stderr)

				var components []fluxexec.Component
				for _, component := range flags.Components {
					components = append(components, fluxexec.Component(component))
				}

				var componentsExtra []fluxexec.ComponentExtra
				for _, component := range flags.ComponentsExtra {
					componentsExtra = append(componentsExtra, fluxexec.ComponentExtra(component))
				}

				if err := flux.Install(ctx,
					fluxexec.Components(components...),
					fluxexec.ComponentsExtra(componentsExtra...),
					fluxexec.WithGlobalOptions(
						fluxexec.Namespace(flags.Namespace),
						fluxexec.Timeout(flags.Timeout),
					),
				); err != nil {
					return err
				}
			} else {
				logMsg(fmt.Sprintf("Flux version %s is found", fluxVersion))
				// log.Successf("Flux version %s is found", fluxVersion)
			}

			logMsg("Checking if GitOps Dashboard is already installed ...")
			// log.Actionf("Checking if GitOps Dashboard is already installed ...")

			dashboardInstalled := install.IsDashboardInstalled(log, ctx, kubeClient, dashboardName, flags.Namespace)

			if dashboardInstalled {
				logMsg("GitOps Dashboard is found")
				// log.Successf("GitOps Dashboard is found")
			} else {
				prompt := promptui.Prompt{
					Label:     "Would you like to install the GitOps Dashboard",
					IsConfirm: true,
					Default:   "Y",
				}

				result, err := prompt.Run()
				if err == nil && strings.ToUpper(result) == "Y" {
					password, err := install.ReadPassword(log)
					if err != nil {
						return err
					}

					secret, err := install.GenerateSecret(log, password)
					if err != nil {
						return err
					}

					man, err := install.NewManager(log, ctx, kubeClient, kubeConfigArgs)
					if err != nil {
						logMsg("Error creating resource manager")
						// log.Failuref("Error creating resource manager")
						return err
					}

					manifests, err := install.CreateDashboardObjects(log, dashboardName, flags.Namespace, adminUsername, secret, HelmChartVersion)
					if err != nil {
						return fmt.Errorf("error creating dashboard objects: %w", err)
					}

					err = install.InstallDashboard(log, ctx, man, manifests)
					if err != nil {
						return fmt.Errorf("gitops dashboard installation failed: %w", err)
					} else {
						dashboardInstalled = true

						logMsg("GitOps Dashboard has been installed")
						// log.Successf("GitOps Dashboard has been installed")
					}
				}
			}

			if dashboardInstalled {
				logMsg(fmt.Sprintf("Request reconciliation of dashboard (timeout %v) ...", flags.Timeout))
				// log.Actionf("Request reconciliation of dashboard (timeout %v) ...", flags.Timeout)

				if err := install.ReconcileDashboard(kubeClient, dashboardName, flags.Namespace, dashboardPodName, flags.Timeout); err != nil {
					logMsg(fmt.Sprintf("Error requesting reconciliation of dashboard: %v", err.Error()))
					// log.Failuref("Error requesting reconciliation of dashboard: %v", err.Error())
				} else {
					logMsg("Dashboard reconciliation is done.")
					// log.Successf("Dashboard reconciliation is done.")
				}

				logAdditionalMsg("Dashboard: http://localhost:9001")
			}

			cancelDevBucketPortForwarding, err := watch.InstallDevBucketServer(log, kubeClient, cfg)
			if err != nil {
				return err
			}

			logAdditionalMsg("Podinfo: http://localhost:9000")

			var cancelDashboardPortForwarding func() = nil

			if dashboardInstalled {
				cancelDashboardPortForwarding, err = watch.EnablePortForwardingForDashboard(log, kubeClient, cfg, flags.Namespace, dashboardPodName, flags.DashboardPort)
				if err != nil {
					return err
				}
			}

			if err := watch.InitializeTargetDir(targetPath); err != nil {
				return fmt.Errorf("couldn't set up against target %s: %w", targetPath, err)
			}

			if err := watch.SetupBucketSourceAndKS(log, kubeClient, flags.Namespace, relativePath, flags.Timeout); err != nil {
				return err
			}

			ignorer := watch.CreateIgnorer(rootDir)

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

			err = filepath.Walk(rootDir, watch.WatchDirsForFileWalker(watcher, ignorer))
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
							logMsg(fmt.Sprintf("Error: %v", err))
							// log.Failuref("Error: %v", err)
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
							logMsg(fmt.Sprintf("%d change events detected", counter))
							// log.Actionf("%d change events detected", counter)

							// reset counter
							atomic.StoreUint64(&counter, 0)

							if err := watch.SyncDir(log, rootDir, "dev-bucket", minioClient, ignorer); err != nil {
								logMsg(fmt.Sprintf("Error syncing dir: %v", err))
								// log.Failuref("Error syncing dir: %v", err)
							}

							if needToRescan {
								// close the old watcher
								if err := watcher.Close(); err != nil {
									logMsg(fmt.Sprintf("Error closing the old watcher: %v", err))
									// log.Warningf("Error closing the old watcher: %v", err)
								}
								// create a new watcher
								watcher, err = fsnotify.NewWatcher()
								if err != nil {
									logMsg(fmt.Sprintf("Error creating new watcher: %v", err))
									// log.Failuref("Error creating new watcher: %v", err)
								}

								err = filepath.Walk(rootDir, watch.WatchDirsForFileWalker(watcher, ignorer))
								if err != nil {
									logMsg(fmt.Sprintf("Error re-walking dir: %v", err))
									// log.Failuref("Error re-walking dir: %v", err)
								}

								needToRescan = false
							}

							logMsg(fmt.Sprintf("Request reconciliation of dev-bucket, and dev-ks (timeout %v) ... ", flags.Timeout))
							// log.Actionf("Request reconciliation of dev-bucket, and dev-ks (timeout %v) ... ", flags.Timeout)

							if err := watch.ReconcileDevBucketSourceAndKS(log, kubeClient, flags.Namespace, flags.Timeout); err != nil {
								logMsg(fmt.Sprintf("Error requesting reconciliation: %v", err))
								// log.Failuref("Error requesting reconciliation: %v", err)
							}

							logMsg("Reconciliation is done.")
							// log.Successf("Reconciliation is done.")

							if flags.PortForward != "" {
								specMap, err := watch.ParsePortForwardSpec(flags.PortForward)
								if err != nil {
									logMsg(fmt.Sprintf("Error parsing port forward spec: %v", err))
									// log.Failuref("Error parsing port forward spec: %v", err)
								}

								// get pod from specMap
								namespacedName := types.NamespacedName{Namespace: specMap.Namespace, Name: specMap.Name}

								pod, err := run.GetPodFromResourceDescription(namespacedName, specMap.Kind, kubeClient)
								if err != nil {
									logMsg(fmt.Sprintf("Error getting pod from specMap: %v", err))
									// log.Failuref("Error getting pod from specMap: %v", err)
								}

								if pod != nil {
									waitFwd := make(chan struct{}, 1)
									readyChannel := make(chan struct{})
									cancelPortFwd = func() {
										close(waitFwd)

										cancelPortFwd = nil
									}

									logMsg(fmt.Sprintf("Port forwarding to pod %s/%s ...", pod.Namespace, pod.Name))
									// log.Actionf("Port forwarding to pod %s/%s ...", pod.Namespace, pod.Name)

									// this function _BLOCKS_ until the stopChannel (waitPwd) is closed.
									if err := watch.ForwardPort(pod, cfg, specMap, waitFwd, readyChannel); err != nil {
										logMsg(fmt.Sprintf("Error forwarding port: %v", err))
										// log.Failuref("Error forwarding port: %v", err)
									}

									logMsg("Port forwarding is stopped.")
									// log.Successf("Port forwarding is stopped.")
								}
							}
						}
					}
				}
			}()

			// wait for interrupt or ctrl+C
			logMsg("Press Ctrl+C to stop GitOps Run ...")
			// log.Waitingf("Press Ctrl+C to stop GitOps Run ...")

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			<-sigs

			if err := watcher.Close(); err != nil {
				logMsg(fmt.Sprintf("Error closing watcher: %v", err.Error()))
				// log.Warningf("Error closing watcher: %v", err.Error())
			}

			// print a blank line to make it easier to read the logs
			cancelDevBucketPortForwarding()

			if cancelDashboardPortForwarding != nil {
				cancelDashboardPortForwarding()
			}

			ticker.Stop()

			if err := watch.CleanupBucketSourceAndKS(log, kubeClient, flags.Namespace); err != nil {
				return err
			}

			// uninstall dev-bucket server
			if err := watch.UninstallDevBucketServer(log, kubeClient); err != nil {
				return err
			}

			// // run bootstrap wizard only if Flux is not installed and env var is set
			// if fluxVersion != "" || os.Getenv("GITOPS_RUN_BOOTSTRAP") == "" {
			// 	return nil
			// }

			// // re-enable listening for ctrl+C
			// signal.Reset(sig)

			// // parse remote
			// repo, err := bootstrap.ParseGitRemote(log, rootDir)
			// if err != nil {
			// 	log.Failuref("Error parsing Git remote: %v", err.Error())
			// }

			// // run the bootstrap wizard
			// log.Actionf("Starting bootstrap wizard ...")

			// log.Waitingf("Press Ctrl+C to stop bootstrap wizard ...")

			// remoteURL, err := bootstrap.ParseRemoteURL(repo)
			// if err != nil {
			// 	log.Failuref("Error parsing remote URL: %v", err.Error())
			// }

			// var gitProvider bootstrap.GitProvider

			// if remoteURL == "" {
			// 	gitProvider, err = bootstrap.SelectGitProvider(log)
			// 	if err != nil {
			// 		log.Failuref("Error selecting git provider: %v", err.Error())
			// 	}
			// } else {
			// 	urlParts := bootstrap.GetURLParts(remoteURL)

			// 	if len(urlParts) > 0 {
			// 		gitProvider = bootstrap.ParseGitProvider(urlParts[0])
			// 	}
			// }

			// if gitProvider == bootstrap.GitProviderUnknown {
			// 	gitProvider, err = bootstrap.SelectGitProvider(log)
			// 	if err != nil {
			// 		log.Failuref("Error selecting git provider: %v", err.Error())
			// 	}
			// }

			// path := filepath.Join(relativePath, "clusters", "my-cluster")
			// path = "./" + path

			// wizard, err := bootstrap.NewBootstrapWizard(log, remoteURL, gitProvider, repo, path)
			// if err != nil {
			// 	return fmt.Errorf("error creating bootstrap wizard: %v", err.Error())
			// }

			// if err = wizard.Run(log); err != nil {
			// 	return fmt.Errorf("error running bootstrap wizard: %v", err.Error())
			// }

			// _ = wizard.BuildCmd(log)

			// log.Successf("Flux bootstrap command successfully built.")

			return nil
		}()

		model := ui.InitialUIModel()

		ui.Program = tea.NewProgram(model)
		// Program = tea.NewProgram(model, tea.WithAltScreen())

		err = ui.Program.Start()
		if err != nil {
			return fmt.Errorf("could not start tea program: %v", err.Error())
		}

		// time.Sleep(600 * time.Second)

		return nil
	}
}
