package crd

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

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

func (s *noCacheFetcher) UpdateCRDList() {
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
			s.logger.Error(err, "unable to list crds")
			continue
		}

		s.crds[clusterName] = crdList.Items
	}
}

func (s *noCacheFetcher) IsAvailable(clusterName, name string) bool {
	s.UpdateCRDList()

	s.Lock()
	defer s.Unlock()

	for _, crd := range s.crds[clusterName] {
		if crd.Name == name {
			return true
		}
	}

	return false
}

func (s *noCacheFetcher) IsAvailableOnClusters(name string) map[string]bool {
	s.UpdateCRDList()

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
