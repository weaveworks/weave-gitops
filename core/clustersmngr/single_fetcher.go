package clustersmngr

import (
	"context"

	"k8s.io/client-go/rest"
)

type singleClusterFetcher struct {
	restConfig *rest.Config
}

func NewSingleClusterFetcher(config *rest.Config, _ string) (ClusterFetcher, error) {
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
			TLSConfig:   cf.restConfig.TLSClientConfig,
		},
	}, nil
}
