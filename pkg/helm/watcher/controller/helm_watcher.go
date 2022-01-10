package controller

import (
	"context"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache"
)

// HelmWatcherReconciler runs the reconcile loop for the watcher.
type HelmWatcherReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Cache       cache.Cache       // this would be an interface ofc.
	RepoManager *helm.RepoManager // TODO: change this to an Interface.
}

// +kubebuilder:rbac:groups=helm.watcher,resources=helmrepositories,verbs=get;list;watch
// +kubebuilder:rbac:groups=helm.watcher,resources=helmrepositories/status,verbs=get

// Reconcile is either called when there is a new HelmRepository or, when there is an update to a HelmRepository.
// Because the watcher watches all helmrepositories, it will update data for all of them.
func (r *HelmWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContext(ctx)

	// get source object
	var repository sourcev1.HelmRepository
	if err := r.Get(ctx, req.NamespacedName, &repository); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if repository.Status.Artifact == nil {
		// Repository was just created, no status field yet. Requeue to the update cycle where status is set.
		return ctrl.Result{}, nil
	}
	log.Info("found the repository: ", "name", repository.Name)
	// Reconcile was called for two reasons. One, the repository was just created, two there is a new revision.
	// Because of that, we don't care what's in the cache. We will always fetch and update.
	//repository.Status.Artifact.

	log.Info("cached data from repository", "url", repository.Status.URL, "name", repository.Name)

	return ctrl.Result{}, nil
}

func (r *HelmWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sourcev1.HelmRepository{}, builder.WithPredicates(HelmWatcherReconcilerPredicate{})).
		Complete(r)
}
