package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
)

// HelmWatcherReconcilerPredicate triggers an update event
// when a HelmRepository revision changes.
type HelmWatcherReconcilerPredicate struct {
	predicate.Funcs
}

func (HelmWatcherReconcilerPredicate) Delete(e event.DeleteEvent) bool {
	src, ok := e.Object.(sourcev1.Source)
	// If it doesn't have an artifact it didn't have a cache to begin with.
	if !ok || src.GetArtifact() == nil {
		return false
	}

	return true
}

func (HelmWatcherReconcilerPredicate) Create(e event.CreateEvent) bool {
	src, ok := e.Object.(sourcev1.Source)
	// GetArtifact here will only be populated once the HelmRepository has been updated with a status.
	if !ok || src.GetArtifact() == nil {
		return false
	}

	return true
}

func (HelmWatcherReconcilerPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}

	oldSource, ok := e.ObjectOld.(sourcev1.Source)
	if !ok {
		return false
	}

	newSource, ok := e.ObjectNew.(sourcev1.Source)
	if !ok {
		return false
	}

	if oldSource.GetArtifact() == nil && newSource.GetArtifact() != nil {
		return true
	}

	// There is no way that the old artifact is newer here. We just care that they are of a different revision.
	// Kubernetes takes care of setting old and new accordingly.
	if oldSource.GetArtifact() != nil && newSource.GetArtifact() != nil &&
		oldSource.GetArtifact().Revision != newSource.GetArtifact().Revision {
		return true
	}

	return false
}
