package kubefakes

import (
	"context"

	"github.com/weaveworks/weave-gitops/pkg/kube"
)

var _ kube.KubeGetter = &FakeKubeGetter{}

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
