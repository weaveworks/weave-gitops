package server

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type PrimaryKinds struct {
	kinds map[string]schema.GroupVersionKind
}

func New() *PrimaryKinds {
	kinds := PrimaryKinds{}
	kinds.kinds = make(map[string]schema.GroupVersionKind)

	return &kinds
}

func DefaultPrimaryKinds() *PrimaryKinds {
	kinds := New()
	scheme, _ := kube.CreateScheme()

	for gvk := range scheme.AllKnownTypes() {
		kinds.kinds[gvk.Kind] = gvk
	}

	return kinds
}

// Add sets another kind name and gvk to resolve an object
// This errors if the kind is already set, as this likely indicates 2
// different uses for the same kind string.
func (pk *PrimaryKinds) Add(kind string, gvk schema.GroupVersionKind) error {
	_, ok := pk.kinds[kind]
	if ok {
		return fmt.Errorf("couldn't add kind %v - already added", kind)
	}

	pk.kinds[kind] = gvk

	return nil
}

// Lookup ensures that a kind name is known, white-listed, and returns
// the full GVK for that kind
func (pk *PrimaryKinds) Lookup(kind string) (*schema.GroupVersionKind, error) {
	gvk, ok := pk.kinds[kind]
	if !ok {
		return nil, fmt.Errorf("looking up objects of kind %v not supported", kind)
	}

	return &gvk, nil
}
