package fetcher

import (
	"context"

	mngr "github.com/weaveworks/weave-gitops/core/clustersmngr"
	"k8s.io/client-go/rest"
)

type singleClusterFetcher struct {
	restConfig *rest.Config
}

func NewSingleClusterFetcher(config *rest.Config) mngr.ClusterFetcher {
	return singleClusterFetcher{
		restConfig: config,
	}
}

func (cf singleClusterFetcher) Fetch(ctx context.Context) ([]mngr.Cluster, error) {
	return []mngr.Cluster{
		{
			Name:        mngr.DefaultCluster,
			Server:      cf.restConfig.Host,
			BearerToken: cf.restConfig.BearerToken,
			TLSConfig:   cf.restConfig.TLSClientConfig,
		},
	}, nil
}
