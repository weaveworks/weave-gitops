package controller

import (
	"context"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-logr/logr"
	"github.com/helm/helm/pkg/chartutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/weaveworks/weave-gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache"
)

const (
	watcherFinalizer = "finalizers.helm.watcher"
)

// HelmWatcherReconciler runs the `reconcile` loop for the watcher.
type HelmWatcherReconciler struct {
	client.Client
	Cache       cache.Cache
	RepoManager helm.HelmRepoManager
}

// +kubebuilder:rbac:groups=helm.watcher,resources=helmrepositories,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=helm.watcher,resources=helmrepositories/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=helm.watcher,resources=helmrepositories/finalizers,verbs=get;create;update;patch;delete

// Reconcile is either called when there is a new HelmRepository or, when there is an update to a HelmRepository.
// Because the watcher watches all helmrepositories, it will update data for all of them.
func (r *HelmWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("repository", req.NamespacedName)

	// get source object
	var repository sourcev1.HelmRepository
	if err := r.Get(ctx, req.NamespacedName, &repository); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Add our finalizer if it does not exist
	if !controllerutil.ContainsFinalizer(&repository, watcherFinalizer) {
		patch := client.MergeFrom(repository.DeepCopy())
		controllerutil.AddFinalizer(&repository, watcherFinalizer)

		if err := r.Patch(ctx, &repository, patch); err != nil {
			log.Error(err, "unable to register finalizer")
			return ctrl.Result{}, err
		}
	}

	// Examine if the object is under deletion
	if !repository.ObjectMeta.GetDeletionTimestamp().IsZero() {
		return r.reconcileDelete(ctx, repository)
	}

	if repository.Status.Artifact == nil {
		// This should not occur because the predicate already checks for artifact's existence, but we do this as a
		// precaution in case that was circumvented.
		return ctrl.Result{}, nil
	}

	log.Info("found the repository: ", "name", repository.Name)
	// Reconcile is called for two reasons. One, the repository was just created, two there is a new revision.
	// Because of that, we don't care what's in the cache. We will always fetch and set it.

	charts, err := r.RepoManager.ListCharts(context.Background(), &repository, helm.Profiles)
	if err != nil {
		return ctrl.Result{}, err
	}

	values := make(cache.ValueMap)

	for _, chart := range charts {
		for _, v := range chart.AvailableVersions {
			// what happens when there are no values? We should just skip that version...
			valueBytes, err := r.RepoManager.GetValuesFile(context.Background(), &repository, &helm.ChartReference{
				Chart:   chart.Name,
				Version: v,
			}, chartutil.ValuesfileName)

			if err != nil {
				log.Error(err, "failed to get values for chart and version, skipping...", "chart", chart.Name, "version", v)
				// log and skip version
				continue
			}

			if _, ok := values[chart.Name]; !ok {
				values[chart.Name] = make(map[string][]byte)
			}

			values[chart.Name][v] = valueBytes
		}
	}

	data := cache.Data{
		Profiles: charts,
		Values:   values,
	}

	if err := r.Cache.Put(logr.NewContext(ctx, log), repository.Namespace, repository.Name, data); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("cached data from repository", "url", repository.Status.URL, "name", repository.Name, "number of profiles", len(charts))

	return ctrl.Result{}, nil
}

func (r *HelmWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sourcev1.HelmRepository{}).
		WithEventFilter(predicate.Or(ArtifactUpdatePredicate{}, DeletePredicate{})).
		Complete(r)
}

func (r *HelmWatcherReconciler) reconcileDelete(ctx context.Context, repository sourcev1.HelmRepository) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx)

	log.Info("deleting repository cache", "namespace", repository.Namespace, "name", repository.Name)

	if err := r.Cache.Delete(ctx, repository.Namespace, repository.Name); err != nil {
		log.Error(err, "failed to remove cache for repository", "namespace", repository.Namespace, "name", repository.Name)
		return ctrl.Result{}, err
	}

	log.Info("deleted repository cache", "namespace", repository.Namespace, "name", repository.Name)
	// Remove our finalizer from the list and update it
	controllerutil.RemoveFinalizer(&repository, watcherFinalizer)

	if err := r.Update(ctx, &repository); err != nil {
		log.Error(err, "failed to update repository to remove the finalizer", "namespace", repository.Namespace, "name", repository.Name)
		return ctrl.Result{}, err
	}

	log.Info("removed finalizer from repository", "namespace", repository.Namespace, "name", repository.Name)
	// Stop reconciliation as the object is being deleted
	return ctrl.Result{}, nil
}
