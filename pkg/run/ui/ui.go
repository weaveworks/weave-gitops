package ui

import (
	"bytes"
	"context"
	"fmt"
	"os"

	// "os/signal"
	"path/filepath"
	"strings"

	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/fluxexec"
	"github.com/weaveworks/weave-gitops/pkg/fluxinstall"
	"github.com/weaveworks/weave-gitops/pkg/logger"

	"github.com/weaveworks/weave-gitops/pkg/run/install"
	"github.com/weaveworks/weave-gitops/pkg/run/watch"

	"github.com/manifoldco/promptui"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	clilogger "github.com/weaveworks/weave-gitops/cmd/gitops/logger"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/run"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UIModel struct {
	Args           []string
	Flags          RunCommandFlags
	GitopsRunCmd   *cobra.Command
	KubeConfigArgs *genericclioptions.ConfigFlags

	fluxVersion string

	ready bool

	logViewport   viewport.Model
	inputViewport viewport.Model
	textInput     textinput.Model
	err           error

	width  int
	height int

	logs           []string
	additionalLogs []string
}

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

const (
	dashboardName    = "ww-gitops"
	dashboardPodName = "ww-gitops-weave-gitops"
	adminUsername    = "admin"
)

var watcher *fsnotify.Watcher

// HelmChartversion TODO: update setting HelmChartVersion var everywhere
var HelmChartVersion = "3.0.0"

var (
	kubeClient     *kube.KubeHTTP
	KubeConfigArgs *genericclioptions.ConfigFlags

	rootDir      string
	relativePath string

	log                           logger.Logger
	cancelDevBucketPortForwarding func()
	cancelDashboardPortForwarding func()
	ticker                        *time.Ticker

	Program   *tea.Program
	LogOutput bytes.Buffer
)

// UI styling
var (
	// add view styles
	logViewportStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#3492e5")).
				Foreground(lipgloss.Color("#050e16")).
				BorderForeground(lipgloss.Color("#191919"))
	inputViewportStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#28aec2")).
				Foreground(lipgloss.Color("#050e16")).
				BorderForeground(lipgloss.Color("#191919"))
)

func InitialUIModel() UIModel {
	return UIModel{}
}

type logMsg struct{ msg string }
type additionalLogMsg struct{ msg string }

type logErrMsg struct {
	err        error
	shouldExit bool
}

type runGitopsRunMsg struct{ msg string }

type stopAndCleanUpGitopsRunMsg struct{ msg string }

func (m UIModel) logMsg(msg string) {
	Program.Send(logMsg{msg: msg})
}

// func (m UIModel) logMsg(msg string) tea.Msg {
// 	return logMsg{msg: msg}
// }

func (m UIModel) logAdditionalMsg(msg string) {
	Program.Send(additionalLogMsg{msg: msg})
}

func (m UIModel) logErr(err error, shouldExit bool) tea.Msg {
	return logErrMsg{err: err, shouldExit: shouldExit}
}

// func (m UIModel) Init() tea.Cmd {
// 	return nil
// }

func (m UIModel) Init() tea.Cmd {
	return tea.Batch(
		m.RunGitopsRun,
	)
}

