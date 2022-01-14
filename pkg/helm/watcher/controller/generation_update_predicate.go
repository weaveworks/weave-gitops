package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// GenerationUpdatePredicate triggers an update event when a HelmRepository generation changes.
// i.e.: Delete events.
type GenerationUpdatePredicate struct {
	predicate.Funcs
}

func (GenerationUpdatePredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}

	return e.ObjectOld.GetGeneration() < e.ObjectNew.GetGeneration()
}
