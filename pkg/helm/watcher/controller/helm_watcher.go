package controller

import (
	"context"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache"
)

// HelmWatcherReconciler runs the reconcile loop for the watcher.
type HelmWatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cache  cache.Cache // this would be an interface ofc.
}

// +kubebuilder:rbac:groups=helm.watcher,resources=helmrepositories,verbs=get;list;watch
// +kubebuilder:rbac:groups=helm.watcher,resources=helmrepositories/status,verbs=get

func (r *HelmWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContext(ctx)

	// get source object
	var repository sourcev1.HelmRepository
	if err := r.Get(ctx, req.NamespacedName, &repository); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// TODO: Call GetChart and cache the resulting profiles data.
	log.Info("found the repository: ", "name", repository.Name)
	//if url := r.Cache.Get(repository.Name); url != "" {
	//	log.Info("already seen this one", "url", url)
	//return ctrl.Result{}, nil
	//}
	//r.Cache.Add(repository.Name, repository.Status.URL)
	log.Info("added this new one", "url", repository.Status.URL)

	return ctrl.Result{}, nil
}

func (r *HelmWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sourcev1.HelmRepository{}, builder.WithPredicates(HelmWatcherReconcilerPredicate{})).
		Complete(r)
}
