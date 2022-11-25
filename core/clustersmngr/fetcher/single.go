package fetcher

import (
	"context"

	mngr "github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
)

type singleClusterFetcher struct {
	cluster cluster.Cluster
}

func NewSingleClusterFetcher(cluster cluster.Cluster) mngr.ClusterFetcher {
	return singleClusterFetcher{cluster}
}

func (cf singleClusterFetcher) Fetch(ctx context.Context) ([]cluster.Cluster, error) {
	return []cluster.Cluster{cf.cluster}, nil
}
