package clustersmngr_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestDelegatingCacheGet(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeReader := &fakeReader{}
	cache, err := cache.New(k8sEnv.Rest, cache.Options{})
	g.Expect(err).To(BeNil())

	delegatingCache := clustersmngr.NewDelegatingCache(fakeReader, cache, scheme.Scheme)

	ns := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
		},
	}

	delegatingCache.Get(context.Background(), client.ObjectKeyFromObject(ns), ns)
	delegatingCache.Get(context.Background(), client.ObjectKeyFromObject(ns), ns)

	g.Expect(fakeReader.Called).To(Equal(1))
}

func TestDelegatingCacheList(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeReader := &fakeReader{}
	cache, err := cache.New(k8sEnv.Rest, cache.Options{})
	g.Expect(err).To(BeNil())

	delegatingCache := clustersmngr.NewDelegatingCache(fakeReader, cache, scheme.Scheme)

	nsList := &corev1.NamespaceList{}

	delegatingCache.List(context.Background(), nsList)
	delegatingCache.List(context.Background(), nsList)

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
