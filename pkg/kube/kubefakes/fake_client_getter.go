package kubefakes

import (
	"context"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ kube.ClientGetter = &FakeClientGetter{}

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
