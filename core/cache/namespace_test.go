package cache_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/cache"
	"github.com/weaveworks/weave-gitops/core/logger"
)

func TestContainer_Namespace(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()
	k := newFakeKubeClient()

	log, err := logger.New("debug", true)
	g.Expect(err).NotTo(HaveOccurred())

	cacheContainer := cache.NewContainer(k, log)

	cacheContainer.Start(ctx)

	defer cacheContainer.Stop()

	// Global Cache
	g.Expect(cache.NewContainer(k, log)).To(Equal(cacheContainer))
	g.Expect(cache.GlobalContainer()).To(Equal(cacheContainer))

	nsList := cacheContainer.Namespaces()

	g.Expect(nsList).To(HaveLen(0))

	newNamespace(ctx, "cache-container", k, g)
	newNamespace(ctx, "cache-container", k, g)

	cacheContainer.ForceRefresh(cache.NamespaceStorage)
	time.Sleep(time.Millisecond)

	nsList = cacheContainer.Namespaces()

	g.Expect(nsList).To(HaveLen(2))
}
