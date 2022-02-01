package fakes

import (
	"context"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/services/applicationv2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ applicationv2.FetcherFactory = &FakeFetcherFactory{}
var _ kube.ClientGetter = &FakeClientGetter{}
var _ kube.KubeGetter = &FakeKubeGetter{}

type FakeFetcherFactory struct {
	fake applicationv2.Fetcher
}

func NewFakeFetcherFactory(fake applicationv2.Fetcher) applicationv2.FetcherFactory {
	return &FakeFetcherFactory{
		fake: fake,
	}
}

func (f *FakeFetcherFactory) Create(client client.Client) applicationv2.Fetcher {
	return f.fake
}

type FakeClientGetter struct {
	client client.Client
}

func NewFakeClientGetter(client client.Client) kube.ClientGetter {
	return &FakeClientGetter{
		client: client,
	}
}

func (g *FakeClientGetter) Client(ctx context.Context) (client.Client, error) {
	return g.client, nil
}

type FakeKubeGetter struct {
	kube kube.Kube
}

func NewFakeKubeGetter(kube kube.Kube) kube.KubeGetter {
	return &FakeKubeGetter{
		kube: kube,
	}
}

func (g *FakeKubeGetter) Kube(ctx context.Context) (kube.Kube, error) {
	return g.kube, nil
}
