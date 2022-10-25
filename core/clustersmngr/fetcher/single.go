package fetcher

import (
	"context"

	mngr "github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clusters"
)

type singleClusterFetcher struct {
	cluster clusters.Cluster
}

func NewSingleClusterFetcher(cluster clusters.Cluster) (mngr.ClusterFetcher, error) {

	return singleClusterFetcher{cluster}, nil
}

func (cf singleClusterFetcher) Fetch(ctx context.Context) ([]clusters.Cluster, error) {
	return []clusters.Cluster{cf.cluster}, nil
}