func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd      tea.Cmd
		logVpCmd   tea.Cmd
		inputVpCmd tea.Cmd
	)

	m.textInput, inputVpCmd = m.textInput.Update(msg)
	m.logViewport, logVpCmd = m.logViewport.Update(msg)
	m.inputViewport, inputVpCmd = m.inputViewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlQ, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyCtrlC:
			// cmd := func() tea.Msg {
			// 	return tea.Quit()
			// }

			cmd := func() tea.Msg {
				return stopAndCleanUpGitopsRunMsg{msg: "stopping gitops run"}
			}

			return m, tea.Batch(cmd)
		case tea.KeyEnter:
			// enter answer
		}
		// Window size is received when starting up and on every resize
	case tea.WindowSizeMsg:
		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.

			logHeight := int(float64(msg.Height) * 0.7)

			m.width = msg.Width
			m.height = msg.Height

			m.logViewport = viewport.New(msg.Width, logHeight)
			m.logViewport.YPosition = 20
			// m.logViewport.HighPerformanceRendering = true

			m.inputViewport = viewport.New(msg.Width, int(float64(msg.Height)*0.3))
			m.inputViewport.YPosition = logHeight
			// m.inputViewport.HighPerformanceRendering = true

			m.logViewport.Style = logViewportStyle
			m.inputViewport.Style = inputViewportStyle

			line := strings.Repeat(" ", msg.Width)

			m.logViewport.SetContent(line + "\n\n\n")
			m.inputViewport.SetContent(line + "\n\n")

			m.ready = true
		} else {
			m.width = msg.Width
			m.height = msg.Height

			m.logViewport.Width = msg.Width
			m.logViewport.Height = int(float64(msg.Height) * 0.7)

			m.inputViewport.Width = msg.Width
			m.inputViewport.Height = int(float64(msg.Height) * 0.3)

			m.logViewport.Style = logViewportStyle
			m.inputViewport.Style = inputViewportStyle

			line := strings.Repeat(" ", msg.Width)

			m.logViewport.SetContent(line + "\n\n\n" + strings.Join(m.logs, "\n"))
			m.inputViewport.SetContent(line + "\n\n" + strings.Join(m.additionalLogs, "\n"))
		}

		// if useHighPerformanceRenderer {
		// Render (or re-render) the whole viewport. Necessary both to
		// initialize the viewport and when the window is resized.
		//
		// This is needed for high-performance rendering only.
		// }
		// return m, nil
		return m, tea.Batch(tiCmd, logVpCmd, inputVpCmd)
		// return m, tea.Batch(tiCmd, logVpCmd, inputVpCmd, viewport.Sync(m.logViewport), viewport.Sync(m.inputViewport))
	case logMsg:
		m.logs = append(m.logs, msg.msg)
		line := strings.Repeat(" ", m.width)
		m.logViewport.SetContent(line + "\n\n\n" + strings.Join(m.logs, "\n"))
		m.logViewport.GotoBottom()

		return m, tea.Batch(tiCmd, logVpCmd, inputVpCmd)
		// return m, tea.Batch(tiCmd, logVpCmd, inputVpCmd, viewport.Sync(m.logViewport), viewport.Sync(m.inputViewport))
	case additionalLogMsg:
		m.additionalLogs = append(m.additionalLogs, msg.msg)

		line := strings.Repeat(" ", m.width)
		m.inputViewport.SetContent(line + "\n\n" + strings.Join(m.additionalLogs, "\n"))
		m.inputViewport.GotoBottom()

		return m, tea.Batch(tiCmd, logVpCmd, inputVpCmd)
		// return m, tea.Batch(tiCmd, logVpCmd, inputVpCmd, viewport.Sync(m.logViewport), viewport.Sync(m.inputViewport))
	case logErrMsg:
		fmt.Println(msg.err.Error())
		if msg.shouldExit {
			os.Exit(1)
		}
	case stopAndCleanUpGitopsRunMsg:
		cmd := func() tea.Msg {
			return m.StopAndCleanUpGitopsRun(m.fluxVersion)
		}

		return m, cmd
	}

	return m, nil
}

func (m UIModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	return fmt.Sprintf(
		"%s\n\n%s",
		m.logViewport.View(),
		m.inputViewport.View(),
	) + "\n\n"
}

