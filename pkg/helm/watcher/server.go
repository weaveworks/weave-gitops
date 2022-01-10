package watcher

import (
	"io/ioutil"
	"os"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/weaveworks/weave-gitops/pkg/helm"
	//+kubebuilder:scaffold:imports
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/controller"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

type Watcher struct {
	cache       cache.Cache
	repoManager *helm.RepoManager
	kubeClient  client.Client
}

func NewWatcher(kubeClient client.Client, cache cache.Cache) (*Watcher, error) {
	tempDir, err := ioutil.TempDir("", "helmrepocache")
	if err != nil {
		return nil, err
	}

	return &Watcher{
		cache:       cache,
		repoManager: helm.NewRepoManager(kubeClient, tempDir),
	}, nil
}

func (w *Watcher) StartWatcher() error {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(sourcev1.AddToScheme(scheme))
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Development: true,
	})))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     ":9980",
		Port:                   9443,
		HealthProbeBindAddress: ":9981",
		LeaderElection:         false,
		LeaderElectionID:       "25a858a4.helm.watcher",
		Logger:                 ctrl.Log,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return err
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	if err = (&controller.HelmWatcherReconciler{
		Client:      mgr.GetClient(),
		Scheme:      mgr.GetScheme(),
		Cache:       w.cache,
		RepoManager: w.repoManager,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "HelmWatcherReconciler")
		return err
	}

	setupLog.Info("starting manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}

	return nil
}
