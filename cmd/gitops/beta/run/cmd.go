package run

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/fsnotify/fsnotify"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/fluxexec"
	"github.com/weaveworks/weave-gitops/pkg/fluxinstall"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run"
	"github.com/weaveworks/weave-gitops/pkg/run/bootstrap"
	"github.com/weaveworks/weave-gitops/pkg/run/install"
	"github.com/weaveworks/weave-gitops/pkg/run/watch"
	"github.com/weaveworks/weave-gitops/pkg/s3"
	"github.com/weaveworks/weave-gitops/pkg/validate"
	"github.com/weaveworks/weave-gitops/pkg/version"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

const (
	dashboardName    = "ww-gitops"
	dashboardPodName = "ww-gitops-weave-gitops"
	adminUsername    = "admin"
)

var HelmChartVersion = "3.0.0"

type RunCommandFlags struct {
	FluxVersion       string
	AllowK8sContext   []string
	Components        []string
	ComponentsExtra   []string
	Timeout           time.Duration
	PortForward       string // port forward specifier, e.g. "port=8080:8080,resource=svc/app"
	RootDir           string
	DecryptionKeyFile string

	// Dashboard
	DashboardPort           string
	DashboardHashedPassword string
	SkipDashboardInstall    bool
	DashboardImage          string

	// Session
	SessionName         string
	SessionNamespace    string
	NoSession           bool
	SkipResourceCleanup bool
	NoBootstrap         bool

	// Global flags.
	Namespace  string
	KubeConfig string

	// Flags, created by genericclioptions.
	Context string

	// Hidden session name for the sub-process
	HiddenSessionName string
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
gitops beta run ./deploy/overlays/dev --no-session --timeout 3m --port-forward namespace=dev,resource=svc/backend,port=9898:9898

# Run the sync on the podinfo demo in the session mode.
git clone https://github.com/stefanprodan/podinfo
cd podinfo
gitops beta run ./deploy/overlays/dev --timeout 3m --port-forward namespace=dev,resource=svc/backend,port=9898:9898

# Run the sync on the podinfo Helm chart, in the session mode. Please note that file Chart.yaml must exist in the directory.
git clone https://github.com/stefanprodan/podinfo
cd podinfo
gitops beta run ./chart/podinfo --timeout 3m --port-forward namespace=flux-system,resource=svc/run-dev-helm-podinfo,port=9898:9898`,
		SilenceUsage:      true,
		SilenceErrors:     true,
		PreRunE:           betaRunCommandPreRunE(&opts.Endpoint),
		RunE:              betaRunCommandRunE(opts),
		DisableAutoGenTag: true,
	}

	cmdFlags := cmd.Flags()

	cmdFlags.StringVar(&flags.FluxVersion, "flux-version", version.FluxVersion, "The version of Flux to install.")
	cmdFlags.StringSliceVar(&flags.AllowK8sContext, "allow-k8s-context", []string{}, "The name of the KubeConfig context to explicitly allow.")
	cmdFlags.StringSliceVar(&flags.Components, "components", []string{"source-controller", "kustomize-controller", "helm-controller", "notification-controller"}, "The Flux components to install.")
	cmdFlags.StringSliceVar(&flags.ComponentsExtra, "components-extra", []string{}, "Additional Flux components to install, allowed values are image-reflector-controller,image-automation-controller.")
	cmdFlags.DurationVar(&flags.Timeout, "timeout", 5*time.Minute, "The timeout for operations during GitOps Run.")
	cmdFlags.StringVar(&flags.PortForward, "port-forward", "", "Forward the port from a cluster's resource to your local machine i.e. 'port=8080:8080,resource=svc/app'.")
	cmdFlags.StringVar(&flags.DashboardPort, "dashboard-port", "9001", "GitOps Dashboard port")
	cmdFlags.BoolVar(&flags.SkipDashboardInstall, "skip-dashboard-install", false, "Skip installation of the Dashboard. This also disables the prompt asking whether the Dashboard should be installed.")
	cmdFlags.StringVar(&flags.DashboardHashedPassword, "dashboard-hashed-password", "", "GitOps Dashboard password in BCrypt hash format")
	cmdFlags.StringVar(&flags.RootDir, "root-dir", "", "Specify the root directory to watch for changes. If not specified, the root of Git repository will be used.")
	cmdFlags.StringVar(&flags.SessionName, "session-name", getSessionNameFromGit(), "Specify the name of the session. If not specified, the name of the current branch and the last commit id will be used.")
	cmdFlags.StringVar(&flags.SessionNamespace, "session-namespace", "default", "Specify the namespace of the session.")
	cmdFlags.BoolVar(&flags.NoSession, "no-session", false, "Disable session management. If not specified, the session will be enabled by default.")
	cmdFlags.BoolVar(&flags.NoBootstrap, "no-bootstrap", false, "Disable bootstrapping at shutdown.")
	cmdFlags.BoolVar(&flags.SkipResourceCleanup, "skip-resource-cleanup", false, "Skip resource cleanup. If not specified, the GitOps Run resources will be deleted by default.")
	cmdFlags.StringVar(&flags.DecryptionKeyFile, "decryption-key-file", "", "Path to an age key file used for decrypting Secrets using SOPS.")

	cmdFlags.StringVar(&flags.DashboardImage, "dashboard-image", "", "Override GitOps Dashboard image")
	_ = cmdFlags.MarkHidden("dashboard-image")

	cmdFlags.StringVar(&flags.HiddenSessionName, "x-session-name", "", "The session name acknowledged by the sub-process. This is a hidden flag and should not be used.")
	_ = cmdFlags.MarkHidden("x-session-name")

	kubeConfigArgs = run.GetKubeConfigArgs()

	kubeConfigArgs.AddFlags(cmd.Flags())

	return cmd
}

func getSessionNameFromGit() string {
	const prefix = "run"

	branch, err := run.GetBranchName()
	if err != nil {
		return ""
	}

	commit, err := run.GetCommitID()
	if err != nil {
		return ""
	}

	isDirty, err := run.IsDirty()
	if err != nil {
		return ""
	}

	sessionName := fmt.Sprintf("%s-%s-%s", prefix, branch, commit)
	if isDirty {
		sessionName = fmt.Sprintf("%s-%s-%s-dirty", prefix, branch, commit)
	}

	sessionName = strings.ToLower(strings.ReplaceAll(sessionName, "/", "-"))

	return sessionName
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

func getKubeClient(cmd *cobra.Command, args []string) (*kube.KubeHTTP, *rest.Config, error) {
	var err error

	log := logger.NewCLILogger(os.Stdout)

	if flags.Namespace, err = cmd.Flags().GetString("namespace"); err != nil {
		return nil, nil, err
	}

	kubeConfigArgs.Namespace = &flags.Namespace

	if flags.KubeConfig, err = cmd.Flags().GetString("kubeconfig"); err != nil {
		return nil, nil, err
	}

	if flags.Context, err = cmd.Flags().GetString("context"); err != nil {
		return nil, nil, err
	}

	if flags.KubeConfig != "" {
		kubeConfigArgs.KubeConfig = &flags.KubeConfig

		if flags.Context == "" {
			log.Failuref("A context should be provided if a kubeconfig is provided")
			return nil, nil, cmderrors.ErrNoContextForKubeConfig
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
			return nil, nil, cmderrors.ErrNoCluster
		}
	}

	cfg, err := kubeConfigArgs.ToRESTConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting a restconfig from kube config args: %w", err)
	}

	kubeClientOpts := run.GetKubeClientOptions()
	kubeClientOpts.BindFlags(cmd.Flags())

	kubeClient, err := run.GetKubeClient(log, contextName, cfg, kubeClientOpts)
	if err != nil {
		return nil, nil, cmderrors.ErrGetKubeClient
	}

	return kubeClient, cfg, nil
}

func fluxStep(log logger.Logger, kubeClient *kube.KubeHTTP) (fluxVersion string, justInstalled bool, err error) {
	ctx := context.Background()

	log.Actionf("Checking if Flux is already installed ...")

	guessed := false
	if fluxVersion, guessed, err = install.GetFluxVersion(ctx, log, kubeClient); err != nil {
		log.Warningf("Flux is not found: %v", err.Error())

		product := fluxinstall.NewProduct(flags.FluxVersion)

		installer := fluxinstall.NewInstaller()

		execPath, err := installer.Ensure(ctx, product)
		if err != nil {
			execPath, err = installer.Install(ctx, product)
			if err != nil {
				return "", false, err
			}
		}

		wd, err := os.Getwd()
		if err != nil {
			return "", false, err
		}

		flux, err := fluxexec.NewFlux(wd, execPath)
		if err != nil {
			return "", false, err
		}

		// This means that Flux logs will be printed to the console, but not be sent to S3
		flux.SetLogger(log.L())

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
			return "", false, err
		}

		fluxVersion = flags.FluxVersion

		return fluxVersion, true, nil
	} else {
		if guessed {
			log.Warningf("Flux version could not be determined, assuming %s by mapping from the version of the Source controller", fluxVersion)
		} else {
			log.Successf("Flux %s is already installed", fluxVersion)
		}
	}

	return fluxVersion, false, nil
}

func dashboardStep(ctx context.Context, log logger.Logger, kubeClient *kube.KubeHTTP, generateManifestsOnly bool, dashboardHashedPassword string) (bool, []byte, string, error) {
	log.Actionf("Checking if GitOps Dashboard is already installed ...")

	dashboardInstalled := install.IsDashboardInstalled(ctx, log, kubeClient, dashboardName, flags.Namespace)

	var dashboardManifests []byte

	if dashboardInstalled {
		log.Successf("GitOps Dashboard is found")
	} else {
		wantToInstallTheDashboard := false
		if dashboardHashedPassword != "" {
			wantToInstallTheDashboard = true
		} else if !flags.SkipDashboardInstall && dashboardHashedPassword == "" {
			prompt := promptui.Prompt{
				Label:     "Would you like to install the GitOps Dashboard",
				IsConfirm: true,
				Default:   "Y",
			}

			// Answering "n" causes err to not be nil. Hitting enter without typing
			// does not return the default.
			_, err := prompt.Run()
			if err == nil {
				wantToInstallTheDashboard = true
			}
		}

		if wantToInstallTheDashboard {
			passwordHash := ""
			if dashboardHashedPassword == "" {
				password, err := install.ReadPassword(log)
				if err != nil {
					return false, nil, "", err
				}

				passwordHash, err = install.GeneratePasswordHash(log, password)
				if err != nil {
					return false, nil, "", err
				}
			} else {
				passwordHash = dashboardHashedPassword
			}

			dashboardManifests, err := install.CreateDashboardObjects(log, dashboardName, flags.Namespace, adminUsername, passwordHash, HelmChartVersion, flags.DashboardImage)
			if err != nil {
				return false, nil, "", fmt.Errorf("error creating dashboard objects: %w", err)
			}

			if generateManifestsOnly {
				return false, dashboardManifests, passwordHash, nil
			}

			man, err := install.NewManager(ctx, log, kubeClient, kubeConfigArgs)
			if err != nil {
				log.Failuref("Error creating resource manager")
				return false, nil, "", err
			}

			err = install.InstallDashboard(ctx, log, man, dashboardManifests)
			if err != nil {
				return false, nil, "", fmt.Errorf("gitops dashboard installation failed: %w", err)
			} else {
				dashboardInstalled = true

				log.Successf("GitOps Dashboard has been installed")
			}
		}
	}

	if dashboardInstalled {
		log.Actionf("Request reconciliation of dashboard (timeout %v) ...", flags.Timeout)

		if err := install.ReconcileDashboard(ctx, kubeClient, dashboardName, flags.Namespace, dashboardPodName, flags.Timeout); err != nil {
			log.Failuref("Error requesting reconciliation of dashboard: %v", err.Error())
		} else {
			log.Successf("Dashboard reconciliation is done.")
		}
	}

	return dashboardInstalled, dashboardManifests, "", nil
}

func runCommandWithSession(cmd *cobra.Command, args []string) (retErr error) {
	paths, err := run.NewPaths(args[0], flags.RootDir)
	if err != nil {
		return err
	}

	kubeClient, _, err := getKubeClient(cmd, args)
	if err != nil {
		return err
	}

	// create session
	sessionLog := logger.NewCLILogger(os.Stdout)
	sessionLog.Actionf("Preparing the cluster for GitOps Run session ...\n")

	sessionLog.Println("You can run `gitops beta run --no-session` to disable session management.\n")

	sessionLog.Println("If you are running GitOps Run for the first time, it may take a few minutes to download the required images.")
	sessionLog.Println("GitOps Run session is also required to install Flux components, if it is not installed yet.")
	sessionLog.Println("You may see Flux installation logs in the next step.\n")

	// showing Flux installation twice is confusing
	log := logger.NewCLILogger(io.Discard)

	var fluxJustInstalled bool

	if _, fluxJustInstalled, err = fluxStep(log, kubeClient); err != nil {
		return fmt.Errorf("failed to install Flux on the host cluster: %v", err)
	}

	_, dashboardManifests, dashboardHashedPassword, err := dashboardStep(context.Background(), log, kubeClient, true, flags.DashboardHashedPassword)
	if err != nil {
		return fmt.Errorf("failed to generate dashboard manifests: %v", err)
	}

	sessionLog.Actionf("Creating GitOps Run session %s in namespace %s ...", flags.SessionName, flags.SessionNamespace)

	sessionLog.Println("\nYou may see Flux installation logs again, as it is being installed inside the session.\n")

	portForwardsForSession := []string{flags.DashboardPort}

	if flags.PortForward != "" {
		spec, err := watch.ParsePortForwardSpec(flags.PortForward)
		if err != nil {
			return err
		}

		portForwardsForSession = append(portForwardsForSession, spec.HostPort)
	}

	var kind string
	if yes, err := isHelm(paths.GetAbsoluteTargetDir()); yes && err == nil {
		kind = "helm"
	} else if !yes && err == nil {
		kind = "ks"
	} else {
		return err
	}

	session, err := install.NewSession(
		sessionLog,
		kubeClient,
		flags.SessionName,
		flags.SessionNamespace,
		portForwardsForSession,
		dashboardHashedPassword,
		kind,
	)

	if err != nil {
		return err
	}

	sessionLog.Actionf("Waiting for GitOps Run session %s to be ready ...", flags.SessionName)

	if err := session.Start(); err != nil {
		return err
	}

	sessionLog.Successf("Session %s is ready", flags.SessionName)

	sessionLog.Actionf("Connecting to GitOps Run session %s ...", flags.SessionName)

	if err := session.Connect(); err != nil {
		return err
	}

	sessionLog.Println("")
	sessionLog.Actionf("Deleting GitOps Run session %s ...", flags.SessionName)

	if err := session.Close(); err != nil {
		sessionLog.Failuref("Failed to delete session %s: %v", flags.SessionName, err)
		return err
	} else {
		sessionLog.Successf("Session %s is deleted successfully", flags.SessionName)
	}

	// now that the session is deleted, we return to the host cluster

	// run bootstrap wizard only if Flux was not installed
	if fluxJustInstalled && !flags.NoBootstrap {
		prompt := promptui.Prompt{
			Label:     "Would you like to bootstrap your cluster into GitOps mode",
			IsConfirm: true,
			Default:   "Y",
		}

		_, err = prompt.Run()
		if err != nil {
			return nil
		}

		for {
			err := runBootstrap(context.Background(), log, paths, dashboardManifests)
			if err == nil {
				break
			}

			log.Warningf("Error bootstrapping: %v", err)

			prompt := promptui.Prompt{
				Label:     "Couldn't bootstrap - would you like to try again",
				IsConfirm: true,
				Default:   "Y",
			}

			_, err = prompt.Run()
			if err != nil {
				return nil
			}
		}
	}

	return err
}

func runCommandWithoutSession(cmd *cobra.Command, args []string) error {
	// There are two loggers in this function.
	// 1. log0 is the os.Stdout logger
	// 2. log is the S3 logger that also delegates its outputs to "log0".
	log0 := logger.NewCLILogger(os.Stdout)

	paths, err := run.NewPaths(args[0], flags.RootDir)
	if err != nil {
		return err
	}

	kubeClient, cfg, err := getKubeClient(cmd, args)
	if err != nil {
		return err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}

	serverVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		return err
	}

	// We need the server version to pass to the validation package
	kubernetesVersion := trimK8sVersion(serverVersion.GitVersion)

	contextName := kubeClient.ClusterName
	validAllowedContext := false

	for _, allowContext := range flags.AllowK8sContext {
		if allowContext == contextName {
			log0.Actionf("Explicitly allow GitOps Run on %s context", contextName)

			validAllowedContext = true

			break
		}
	}

	if !validAllowedContext {
		if !run.IsLocalCluster(kubeClient) {
			return fmt.Errorf("to run against a remote cluster, use --allow-k8s-context=%s", contextName)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	var (
		fluxJustInstalled bool
		fluxVersion       string
	)

	fluxVersion, fluxJustInstalled, err = fluxStep(log0, kubeClient)

	if err != nil {
		cancel()
		return err
	}

	sessionName := flags.HiddenSessionName
	if sessionName == "" {
		sessionName = "no-session"
	}

	var username string
	if current, err := user.Current(); err != nil {
		username = "unknown"
	} else {
		username = current.Username
	}

	// ====================== Dev-bucket ======================
	// Install dev-bucket server before everything, so that we can also forward logs to it
	unusedPorts, err := run.GetUnusedPorts(2)
	if err != nil {
		cancel()
		return err
	}

	devBucketHTTPPort := unusedPorts[0]
	devBucketHTTPSPort := unusedPorts[1]

	// generate access key and secret key for Minio auth
	accessKey, err := s3.GenerateAccessKey(s3.DefaultRandIntFunc)
	if err != nil {
		cancel()
		return fmt.Errorf("failed generating access key: %w", err)
	}

	secretKey, err := s3.GenerateSecretKey(s3.DefaultRandIntFunc)
	if err != nil {
		cancel()
		return fmt.Errorf("failed generating secret key: %w", err)
	}

	cancelDevBucketPortForwarding, cert, err := watch.InstallDevBucketServer(ctx, log0, kubeClient, cfg, devBucketHTTPPort, devBucketHTTPSPort, accessKey, secretKey)
	if err != nil {
		cancel()
		return fmt.Errorf("unable to install S3 bucket server: %w", err)
	}

	minioClient, err := s3.NewMinioClient(fmt.Sprintf("localhost:%d", devBucketHTTPSPort), accessKey, secretKey, cert)
	if err != nil {
		cancel()
		return err
	}

	if err := logger.CreateBucket(minioClient, logger.SessionLogBucketName); err != nil {
		cancel()
		return err
	}

	if err := logger.CreateBucket(minioClient, logger.PodLogBucketName); err != nil {
		cancel()
		return err
	}

	log, err := logger.NewS3LogWriter(minioClient, sessionName, log0)
	if err != nil {
		cancel()
		return fmt.Errorf("failed creating S3 log writer: %w", err)
	}

	// ====================== Dashboard ======================
	var (
		dashboardInstalled bool
		dashboardManifests []byte
	)

	dashboardInstalled, dashboardManifests, _, err = dashboardStep(ctx, log, kubeClient, false, flags.DashboardHashedPassword)
	if err != nil {
		cancel()
		return err
	}

	var cancelDashboardPortForwarding func() = nil

	if dashboardInstalled {
		cancelDashboardPortForwarding, err = watch.EnablePortForwardingForDashboard(ctx, log, kubeClient, cfg, flags.Namespace, dashboardPodName, flags.DashboardPort)
		if err != nil {
			cancel()
			return err
		}
	}

	if err := watch.InitializeTargetDir(paths.GetAbsoluteTargetDir()); err != nil {
		cancel()
		return fmt.Errorf("couldn't set up against target %s: %w", paths.TargetDir, err)
	}

	setupParams := watch.SetupRunObjectParams{
		Namespace:         flags.Namespace,
		Path:              paths.TargetDir,
		Timeout:           flags.Timeout,
		DevBucketPort:     devBucketHTTPPort,
		SessionName:       sessionName,
		Username:          username,
		AccessKey:         accessKey,
		SecretKey:         secretKey,
		DecryptionKeyFile: flags.DecryptionKeyFile,
	}

	if yes, err := isHelm(paths.GetAbsoluteTargetDir()); yes && err == nil {
		if err := watch.SetupBucketSourceAndHelm(ctx, log, kubeClient, setupParams); err != nil {
			cancel()
			return err
		}
	} else if !yes && err == nil {
		if err := watch.SetupBucketSourceAndKS(ctx, log, kubeClient, setupParams); err != nil {
			cancel()
			return err
		}
	} else if err != nil {
		log.Actionf("Unable to determine if target is a Helm or Kustomization directory: %v", err)
		cancel()
		return err
	}

	minioClient, err = s3.NewMinioClient(fmt.Sprintf("localhost:%d", devBucketHTTPSPort), accessKey, secretKey, cert)
	if err != nil {
		cancel()
		return err
	}

	// watch for file changes in dir gitRepoRoot
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		cancel()
		return err
	}

	ignorer := watch.CreateIgnorer(paths.RootDir)

	err = filepath.Walk(paths.RootDir, watch.WatchDirsForFileWalker(watcher, ignorer))
	if err != nil {
		cancel()
		return err
	}

	// cancel function to stop forwarding port
	var (
		cancelPortFwd func()
		// atomic counter for the number of file change events that have changed
		counter      uint64 = 1
		needToRescan        = false
	)

	watcherCtx, watcherCancel := context.WithCancel(ctx)
	lastReconcile := time.Now()
	stopUploadCh := make(chan struct{})

	go func() {
		for {
			select {
			case <-watcherCtx.Done():
				return
			case <-stopUploadCh:
				return
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

				// If there are still changes and it's been a few seconds,
				// cancel the old context and start over.
				if time.Since(lastReconcile) > (10 * time.Second) {
					watcherCancel()
					watcherCtx, watcherCancel = context.WithCancel(ctx)
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
			case <-stopUploadCh:
				return
			case <-ticker.C:
				if counter > 0 {
					log.Actionf("%d change events detected", counter)

					// reset counter
					atomic.StoreUint64(&counter, 0)

					// we have to skip validation for helm charts
					if yes, err := isHelm(paths.GetAbsoluteTargetDir()); !yes && err == nil {
						// validate only files under the target dir
						log.Actionf("Validating files under %s/ ...", paths.TargetDir)

						if err := validate.Validate(paths.GetAbsoluteTargetDir(), kubernetesVersion, fluxVersion); err != nil {
							log.Failuref("Validation failed: please review the errors and try again: %v", err)
							continue
						}
					}

					// use ctx, not thisCtx - incomplete uploads will never make anybody happy
					if err := watch.SyncDir(ctx, log, paths.RootDir, watch.RunDevBucketName, minioClient, ignorer); err != nil {
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

						err = filepath.Walk(paths.RootDir, watch.WatchDirsForFileWalker(watcher, ignorer))
						if err != nil {
							log.Failuref("Error re-walking dir: %v", err)
						}

						needToRescan = false
					}

					log.Actionf("Request reconciliation of GitOps Run resources (timeout %v) ... ", flags.Timeout)

					lastReconcile = time.Now()
					// context that cancels when files change
					thisCtx := watcherCtx

					var reconcileErr error
					if yes, err := isHelm(paths.GetAbsoluteTargetDir()); yes && err == nil {
						reconcileErr = watch.ReconcileDevBucketSourceAndHelm(thisCtx, log, kubeClient, flags.Namespace, flags.Timeout)
					} else if !yes && err == nil {
						reconcileErr = watch.ReconcileDevBucketSourceAndKS(thisCtx, log, kubeClient, flags.Namespace, flags.Timeout)
					} else if err != nil {
						log.Actionf("Unable to determine if target is a Helm or Kustomization directory: %v", err)
						reconcileErr = err
					}

					if reconcileErr != nil {
						log.Failuref("Error requesting reconciliation: %v", reconcileErr)
					} else {
						log.Successf("Reconciliation is done.")
					}

					portForwards := map[rune]watch.PortForwardShortcut{}

					if dashboardInstalled {
						portForwardKey, err := watch.GetNextPortForwardKey(portForwards)
						if err != nil {
							log.Failuref("Error adding a portForward: %v", err)
						} else {
							portForwards[portForwardKey] = watch.PortForwardShortcut{
								Name:     dashboardName,
								HostPort: flags.DashboardPort,
							}
						}
					}

					var specMap *watch.PortForwardSpec

					if flags.PortForward != "" {
						specMap, err = watch.ParsePortForwardSpec(flags.PortForward)
						if err != nil {
							log.Failuref("Error parsing port forward spec: %v", err)
						} else {
							serviceName := specMap.Name
							if serviceName == "" {
								serviceName = "service"
							}

							portForwardKey, err := watch.GetNextPortForwardKey(portForwards)
							if err != nil {
								log.Failuref("Error adding a port forward: %v", err)
							} else {
								portForwards[portForwardKey] = watch.PortForwardShortcut{
									Name:     serviceName,
									HostPort: specMap.HostPort,
								}
							}
						}
					}

					if len(portForwards) > 0 {
						watch.ShowPortForwards(ctx, log, portForwards)
					}

					if specMap != nil {
						// get pod from specMap
						namespacedName := types.NamespacedName{Namespace: specMap.Namespace, Name: specMap.Name}

						var (
							pod    *corev1.Pod
							podErr error
						)

						if pollErr := wait.PollImmediate(2*time.Second, flags.Timeout, func() (bool, error) {
							pod, podErr = run.GetPodFromResourceDescription(thisCtx, namespacedName, specMap.Kind, kubeClient)
							if pod != nil && podErr == nil {
								return true, nil
							}

							log.Waitingf("Waiting for a pod from specMap: %v", podErr)
							return false, nil
						}); pollErr != nil {
							log.Failuref("Waiting for a pod from specMap: %v", pollErr)
						}

						if pod == nil {
							log.Failuref("Error getting pod from specMap")
						} else /* pod is available */ {
							waitFwd := make(chan struct{}, 1)
							readyChannel := make(chan struct{})
							cancelPortFwd = func() {
								close(waitFwd)

								cancelPortFwd = nil
							}

							log.Actionf("Port forwarding to pod %s/%s ...", pod.Namespace, pod.Name)

							// this function _BLOCKS_ until the stopChannel (waitPwd) is closed.
							if err := watch.ForwardPort(log.L(), pod, cfg, specMap, waitFwd, readyChannel); err != nil {
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

	sig := <-sigs

	close(stopUploadCh)
	cancel()
	// create new context that isn't cancelled, for bootstrapping
	ctx = context.Background()

	// re-enable listening for ctrl+C
	signal.Reset(sig)

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

	// this is the default behaviour
	if !flags.SkipResourceCleanup {
		if yes, err := isHelm(paths.GetAbsoluteTargetDir()); yes && err == nil {
			if err := watch.CleanupBucketSourceAndHelm(ctx, log0, kubeClient, flags.Namespace); err != nil {
				return err
			}
		} else if !yes && err == nil {
			if err := watch.CleanupBucketSourceAndKS(ctx, log0, kubeClient, flags.Namespace); err != nil {
				return err
			}
		} else if err != nil {
			log0.Actionf("Unable to determine if target is a Helm or Kustomization directory: %v", err)
			return err
		}

		// uninstall dev-bucket server
		if err := watch.UninstallDevBucketServer(ctx, log0, kubeClient); err != nil {
			return err
		}
	}

	// run bootstrap wizard only if Flux was not installed
	if fluxJustInstalled && !flags.NoBootstrap {
		prompt := promptui.Prompt{
			Label:     "Would you like to bootstrap your cluster into GitOps mode",
			IsConfirm: true,
			Default:   "Y",
		}

		_, err = prompt.Run()
		if err != nil {
			return nil
		}

		for {
			err := runBootstrap(ctx, log0, paths, dashboardManifests)
			if err == nil {
				break
			}

			log0.Warningf("Error bootstrapping: %v", err)

			prompt := promptui.Prompt{
				Label:     "Couldn't bootstrap - would you like to try again",
				IsConfirm: true,
				Default:   "Y",
			}

			_, err = prompt.Run()
			if err != nil {
				return nil
			}
		}
	}

	return nil
}

func isHelm(dir string) (bool, error) {
	_, err := os.Stat(filepath.Join(dir, "Chart.yaml"))
	if err != nil && os.IsNotExist(err) {
		// check Chart.yml
		_, err = os.Stat(filepath.Join(dir, "Chart.yml"))
		if err != nil && os.IsNotExist(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func runBootstrap(ctx context.Context, log logger.Logger, paths *run.Paths, manifests []byte) (err error) {
	// parse remote
	repo, err := bootstrap.ParseGitRemote(log, paths.RootDir)
	if err != nil {
		log.Failuref("Error parsing Git remote: %v", err.Error())
	}

	// run the bootstrap wizard
	log.Actionf("Starting bootstrap wizard ...")

	host := bootstrap.GetHost(repo)
	gitProvider := bootstrap.ParseGitProvider(host)

	log.Waitingf("Press Ctrl+C to stop bootstrap wizard ...")

	if gitProvider == bootstrap.GitProviderUnknown {
		gitProvider, err = bootstrap.SelectGitProvider(log)
		if err != nil {
			log.Failuref("Error selecting git provider: %v", err.Error())
		}
	}

	wizard, err := bootstrap.NewBootstrapWizard(log, gitProvider, repo)

	if err != nil {
		return fmt.Errorf("error creating bootstrap wizard: %v", err.Error())
	}

	if err = wizard.Run(log); err != nil {
		return fmt.Errorf("error running bootstrap wizard: %v", err.Error())
	}

	params := wizard.BuildCmd(log)

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

	flux.SetLogger(log.L())

	slugifiedWorkloadPath := strings.ReplaceAll(paths.TargetDir, "/", "-")

	workloadKustomizationPath := strings.Join([]string{paths.ClusterDir, slugifiedWorkloadPath, slugifiedWorkloadPath + "-kustomization.yaml"}, "/")
	workloadKustomization := kustomizev1.Kustomization{
		TypeMeta: metav1.TypeMeta{
			Kind:       kustomizev1.KustomizationKind,
			APIVersion: kustomizev1.GroupVersion.Identifier(),
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      slugifiedWorkloadPath,
			Namespace: flags.Namespace,
		},
		Spec: kustomizev1.KustomizationSpec{
			Interval: metav1.Duration{Duration: 1 * time.Hour},
			Prune:    true, // GC the kustomization
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind:      sourcev1.GitRepositoryKind,
				Name:      "flux-system",
				Namespace: "flux-system",
			},
			Timeout: &metav1.Duration{Duration: 5 * time.Minute},
			Path:    "./" + paths.TargetDir,
			Wait:    true,
		},
	}

	workloadKustomizationContent, err := yaml.Marshal(workloadKustomization)
	if err != nil {
		return err
	}

	workloadKustomizationContent, err = install.SanitizeResourceData(log, workloadKustomizationContent)
	if err != nil {
		return err
	}

	workloadKustomizationContentStr := string(workloadKustomizationContent)

	commitFiles := []gitprovider.CommitFile{{
		Path:    &workloadKustomizationPath,
		Content: &workloadKustomizationContentStr,
	}}

	if len(manifests) > 0 {
		strManifests := string(manifests)
		dashboardPath := strings.Join([]string{paths.ClusterDir, "weave-gitops", "dashboard.yaml"}, "/")

		commitFiles = append(commitFiles, gitprovider.CommitFile{
			Path:    &dashboardPath,
			Content: &strManifests,
		})
	}

	err = filepath.WalkDir(paths.GetAbsoluteTargetDir(), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			log.Warningf("Error: %v", err.Error())
			return err
		}
		if entry.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		strContent := string(content)
		if err != nil {
			log.Warningf("Error: %v", err.Error())
			return err
		}
		relativePath, err := run.GetRelativePathToRootDir(paths.RootDir, path)
		if err != nil {
			log.Warningf("Error: %v", err.Error())
			return err
		}
		commitFiles = append(commitFiles, gitprovider.CommitFile{
			Path:    &relativePath,
			Content: &strContent,
		})
		return nil
	})
	if err != nil {
		return err
	}

	bs := bootstrap.NewBootstrap(paths.ClusterDir, params.Options, params.Provider)

	err = bs.RunBootstrapCmd(ctx, flux)
	if err != nil {
		return err
	}

	err = bs.SyncResources(ctx, log, commitFiles)
	if err != nil {
		return err
	}

	return nil
}

func betaRunCommandRunE(opts *config.Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if flags.NoSession {
			return runCommandWithoutSession(cmd, args)
		} else {
			return runCommandWithSession(cmd, args)
		}
	}
}
