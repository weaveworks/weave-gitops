package cache_test

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/cache"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func TestContainer_Namespace(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	log := logr.Discard()

	opts := cache.WithSimpleCaches(cache.WithNamespaceCache(k8sEnv.Rest))
	cacheContainer := cache.NewContainer(log, opts)

	cacheContainer.Start(ctx)

	defer cacheContainer.Stop()

	g.Expect(cacheContainer.ForceRefresh(cache.NamespaceStorage)).To(Succeed())

	objs, err := cacheContainer.List(cache.NamespaceStorage)
	g.Expect(err).NotTo(HaveOccurred())

	nsList, ok := objs.(map[string][]v1.Namespace)
	g.Expect(ok).To(BeTrue())
	g.Expect(nsList).To(HaveLen(1))

	initialDefault := len(nsList["Default"])

	newNamespace(ctx, "cache-container", g)
	newNamespace(ctx, "cache-container", g)

	g.Expect(cacheContainer.ForceRefresh(cache.NamespaceStorage)).To(Succeed())

	g.Eventually(func(g Gomega) int {
		objs, err := cacheContainer.List(cache.NamespaceStorage)
		g.Expect(err).NotTo(HaveOccurred())
		nsList, ok := objs.(map[string][]v1.Namespace)
		g.Expect(ok).To(BeTrue())

		return len(nsList["Default"])
	}).Should(Equal(initialDefault + 2))
}

func TestContainer_NamespaceWithClusterSync(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	log := logr.Discard()

	opts := cache.WithSyncedCaches(
		withFakeClusterCache(k8sEnv.Rest),
		cache.WithNamespaceCache(k8sEnv.Rest),
	)
	cacheContainer := cache.NewContainer(log, opts)

	cacheContainer.Start(ctx)

	defer cacheContainer.Stop()

	g.Expect(cacheContainer.ForceRefresh(cache.NamespaceStorage)).To(Succeed())

	objs, err := cacheContainer.List(cache.NamespaceStorage)
	g.Expect(err).NotTo(HaveOccurred())

	nsList, ok := objs.(map[string][]v1.Namespace)
	g.Expect(ok).To(BeTrue())
	g.Expect(nsList).To(HaveLen(3))
}

func newNamespace(ctx context.Context, prefix string, g *GomegaWithT) *v1.Namespace {
	ns := &v1.Namespace{}
	ns.Name = prefix + "kube-test-" + rand.String(5)

	g.Expect(k8sEnv.Client.Create(ctx, ns)).To(Succeed())

	return ns
}
