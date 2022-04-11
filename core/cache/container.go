package cache

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const (
	NamespaceStorage StorageType = "namespace"
)

type StorageType string

//counterfeiter:generate . Container
type Container interface {
	Start(ctx context.Context)
	Stop()
	ForceRefresh(name StorageType)
	Namespaces() map[string][]v1.Namespace
}

type container struct {
	namespace namespaceStore
	logger    logr.Logger
}

var globalCacheContainer Container

func NewContainer(ctx context.Context, restCfg *rest.Config, logger logr.Logger) (Container, error) {
	if globalCacheContainer != nil {
		return globalCacheContainer, nil
	}

	clusterClient, err := newClustersClient(ctx, restCfg)
	if err != nil {
		return nil, err
	}

	globalCacheContainer = &container{
		namespace: newNamespaceStore(clusterClient, logger),
		logger:    logger,
	}

	return globalCacheContainer, nil
}

func GlobalContainer() Container {
	return globalCacheContainer
}

func (c *container) Start(ctx context.Context) {
	c.namespace.Start(ctx)
}

func (c *container) Stop() {
	c.namespace.Stop()
}

func (c *container) ForceRefresh(name StorageType) {
	switch name {
	case NamespaceStorage:
		c.namespace.ForceRefresh()
	}
}

func (c *container) Namespaces() map[string][]v1.Namespace {
	return c.namespace.Namespaces()
}

func newClustersClient(ctx context.Context, restCfg *rest.Config) (clustersmngr.Client, error) {
	clustersFetcher, err := clustersmngr.NewSingleClusterFetcher(restCfg)
	if err != nil {
		return nil, fmt.Errorf("failed fetching clusters: %w", err)
	}

	clusters, err := clustersFetcher.Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed fetching clusters: %w", err)
	}

	clientsPool := clustersmngr.NewClustersClientsPool()
	for _, c := range clusters {
		if err := clientsPool.Add(clientConfig(restCfg), c); err != nil {
			return nil, err
		}
	}

	return clustersmngr.NewClient(clientsPool), nil
}

func clientConfig(restCfg *rest.Config) clustersmngr.ClusterClientConfig {
	return func(cluster clustersmngr.Cluster) *rest.Config {
		return restCfg
	}
}
