package cache

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	mngr "github.com/weaveworks/weave-gitops/core/clustersmngr"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const DefaultPollIntervalSeconds = 120

type (
	// StorageType is the name of the cache, eg NamespaceStorage
	StorageType string
	// CacheCollection is a list of cache info held
	// in the container.
	CacheCollection []cacheInfo
)

// CacheNotFoundErr cache cannot be found in the collection
type CacheNotFoundErr struct {
	Cache StorageType
}

func (e CacheNotFoundErr) Error() string {
	return fmt.Sprintf("cache %s not found", e.Cache)
}

//counterfeiter:generate . Container
type Container interface {
	Start(ctx context.Context)
	Stop()
	ForceRefresh(name StorageType) error
	List(name StorageType) (interface{}, error)
}

type container struct {
	caches CacheCollection
}

//counterfeiter:generate . Cache
// Cache is the interface for each cache in the container's collection
type Cache interface {
	// Start will start the cache's tick/watch mechanism
	Start(ctx context.Context)
	// Stop will stop the cache's ticker
	Stop()
	// ForceRefresh will make the watcher refresh the cache immediately
	ForceRefresh()
	// List will return the objects from the cache. The returned value is
	// an interface{} which MUST be type-checked by the caller.
	List() interface{}
}

type cacheInfo struct {
	StorageType StorageType
	Cache       Cache
}

// NewContainer returns a new cache container. The caches in the container
// will be built up from the list of opts.
// If already set, the globalCacheContainer will be returned. Otherwise
// the new container will be set as the globalCacheContainer which, as the
// name suggests, will be made globally available.
func NewContainer(logger logr.Logger, cachesFn CachesFunc) Container {
	return &container{
		caches: With(logger, cachesFn),
	}
}

// WithCacheFunc returns a storage name and an implementor of the Cache
// interface to be added to the container's cache collection
type WithCacheFunc func(logr.Logger, chan mngr.Cluster) (StorageType, Cache)

// CachesFunc is a helper wrapper to set whether caches should be synced.
type CachesFunc func() (bool, []WithCacheFunc)

// WithSyncedCaches will return true and configure resulting caches with a sync
// channel
func WithSyncedCaches(opts ...WithCacheFunc) CachesFunc {
	return func() (bool, []WithCacheFunc) {
		return true, opts
	}
}

// WithSimpleCaches will return false and resulting caches will work
// independently of each other
func WithSimpleCaches(opts ...WithCacheFunc) CachesFunc {
	return func() (bool, []WithCacheFunc) {
		return false, opts
	}
}

// With adds caches to the collection
func With(l logr.Logger, cachesFn CachesFunc) CacheCollection {
	var clusterChan chan mngr.Cluster

	sync, options := cachesFn()

	if sync {
		clusterChan = make(chan mngr.Cluster)
	}

	caches := make(CacheCollection, 0)

	for _, opt := range options {
		name, store := opt(l, clusterChan)

		caches = append(caches, cacheInfo{StorageType: name, Cache: store})
	}

	return caches
}

// Start starts all caches in the container
func (c *container) Start(ctx context.Context) {
	for _, item := range c.caches {
		item.Cache.Start(ctx)
	}
}

// Stop stops all caches in the container
func (c *container) Stop() {
	for _, item := range c.caches {
		item.Cache.Stop()
	}
}

// ForceRefresh forces all caches in the container to refresh. If a name
// is given then only that cache is refreshed
func (c *container) ForceRefresh(name StorageType) error {
	for _, item := range c.caches {
		if name != "" && name == item.StorageType {
			item.Cache.ForceRefresh()

			return nil
		} else if name == "" {
			item.Cache.ForceRefresh()
		}
	}

	if name != "" {
		return CacheNotFoundErr{name}
	}

	return nil
}

// List returns all cached objects from the named cache. The caller MUST
// check the type of the returned value.
func (c *container) List(name StorageType) (interface{}, error) {
	for _, item := range c.caches {
		if item.StorageType == name {
			return item.Cache.List(), nil
		}
	}

	return nil, CacheNotFoundErr{name}
}
