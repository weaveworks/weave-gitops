package controller

import (
	"context"
	"sort"

	"github.com/Masterminds/semver/v3"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-logr/logr"
	"github.com/helm/helm/pkg/chartutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache"
)

const (
	watcherFinalizer = "finalizers.helm.watcher"
)

// EventRecorder defines an external event recorder's function for creating events for the notification controller.
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . eventRecorder
type eventRecorder interface {
	EventInfof(object corev1.ObjectReference, metadata map[string]string, reason string, messageFmt string, args ...interface{}) error
}

// HelmWatcherReconciler runs the `reconcile` loop for the watcher.
type HelmWatcherReconciler struct {
	client.Client
	Cache                 cache.Cache
	RepoManager           helm.HelmRepoManager
	ExternalEventRecorder eventRecorder
	Scheme                *runtime.Scheme
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
		if v, err := r.checkForNewVersion(ctx, chart); err != nil {
			log.Error(err, "checking for new versions failed")
		} else if v != "" {
			log.Info("sending notification event for new version", "version", v)
			r.sendEvent(log, &repository, "info", chart.Name, v)
		}

		for _, v := range chart.AvailableVersions {
			valueBytes, err := r.RepoManager.GetValuesFile(context.Background(), &repository, &helm.ChartReference{
				Chart:   chart.Name,
				Version: v,
			}, chartutil.ValuesfileName)

			if err != nil {
				log.Error(err, "failed to get values for chart and version, skipping...", "chart", chart.Name, "version", v)
				// log error and skip version
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

// sendEvent emits an event and forwards it to the notification controller if configured.
func (r *HelmWatcherReconciler) sendEvent(log logr.Logger, hr *sourcev1.HelmRepository, severity, profileName, version string) {
	if r.ExternalEventRecorder == nil {
		return
	}

	objRef, err := reference.GetReference(r.Scheme, hr)
	if err != nil {
		log.Error(err, "unable to get reference")
		return
	}

	var meta map[string]string
	if hr.Status.Artifact.Revision != "" {
		meta = map[string]string{"revision": hr.Status.Artifact.Revision}
	}

	if err := r.ExternalEventRecorder.EventInfof(*objRef, meta, severity, "New version available for profile %s with version %s", profileName, version); err != nil {
		log.Error(err, "unable to send event")
		return
	}
}

// checkForNewVersion uses existing data to determine if there are newer versions in the incoming data
// compared to what's already stored in the cache. It returns the LATEST version which is greater than
// the last version that was stored.
func (r *HelmWatcherReconciler) checkForNewVersion(ctx context.Context, chart *pb.Profile) (string, error) {
	versions, err := r.Cache.ListAvailableVersionsForProfile(ctx, chart.GetHelmRepository().GetNamespace(), chart.GetHelmRepository().GetName(), chart.Name)
	if err != nil {
		return "", err
	}

	newVersions, err := ConvertStringListToSemanticVersionList(chart.AvailableVersions)
	if err != nil {
		return "", err
	}

	oldVersions, err := ConvertStringListToSemanticVersionList(versions)
	if err != nil {
		return "", err
	}

	SortVersions(newVersions)
	SortVersions(oldVersions)

	// If there are no old versions stored, it's likely that the profile didn't exist before. So we don't notify.
	// Same in case there are no new versions ( which is unlikely to happen, but we ward against it nevertheless ).
	if len(oldVersions) == 0 || len(newVersions) == 0 {
		return "", nil
	}

	// Notify in case the latest new version is greater than the latest old version.
	if newVersions[0].GreaterThan(oldVersions[0]) {
		return newVersions[0].String(), nil
	}

	return "", nil
}

// ConvertStringListToSemanticVersionList converts a slice of strings into a slice of semantic version.
func ConvertStringListToSemanticVersionList(versions []string) ([]*semver.Version, error) {
	var result []*semver.Version

	for _, v := range versions {
		ver, err := semver.NewVersion(v)
		if err != nil {
			return nil, err
		}

		result = append(result, ver)
	}

	return result, nil
}

// SortVersions sorts semver versions in decreasing order.
func SortVersions(versions []*semver.Version) {
	sort.SliceStable(versions, func(i, j int) bool {
		return versions[i].GreaterThan(versions[j])
	})
}
