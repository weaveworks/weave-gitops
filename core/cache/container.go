package cache

import (
	"context"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	v1 "k8s.io/api/core/v1"
)

type Container struct {
	namespace *namespaceStore
}

func NewContainer(ctx context.Context, clientGetter kube.ClientGetter) *Container {
	c, _ := clientGetter.Client(ctx)

	return &Container{
		namespace: newNamespaceStore(c),
	}
}

func (c *Container) Start(ctx context.Context) {
	c.namespace.Start(ctx)
}

func (c *Container) Namespaces() []v1.Namespace {
	return c.namespace.Namespaces()
}
