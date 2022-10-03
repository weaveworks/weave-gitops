package clustersmngr

import (
	"context"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type delegatingCache struct {
	cache.Cache

	Client client.Reader

	scheme        *runtime.Scheme
	checkedGVKs   map[schema.GroupVersionKind]struct{}
	checkedGVKsMu *sync.Mutex
}

func NewDelegatingCache(cr client.Reader, cache cache.Cache, scheme *runtime.Scheme) cache.Cache {
	dc := delegatingCache{
		Cache:         cache,
		Client:        cr,
		scheme:        scheme,
		checkedGVKs:   map[schema.GroupVersionKind]struct{}{},
		checkedGVKsMu: &sync.Mutex{},
	}

	return dc
}

func (dc delegatingCache) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	gvk, err := apiutil.GVKForObject(obj, dc.scheme)
	if err != nil {
		return err
	}

	bypass := dc.shouldBypassCheck(obj, gvk)
	if !bypass {
		partial := &metav1.PartialObjectMetadata{}
		partial.SetGroupVersionKind(gvk)

		if err := dc.Client.Get(ctx, key, partial); err != nil {
			return err
		}

		dc.markGVKChecked(gvk)
	}

	return dc.Cache.Get(ctx, key, obj)
}

func (dc delegatingCache) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	gvk, err := apiutil.GVKForObject(list, dc.scheme)
	if err != nil {
		return err
	}

	bypass := dc.shouldBypassCheck(list, gvk)
	if !bypass {
		partial := &metav1.PartialObjectMetadataList{}
		partial.SetGroupVersionKind(gvk)

		if err := dc.Client.List(ctx, partial, opts...); err != nil {
			return err
		}

		dc.markGVKChecked(gvk)
	}

	return dc.Cache.List(ctx, list, opts...)
}

func (dc *delegatingCache) shouldBypassCheck(obj runtime.Object, gvk schema.GroupVersionKind) bool {
	dc.checkedGVKsMu.Lock()
	defer dc.checkedGVKsMu.Unlock()

	if _, isCheckedGVK := dc.checkedGVKs[gvk]; isCheckedGVK {
		return true
	}

	return false
}

func (dc *delegatingCache) markGVKChecked(gvk schema.GroupVersionKind) {
	dc.checkedGVKsMu.Lock()
	defer dc.checkedGVKsMu.Unlock()
	dc.checkedGVKs[gvk] = struct{}{}
}
