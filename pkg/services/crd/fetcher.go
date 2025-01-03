package crd

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
)

const watchCRDsFrequency = 30 * time.Second

type Fetcher interface {
	IsAvailable(ctx context.Context, clusterName, name string) bool
	IsAvailableOnClusters(ctx context.Context, name string) map[string]bool
	UpdateCRDList(context.Context)
}

// NewFetcher creates a new default fetcher with cache.
//
// With NewFetcher, it will automatically start a background go routine to watch
// CRDs.
func NewFetcher(ctx context.Context, logger logr.Logger, clustersManager clustersmngr.ClustersManager) Fetcher {
	fetcher := &defaultFetcher{
		logger:          logger,
		clustersManager: clustersManager,
		crds:            map[string][]v1.CustomResourceDefinition{},
	}

	go fetcher.watchCRDs(ctx)

	return fetcher
}

type defaultFetcher struct {
	sync.RWMutex
	logger          logr.Logger
	clustersManager clustersmngr.ClustersManager
	crds            map[string][]v1.CustomResourceDefinition
}

func (s *defaultFetcher) watchCRDs(ctx context.Context) {
	_ = wait.PollUntilContextCancel(ctx, watchCRDsFrequency, true, func(ctx context.Context) (bool, error) {
		s.UpdateCRDList(ctx)

		return false, nil
	})
}

// UpdateCRDList updates the cached CRD list.
func (s *defaultFetcher) UpdateCRDList(ctx context.Context) {
	s.Lock()
	defer s.Unlock()

	client, err := s.clustersManager.GetServerClient(ctx)
	if err != nil {
		s.logger.Error(err, "unable to get client pool")

		return
	}

	for clusterName, client := range client.ClientsPool().Clients() {
		crdList := &v1.CustomResourceDefinitionList{}

		s.crds[clusterName] = []v1.CustomResourceDefinition{}

		err := client.List(ctx, crdList)
		if err != nil {
			s.logger.Error(err, "unable to list crds", "cluster", clusterName)

			continue
		}

		s.crds[clusterName] = crdList.Items
	}
}

// IsAvailable tells if a given CRD is available on the specified cluster.
func (s *defaultFetcher) IsAvailable(_ context.Context, clusterName, name string) bool {
	s.Lock()
	defer s.Unlock()

	for _, crd := range s.crds[clusterName] {
		if crd.Name == name {
			return true
		}
	}

	return false
}

// IsAvailableOnClusters tells the availability of a given CRD on all clusters.
func (s *defaultFetcher) IsAvailableOnClusters(_ context.Context, name string) map[string]bool {
	result := map[string]bool{}

	s.Lock()
	defer s.Unlock()

	for clusterName, crds := range s.crds {
		// Set this to be sure the key is there with false value if the following
		// look-up does not say it's there.
		result[clusterName] = false

		for _, crd := range crds {
			if crd.Name == name {
				result[clusterName] = true
				break
			}
		}
	}

	return result
}
