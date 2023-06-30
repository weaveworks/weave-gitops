package crd

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const watchCRDsFrequency = 30 * time.Second

type Fetcher interface {
	IsAvailable(clusterName, name string) bool
	IsAvailableOnClusters(name string) map[string]bool
	UpdateCRDList()
}

// NewFetcher creates a new default fetcher with cache.
//
// With NewFetcher, it will automatically start a background go routine to watch
// CRDs.
func NewFetcher(logger logr.Logger, clustersManager clustersmngr.ClustersManager) Fetcher {
	fetcher := &defaultFetcher{
		logger:          logger,
		clustersManager: clustersManager,
		crds:            map[string][]v1.CustomResourceDefinition{},
	}

	go fetcher.watchCRDs()

	return fetcher
}

type defaultFetcher struct {
	sync.RWMutex
	logger          logr.Logger
	clustersManager clustersmngr.ClustersManager
	crds            map[string][]v1.CustomResourceDefinition
}

func (s *defaultFetcher) watchCRDs() {
	//nolint:staticcheck
	_ = wait.PollImmediateInfinite(watchCRDsFrequency, func() (bool, error) {
		s.UpdateCRDList()

		return false, nil
	})
}

// UpdateCRDList updates the cached CRD list.
func (s *defaultFetcher) UpdateCRDList() {
	s.Lock()
	defer s.Unlock()

	ctx := context.Background()

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
func (s *defaultFetcher) IsAvailable(clusterName, name string) bool {
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
func (s *defaultFetcher) IsAvailableOnClusters(name string) map[string]bool {
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
