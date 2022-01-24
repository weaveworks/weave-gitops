package controller

import (
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// DeletePredicate triggers an update event when a HelmRepository generation changes.
// i.e.: Delete events.
type DeletePredicate struct {
	predicate.Funcs
}

func (DeletePredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}

	src, ok := e.ObjectNew.(*sourcev1.HelmRepository)
	if !ok {
		return false
	}

	return e.ObjectOld.GetGeneration() < e.ObjectNew.GetGeneration() && src.DeletionTimestamp != nil
}