func (m UIModel) RunGitopsRun() tea.Msg {
	// go func() tea.Cmd {
	cmd := m.GitopsRunCmd
	flags := m.Flags

	gitRepoRoot, err := install.FindGitRepoDir()
	if err != nil {
		return m.logErr(err, true)
		// return err
	}

	rootDir = flags.RootDir
	if rootDir == "" {
		rootDir = gitRepoRoot
	}

	// check if rootDir is valid
	if _, err := os.Stat(rootDir); err != nil {
		return m.logErr(fmt.Errorf("root directory %s does not exist", rootDir), true)
		// return fmt.Errorf("root directory %s does not exist", rootDir)
	}

	// find absolute path of the root Dir
	rootDir, err = filepath.Abs(rootDir)
	if err != nil {
		return m.logErr(err, true)
		// return err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return m.logErr(err, true)
		// return err
	}

	targetPath, err := filepath.Abs(filepath.Join(currentDir, m.Args[0]))
	if err != nil {
		return m.logErr(err, true)
		// return err
	}

	relativePath, err = install.GetRelativePathToRootDir(rootDir, targetPath)
	if err != nil { // if there is no git repo, we return an error
		return m.logErr(err, true)
		// return err
	}

	log = clilogger.NewCLILogger(&LogOutput)
	// log = clilogger.NewCLILogger(os.Stdout)

	if flags.KubeConfig != "" {
		m.KubeConfigArgs.KubeConfig = &flags.KubeConfig

		if flags.Context == "" {
			m.logMsg("A context should be provided if a kubeconfig is provided")
			// log.Failuref("A context should be provided if a kubeconfig is provided")
			return m.logErr(cmderrors.ErrNoContextForKubeConfig, true)
			// return cmderrors.ErrNoContextForKubeConfig
		}
	}

	log.Actionf("Checking for a cluster in the kube config ...")

	var contextName string

	if flags.Context != "" {
		contextName = flags.Context
	} else {
		_, contextName, err = kube.RestConfig()
		if err != nil {
			m.logMsg(fmt.Sprintf("Error getting a restconfig: %v", err.Error()))
			// log.Failuref("Error getting a restconfig: %v", err.Error())
			return m.logErr(cmderrors.ErrNoCluster, true)
			// return cmderrors.ErrNoCluster
		}
	}

	cfg, err := m.KubeConfigArgs.ToRESTConfig()
	if err != nil {
		return m.logErr(fmt.Errorf("error getting a restconfig from kube config args: %w", err), true)
		// return fmt.Errorf("error getting a restconfig from kube config args: %w", err)
	}

	kubeClientOpts := run.GetKubeClientOptions()
	kubeClientOpts.BindFlags(cmd.Flags())

	kubeClient, err = run.GetKubeClient(log, contextName, cfg, kubeClientOpts)
	if err != nil {
		return m.logErr(cmderrors.ErrGetKubeClient, true)
		// return cmderrors.ErrGetKubeClient
	}

	contextName = kubeClient.ClusterName
	if flags.AllowK8sContext == contextName {
		m.logMsg(fmt.Sprintf("Explicitly allow GitOps Run on %s context", contextName))
		// log.Actionf("Explicitly allow GitOps Run on %s context", contextName)
	} else if !run.IsLocalCluster(kubeClient) {
		return m.logErr(fmt.Errorf("to run against a remote cluster, use --allow-k8s-context=%s", contextName), true)
		// return fmt.Errorf("to run against a remote cluster, use --allow-k8s-context=%s", contextName)
	}

	ctx := context.Background()

	m.logMsg("Checking if Flux is already installed ...")
	// log.Actionf("Checking if Flux is already installed ...")

	if m.fluxVersion, err = install.GetFluxVersion(log, ctx, kubeClient); err != nil {
		m.logMsg(fmt.Sprintf("Flux is not found: %v", err.Error()))
		// log.Warningf("Flux is not found: %v", err.Error())

		product := fluxinstall.NewProduct(flags.FluxVersion)

		installer := fluxinstall.NewInstaller()

		execPath, err := installer.Ensure(ctx, product)
		if err != nil {
			execPath, err = installer.Install(ctx, product)
			if err != nil {
				return m.logErr(err, true)
				// return err
			}
		}

		wd, err := os.Getwd()
		if err != nil {
			return m.logErr(err, true)
			// return err
		}

		flux, err := fluxexec.NewFlux(wd, execPath)
		if err != nil {
			return m.logErr(err, true)
			// return err
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
			return m.logErr(err, true)
			// return err
		}
	} else {
		m.logMsg(fmt.Sprintf("Flux version %s is found", m.fluxVersion))
		// log.Successf("Flux version %s is found", m.fluxVersion)
	}

	m.logMsg("Checking if GitOps Dashboard is already installed ...")
	// log.Actionf("Checking if GitOps Dashboard is already installed ...")

	dashboardInstalled := install.IsDashboardInstalled(log, ctx, kubeClient, dashboardName, flags.Namespace)

	if dashboardInstalled {
		m.logMsg("GitOps Dashboard is found")
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
				return m.logErr(err, true)
				// return err
			}

			secret, err := install.GenerateSecret(log, password)
			if err != nil {
				return m.logErr(err, true)
				// return err
			}

			man, err := install.NewManager(log, ctx, kubeClient, m.KubeConfigArgs)
			if err != nil {
				m.logMsg("Error creating resource manager")
				// log.Failuref("Error creating resource manager")
				return m.logErr(err, true)
				// return err
			}

			manifests, err := install.CreateDashboardObjects(log, dashboardName, flags.Namespace, adminUsername, secret, HelmChartVersion)
			if err != nil {
				return m.logErr(fmt.Errorf("error creating dashboard objects: %w", err), true)
				// return fmt.Errorf("error creating dashboard objects: %w", err)
			}

			err = install.InstallDashboard(log, ctx, man, manifests)
			if err != nil {
				return m.logErr(fmt.Errorf("gitops dashboard installation failed: %w", err), true)
				// return fmt.Errorf("gitops dashboard installation failed: %w", err)
			} else {
				dashboardInstalled = true
				m.logMsg("GitOps Dashboard has been installed")
				// log.Successf("GitOps Dashboard has been installed")
			}
		}
	}

	if dashboardInstalled {
		m.logMsg(fmt.Sprintf("Request reconciliation of dashboard (timeout %v) ...", flags.Timeout))
		// log.Actionf("Request reconciliation of dashboard (timeout %v) ...", flags.Timeout)

		if err := install.ReconcileDashboard(kubeClient, dashboardName, flags.Namespace, dashboardPodName, flags.Timeout); err != nil {
			m.logMsg(fmt.Sprintf("Error requesting reconciliation of dashboard: %v", err.Error()))
			// log.Failuref("Error requesting reconciliation of dashboard: %v", err.Error())
		} else {
			m.logMsg("Dashboard reconciliation is done.")
			// log.Successf("Dashboard reconciliation is done.")
		}
	}

	cancelDevBucketPortForwarding, err = watch.InstallDevBucketServer(log, kubeClient, cfg)
	if err != nil {
		return m.logErr(err, true)
		// return err
	}

	m.logAdditionalMsg("Podinfo: http://localhost:9000")

	cancelDashboardPortForwarding = nil

	if dashboardInstalled {
		cancelDashboardPortForwarding, err = watch.EnablePortForwardingForDashboard(log, kubeClient, cfg, flags.Namespace, dashboardPodName, flags.DashboardPort)
		if err != nil {
			return m.logErr(err, true)
			// return err
		}

		m.logAdditionalMsg("Dashboard: http://localhost:9001")
	}

	if err := watch.InitializeTargetDir(targetPath); err != nil {
		return m.logErr(fmt.Errorf("couldn't set up against target %s: %w", targetPath, err), true)
		// return fmt.Errorf("couldn't set up against target %s: %w", targetPath, err)
	}

	if err := watch.SetupBucketSourceAndKS(log, kubeClient, flags.Namespace, relativePath, flags.Timeout); err != nil {
		return m.logErr(err, true)
		// return err
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
		return m.logErr(err, true)
		// return err
	}

	// watch for file changes in dir gitRepoRoot
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return m.logErr(err, true)
		// return err
	}

	err = filepath.Walk(rootDir, watch.WatchDirsForFileWalker(watcher, ignorer))
	if err != nil {
		return m.logErr(err, true)
		// return err
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
					m.logMsg(fmt.Sprintf("Error: %v", err))
					// log.Failuref("Error: %v", err)
				}
			}
		}
	}()

	// event aggregation loop
	ticker = time.NewTicker(680 * time.Millisecond)

	go func() {
		for { // nolint:gosimple
			select {
			case <-ticker.C:
				if counter > 0 {
					m.logMsg(fmt.Sprintf("%d change events detected", counter))
					// log.Actionf("%d change events detected", counter)

					// reset counter
					atomic.StoreUint64(&counter, 0)

					if err := watch.SyncDir(log, rootDir, "dev-bucket", minioClient, ignorer); err != nil {
						m.logMsg(fmt.Sprintf("Error syncing dir: %v", err))
						// log.Failuref("Error syncing dir: %v", err)
					}

					if needToRescan {
						// close the old watcher
						if err := watcher.Close(); err != nil {
							m.logMsg(fmt.Sprintf("Error closing the old watcher: %v", err))
							// log.Warningf("Error closing the old watcher: %v", err)
						}
						// create a new watcher
						watcher, err = fsnotify.NewWatcher()
						if err != nil {
							m.logMsg(fmt.Sprintf("Error creating new watcher: %v", err))
							// log.Failuref("Error creating new watcher: %v", err)
						}

						err = filepath.Walk(rootDir, watch.WatchDirsForFileWalker(watcher, ignorer))
						if err != nil {
							m.logMsg(fmt.Sprintf("Error re-walking dir: %v", err))
							// log.Failuref("Error re-walking dir: %v", err)
						}

						needToRescan = false
					}

					m.logMsg(fmt.Sprintf("Request reconciliation of dev-bucket, and dev-ks (timeout %v) ... ", flags.Timeout))
					// log.Actionf("Request reconciliation of dev-bucket, and dev-ks (timeout %v) ... ", flags.Timeout)

					if err := watch.ReconcileDevBucketSourceAndKS(log, kubeClient, flags.Namespace, flags.Timeout); err != nil {
						m.logMsg(fmt.Sprintf("Error requesting reconciliation: %v", err))
						// log.Failuref("Error requesting reconciliation: %v", err)
					}

					m.logMsg("Reconciliation is done.")
					// log.Successf("Reconciliation is done.")

					if flags.PortForward != "" {
						specMap, err := watch.ParsePortForwardSpec(flags.PortForward)
						if err != nil {
							m.logMsg(fmt.Sprintf("Error parsing port forward spec: %v", err))
							// log.Failuref("Error parsing port forward spec: %v", err)
						}

						// get pod from specMap
						namespacedName := types.NamespacedName{Namespace: specMap.Namespace, Name: specMap.Name}

						pod, err := run.GetPodFromResourceDescription(namespacedName, specMap.Kind, kubeClient)
						if err != nil {
							m.logMsg(fmt.Sprintf("Error getting pod from specMap: %v", err))
							// log.Failuref("Error getting pod from specMap: %v", err)
						}

						if pod != nil {
							waitFwd := make(chan struct{}, 1)
							readyChannel := make(chan struct{})
							cancelPortFwd = func() {
								close(waitFwd)

								cancelPortFwd = nil
							}

							m.logMsg(fmt.Sprintf("Port forwarding to pod %s/%s ...", pod.Namespace, pod.Name))
							// log.Actionf("Port forwarding to pod %s/%s ...", pod.Namespace, pod.Name)

							// this function _BLOCKS_ until the stopChannel (waitPwd) is closed.
							if err := watch.ForwardPort(pod, cfg, specMap, waitFwd, readyChannel); err != nil {
								m.logMsg(fmt.Sprintf("Error forwarding port: %v", err))
								// log.Failuref("Error forwarding port: %v", err)
							}

							m.logMsg("Port forwarding is stopped.")
							// log.Successf("Port forwarding is stopped.")
						}
					}
				}
			}
		}
	}()

	// wait for interrupt or ctrl+C
	m.logMsg("Press Ctrl+C to stop GitOps Run ...")
	// log.Waitingf("Press Ctrl+C to stop GitOps Run ...")

	return runGitopsRunMsg{msg: "finished running gitops run"}

	// return func() tea.Msg {
	// 	return runGitopsRunMsg{msg: "finished running gitops run"}
	// }
}

