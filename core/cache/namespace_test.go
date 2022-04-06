package cache_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/cache"
	"github.com/weaveworks/weave-gitops/core/logger"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestContainer_Namespace(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()
	k := fake.NewClientBuilder().WithScheme(kube.CreateScheme()).Build()

	log, err := logger.New("debug", true)
	g.Expect(err).NotTo(HaveOccurred())

	cacheContainer := cache.NewContainer(k, log)

	cacheContainer.Start(ctx)

	defer cacheContainer.Stop()

	// Global Cache
	g.Expect(cache.GlobalContainer()).To(Equal(cacheContainer))
	g.Expect(cache.NewContainer(k, log)).ToNot(Equal(cacheContainer))

	nsList := cacheContainer.Namespaces()

	g.Expect(nsList).To(HaveLen(0))

	newNamespace(ctx, "cache-container", k, g)
	newNamespace(ctx, "cache-container", k, g)

	cacheContainer.ForceRefresh(cache.NamespaceStorage)
	time.Sleep(time.Millisecond)

	nsList = cacheContainer.Namespaces()

	g.Expect(nsList).To(HaveLen(2))
}
