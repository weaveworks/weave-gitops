package crd

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
)

// NewNoCacheFetcher creates a new fetcher without cache.
//
// The noCacheFetcher has no background go routine.
func NewNoCacheFetcher(clustersManager clustersmngr.ClustersManager) Fetcher {
	fetcher := &noCacheFetcher{
		clustersManager: clustersManager,
		crds:            map[string][]v1.CustomResourceDefinition{},
	}

	return fetcher
}

type noCacheFetcher struct {
	sync.RWMutex
	logger          logr.Logger
	clustersManager clustersmngr.ClustersManager
	crds            map[string][]v1.CustomResourceDefinition
}

// UpdateCRDList updates the CRD list.
func (s *noCacheFetcher) UpdateCRDList(ctx context.Context) {
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
			s.logger.Error(err, "unable to list crds")
			continue
		}

		s.crds[clusterName] = crdList.Items
	}
}

// IsAvailable tells if a given CRD is available on the specified cluster.
//
// It calls UpdateCRDList always.
func (s *noCacheFetcher) IsAvailable(ctx context.Context, clusterName, name string) bool {
	s.UpdateCRDList(ctx)

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
//
// It calls UpdateCRDList always.
func (s *noCacheFetcher) IsAvailableOnClusters(ctx context.Context, name string) map[string]bool {
	s.UpdateCRDList(ctx)

	s.Lock()
	defer s.Unlock()

	result := map[string]bool{}

	for clusterName, crds := range s.crds {
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
