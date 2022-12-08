package cluster

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type delegatingCacheCluster struct {
	restConfig *rest.Config
	cluster    Cluster
	scheme     *runtime.Scheme
}

func NewDelegatingCacheCluster(cluster Cluster, config *rest.Config, scheme *runtime.Scheme) Cluster {
	return &delegatingCacheCluster{
		restConfig: config,
		cluster:    cluster,
		scheme:     scheme,
	}
}

func (c *delegatingCacheCluster) GetName() string {
	return c.cluster.GetName()
}

func (c *delegatingCacheCluster) GetHost() string {
	return c.cluster.GetHost()
}

func (c *delegatingCacheCluster) makeCachingClient(leafClient client.Client) (client.Client, error) {
	mapper, err := apiutil.NewDiscoveryRESTMapper(c.restConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create RESTMapper from config: %w", err)
	}

	cache, err := cache.New(c.restConfig, cache.Options{
		Scheme: c.scheme,
		Mapper: mapper,
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating client cache: %w", err)
	}

	delegatingCache := newDelegatingCache(leafClient, cache, c.scheme)

	delegatingClient, err := client.NewDelegatingClient(client.NewDelegatingClientInput{
		CacheReader: delegatingCache,
		Client:      leafClient,
		// Non-exact field matches are not supported by the cache.
		// https://github.com/kubernetes-sigs/controller-runtime/issues/612
		// TODO: Research if we can change the way we query those events so we can enable the cache for it.
		UncachedObjects:   []client.Object{&v1.Event{}},
		CacheUnstructured: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating DelegatingClient: %w", err)
	}

	ctx := context.Background()

	go delegatingCache.Start(ctx) //nolint:errcheck

	if ok := delegatingCache.WaitForCacheSync(ctx); !ok {
		return nil, errors.New("failed syncing client cache")
	}

	return delegatingClient, nil
}

func (c *delegatingCacheCluster) GetUserClient(user *auth.UserPrincipal) (client.Client, error) {
	client, err := c.cluster.GetUserClient(user)
	if err != nil {
		return nil, err
	}

	return c.makeCachingClient(client)
}

func (c *delegatingCacheCluster) GetServerClient(opts ...RESTConfigOption) (client.Client, error) {
	client, err := c.cluster.GetServerClient(opts...)
	if err != nil {
		return nil, err
	}

	return c.makeCachingClient(client)
}

func (c *delegatingCacheCluster) GetUserClientset(user *auth.UserPrincipal) (kubernetes.Interface, error) {
	return c.cluster.GetUserClientset(user)
}

func (c *delegatingCacheCluster) GetServerClientset() (kubernetes.Interface, error) {
	return c.cluster.GetServerClientset()
}

func (c *delegatingCacheCluster) GetServerConfig() (*rest.Config, error) {
	return c.cluster.GetServerConfig()
}

type delegatingCache struct {
	cache.Cache

	Client client.Reader

	scheme        *runtime.Scheme
	checkedGVKs   map[schema.GroupVersionKind]struct{}
	checkedGVKsMu *sync.Mutex
}

func newDelegatingCache(cr client.Reader, cache cache.Cache, scheme *runtime.Scheme) cache.Cache {
	dc := delegatingCache{
		Cache:         cache,
		Client:        cr,
		scheme:        scheme,
		checkedGVKs:   map[schema.GroupVersionKind]struct{}{},
		checkedGVKsMu: &sync.Mutex{},
	}

	return dc
}

func (dc delegatingCache) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
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