func (m UIModel) StopAndCleanUpGitopsRun(fluxVersion string) tea.Msg {
	flags := m.Flags

	// sigs := make(chan os.Signal, 1)
	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// sig := <-sigs

	if err := watcher.Close(); err != nil {
		m.logMsg(fmt.Sprintf("Error closing watcher: %v", err.Error()))
		// log.Warningf("Error closing watcher: %v", err.Error())
	}

	// print a blank line to make it easier to read the logs
	// fmt.Println()
	cancelDevBucketPortForwarding()

	if cancelDashboardPortForwarding != nil {
		cancelDashboardPortForwarding()
	}

	ticker.Stop()

	if err := watch.CleanupBucketSourceAndKS(log, kubeClient, flags.Namespace); err != nil {
		return m.logErr(err, true)
		// return err
	}

	// uninstall dev-bucket server
	if err := watch.UninstallDevBucketServer(log, kubeClient); err != nil {
		return m.logErr(err, true)
		// return err
	}

	// run bootstrap wizard only if Flux is not installed and env var is set
	// if fluxVersion != "" || os.Getenv("GITOPS_RUN_BOOTSTRAP") == "" {
	// 	return nil
	// }

	// re-enable listening for ctrl+C
	// signal.Reset(sig)

	// parse remote
	// repo, err := bootstrap.ParseGitRemote(log, rootDir)
	// if err != nil {
	// 	m.logMsg(fmt.Sprintf("Error parsing Git remote: %v", err.Error()))
	// 	// log.Failuref("Error parsing Git remote: %v", err.Error())
	// }

	// run the bootstrap wizard
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
	// 	return m.logErr(fmt.Errorf("error creating bootstrap wizard: %v", err.Error()), true)
	// 	// return fmt.Errorf("error creating bootstrap wizard: %v", err.Error())
	// }

	// if err = wizard.Run(log); err != nil {
	// 	return m.logErr(fmt.Errorf("error running bootstrap wizard: %v", err.Error()), true)
	// 	// return fmt.Errorf("error running bootstrap wizard: %v", err.Error())
	// }

	// _ = wizard.BuildCmd(log)

	// log.Successf("Flux bootstrap command successfully built.")

	// return nil
	// }()

	return tea.Quit()
}
