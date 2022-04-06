package cache

import (
	"context"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	NamespaceStorage StorageType = "namespace"
)

type StorageType string

type Container struct {
	namespace namespaceStore
	logger    logr.Logger
}

var globalCacheContainer *Container

func NewContainer(crClient client.Client, logger logr.Logger) *Container {
	globalCacheContainer = &Container{
		namespace: newNamespaceStore(crClient, logger),
		logger:    logger,
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
