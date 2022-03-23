package cache_test

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type fakeKubeClient struct {
	namespaces []v1.Namespace
}

func (f *fakeKubeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return nil
}

func (f *fakeKubeClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if l, ok := list.(*v1.NamespaceList); ok {
		l.Items = f.namespaces
	}

	return nil
}

func (f *fakeKubeClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if ns, ok := obj.(*v1.Namespace); ok {
		f.namespaces = append(f.namespaces, *ns)
	}

	return nil
}

func (f *fakeKubeClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return nil
}

func (f *fakeKubeClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return nil
}

func (f *fakeKubeClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return nil
}

func (f *fakeKubeClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return nil
}

func (f *fakeKubeClient) Status() client.StatusWriter {
	return nil
}

func (f *fakeKubeClient) Scheme() *runtime.Scheme {
	return nil
}

func (f *fakeKubeClient) RESTMapper() meta.RESTMapper {
	return nil
}

func newFakeKubeClient() *fakeKubeClient {
	return &fakeKubeClient{namespaces: []v1.Namespace{}}
}
