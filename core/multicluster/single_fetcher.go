package multicluster

import (
	"context"

	"k8s.io/client-go/rest"
)

type singleClusterFetcher struct {
	restConfig *rest.Config
}

func NewSingleClustersFetcher(config *rest.Config, _ string) (ClustersFetcher, error) {

	return singleClusterFetcher{
		restConfig: config,
	}, nil
}

func (cf singleClusterFetcher) Fetch(ctx context.Context) ([]Cluster, error) {
	return []Cluster{
		{
			Name:        "Default",
			Server:      cf.restConfig.Host,
			BearerToken: cf.restConfig.BearerToken,
		},
	}, nil
}
