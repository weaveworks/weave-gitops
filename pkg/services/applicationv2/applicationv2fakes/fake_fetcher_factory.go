package applicationv2fakes

import (
	"github.com/weaveworks/weave-gitops/pkg/services/applicationv2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ applicationv2.FetcherFactory = &FakeFetcherFactory{}

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
