package cache

import (
	"context"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	v1 "k8s.io/api/core/v1"
)

const (
	NamespaceStorage StorageType = "namespace"
)

type StorageType string

type Container struct {
	namespace namespaceStore
}

var globalCacheContainer *Container

func NewContainer(ctx context.Context, clientGetter kube.ClientGetter) *Container {
	if globalCacheContainer != nil {
		return globalCacheContainer
	}

	c, _ := clientGetter.Client(ctx)

	globalCacheContainer = &Container{
		namespace: newNamespaceStore(c),
	}

	return globalCacheContainer
}

func GlobalContainer() *Container {
	return globalCacheContainer
}

func (c *Container) Start(ctx context.Context) {
	c.namespace.Start(ctx)
}

func (c *Container) Stop() {
	c.namespace.Stop()
}

func (c *Container) ForceRefresh(name StorageType) {
	switch name {
	case NamespaceStorage:
		c.namespace.ForceRefresh()
	}
}

func (c *Container) Namespaces() []v1.Namespace {
	return c.namespace.Namespaces()
}
