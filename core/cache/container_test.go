package cache_test

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/cache"
	"github.com/weaveworks/weave-gitops/core/cache/cachefakes"
	mngr "github.com/weaveworks/weave-gitops/core/clustersmngr"
)

func TestWithSimpleCaches(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()
	log := logr.Discard()

	fakeStorageOne := new(cachefakes.FakeCache)
	fakeStorageTwo := new(cachefakes.FakeCache)
	opts := cache.WithSimpleCaches(
		withFakeCacheOne(fakeStorageOne),
		withFakeCacheTwo(fakeStorageTwo),
	)
	cacheContainer := cache.NewContainer(log, opts)

	cacheContainer.Start(ctx)
	defer cacheContainer.Stop()

	g.Expect(fakeStorageOne.StartCallCount()).To(Equal(1))
	g.Expect(fakeStorageTwo.StartCallCount()).To(Equal(1))

	g.Expect(cacheContainer.ForceRefresh("")).To(Succeed())
	g.Expect(fakeStorageOne.ForceRefreshCallCount()).To(Equal(1))
	g.Expect(fakeStorageTwo.ForceRefreshCallCount()).To(Equal(1))

	g.Expect(cacheContainer.ForceRefresh("StorageTypeOne")).To(Succeed())
	g.Expect(fakeStorageOne.ForceRefreshCallCount()).To(Equal(2))
	g.Expect(fakeStorageTwo.ForceRefreshCallCount()).To(Equal(1))

	g.Expect(cacheContainer.ForceRefresh("NotFound")).To(MatchError(
		cache.CacheNotFoundErr{Cache: "NotFound"},
	))

	_, err := cacheContainer.List("StorageTypeOne")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(fakeStorageOne.ListCallCount()).To(Equal(1))

	_, err = cacheContainer.List("NotFound")
	g.Expect(err).To(MatchError(
		cache.CacheNotFoundErr{Cache: "NotFound"},
	))

	cacheContainer.Stop()
	g.Expect(fakeStorageOne.StopCallCount()).To(Equal(1))
	g.Expect(fakeStorageTwo.StopCallCount()).To(Equal(1))
}

func TestWithFuncs(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeStorage := new(cachefakes.FakeCache)
	opts := cache.WithSimpleCaches(withFakeCacheOne(fakeStorage))

	sync, _ := opts()
	g.Expect(sync).To(BeFalse())

	opts = cache.WithSyncedCaches(withFakeCacheOne(fakeStorage))

	sync, _ = opts()
	g.Expect(sync).To(BeTrue())
}

func withFakeCacheOne(c cache.Cache) cache.WithCacheFunc {
	return func(l logr.Logger, syncChan chan mngr.Cluster) (cache.StorageType, cache.Cache) {
		return "StorageTypeOne", c
	}
}

func withFakeCacheTwo(c cache.Cache) cache.WithCacheFunc {
	return func(l logr.Logger, syncChan chan mngr.Cluster) (cache.StorageType, cache.Cache) {
		return "StorageTypeTwo", c
	}
}
