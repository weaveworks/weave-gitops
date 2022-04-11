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

	log, err := logger.New("debug", true)
	g.Expect(err).NotTo(HaveOccurred())

	cacheContainer, err := cache.NewContainer(ctx, k8sEnv.Rest, log)
	g.Expect(err).NotTo(HaveOccurred())

	cacheContainer.Start(ctx)

	defer cacheContainer.Stop()

	// Global Cache
	g.Expect(cache.NewContainer(ctx, k8sEnv.Rest, log)).To(Equal(cacheContainer))
	g.Expect(cache.GlobalContainer()).To(Equal(cacheContainer))

	cacheContainer.ForceRefresh(cache.NamespaceStorage)

	nsList := cacheContainer.Namespaces()
	initialDefault := len(nsList["Default"])

	g.Expect(nsList).To(HaveLen(1))

	newNamespace(ctx, "cache-container", g)
	newNamespace(ctx, "cache-container", g)

	cacheContainer.ForceRefresh(cache.NamespaceStorage)
	time.Sleep(time.Millisecond * 200)

	nsList = cacheContainer.Namespaces()

	g.Expect(nsList).To(HaveLen(1))
	g.Expect(nsList["Default"]).To(HaveLen(initialDefault + 2))
}
