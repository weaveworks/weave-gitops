package clustersmngr_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestDelegatingCacheGet(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeReader := &fakeReader{}
	cache, err := cache.New(k8sEnv.Rest, cache.Options{})
	g.Expect(err).To(BeNil())

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go cache.Start(ctx)

	if ok := cache.WaitForCacheSync(ctx); !ok {
		g.Fail("failed syncing client cache")
	}

	delegatingCache := clustersmngr.NewDelegatingCache(fakeReader, cache, scheme.Scheme)

	ns := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: "test" + rand.String(5),
		},
	}

	g.Expect(k8sEnv.Client.Create(ctx, ns)).To(Succeed())

	err = delegatingCache.Get(ctx, client.ObjectKeyFromObject(ns), ns)
	g.Expect(err).To(BeNil())

	err = delegatingCache.Get(ctx, client.ObjectKeyFromObject(ns), ns)
	g.Expect(err).To(BeNil())

	g.Expect(fakeReader.Called).To(Equal(1))
}

func TestDelegatingCacheList(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeReader := &fakeReader{}
	cache, err := cache.New(k8sEnv.Rest, cache.Options{})
	g.Expect(err).To(BeNil())

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go cache.Start(ctx)

	if ok := cache.WaitForCacheSync(ctx); !ok {
		g.Fail("failed syncing client cache")
	}

	ns := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: "test" + rand.String(5),
		},
	}

	g.Expect(k8sEnv.Client.Create(ctx, ns)).To(Succeed())

	delegatingCache := clustersmngr.NewDelegatingCache(fakeReader, cache, scheme.Scheme)

	nsList := &corev1.NamespaceList{}

	err = delegatingCache.List(ctx, nsList)
	g.Expect(err).To(BeNil())

	err = delegatingCache.List(ctx, nsList)
	g.Expect(err).To(BeNil())

	g.Expect(fakeReader.Called).To(Equal(1))
}

type fakeReader struct {
	Called int
}

func (f *fakeReader) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	f.Called++
	return nil
}

func (f *fakeReader) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	f.Called++
	return nil
}
