package watcher

import (
	"io/ioutil"

	"github.com/fluxcd/pkg/runtime/events"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"github.com/weaveworks/weave-gitops/for-enterprise/pkg/helm"

	//+kubebuilder:scaffold:imports
	"github.com/weaveworks/weave-gitops/for-enterprise/pkg/helm/watcher/cache"
	"github.com/weaveworks/weave-gitops/for-enterprise/pkg/helm/watcher/controller"
)

const controllerName = "helm-watcher"

var (
	scheme = runtime.NewScheme()
)

type Options struct {
	KubeClient                    client.Client
	Cache                         cache.Cache
	MetricsBindAddress            string
	HealthzBindAddress            string
	NotificationControllerAddress string
	WatcherPort                   int
}

type Watcher struct {
	cache               cache.Cache
	repoManager         helm.HelmRepoManager
	metricsBindAddress  string
	healthzBindAddress  string
	watcherPort         int
	notificationAddress string
}

func NewWatcher(opts Options) (*Watcher, error) {
	tempDir, err := ioutil.TempDir("", "profile_cache")
	if err != nil {
		return nil, err
	}

	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := sourcev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return &Watcher{
		cache:               opts.Cache,
		repoManager:         helm.NewRepoManager(opts.KubeClient, tempDir),
		healthzBindAddress:  opts.HealthzBindAddress,
		metricsBindAddress:  opts.MetricsBindAddress,
		notificationAddress: opts.NotificationControllerAddress,
		watcherPort:         opts.WatcherPort,
	}, nil
}

func (w *Watcher) StartWatcher(log logr.Logger) error {
	ctrl.SetLogger(log.WithName("helm-watcher"))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     w.metricsBindAddress,
		HealthProbeBindAddress: w.healthzBindAddress,
		Port:                   w.watcherPort,
		Logger:                 ctrl.Log,
	})
	if err != nil {
		ctrl.Log.Error(err, "unable to create manager")
		return err
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		ctrl.Log.Error(err, "unable to set up health check")
		return err
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		ctrl.Log.Error(err, "unable to set up ready check")
		return err
	}

	var eventRecorder *events.Recorder

	if w.notificationAddress != "" {
		var err error
		if eventRecorder, err = events.NewRecorder(mgr, ctrl.Log, w.notificationAddress, controllerName); err != nil {
			ctrl.Log.Error(err, "unable to create event recorder")
			return err
		}
	}

	if err = (&controller.HelmWatcherReconciler{
		Client:                mgr.GetClient(),
		Cache:                 w.cache,
		RepoManager:           w.repoManager,
		Scheme:                scheme,
		ExternalEventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create controller", "controller", "HelmWatcherReconciler")
		return err
	}

	ctrl.Log.Info("starting manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		ctrl.Log.Error(err, "problem running manager")
		return err
	}

	return nil
}
