package clustersmngr

import (
	"context"
	"fmt"
	"sync"

	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type key int

const (
	// Clusters Client context key
	ClustersClientCtxKey key = iota
)

// ClusterNotFoundError cluster client can be found in the pool
type ClusterNotFoundError struct {
	Cluster string
}

func (e ClusterNotFoundError) Error() string {
	return fmt.Sprintf("cluster=%s not found", e.Cluster)
}

type clusterFetchers []ClusterFetcher

func (fetchers clusterFetchers) Fetch(ctx context.Context) ([]cluster.Cluster, error) {
	clusters := []cluster.Cluster{}
	for _, fetcher := range fetchers {
		additionalClusters, err := fetcher.Fetch(ctx)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, additionalClusters...)
	}
	return clusters, nil
}

// ClusterFetcher fetches all leaf clusters
//
//counterfeiter:generate . ClusterFetcher
type ClusterFetcher interface {
	Fetch(ctx context.Context) ([]cluster.Cluster, error)
}

// ClientsPool stores all clients to the leaf clusters
//
//counterfeiter:generate . ClientsPool
type ClientsPool interface {
	Add(c client.Client, cluster cluster.Cluster) error
	Clients() map[string]client.Client
	Client(cluster string) (client.Client, error)
}

type clientsPool struct {
	clients map[string]client.Client
	mutex   sync.Mutex
}

// NewClustersClientsPool initializes a new ClientsPool
func NewClustersClientsPool() ClientsPool {
	return &clientsPool{
		clients: map[string]client.Client{},
		mutex:   sync.Mutex{},
	}
}

// Add adds a cluster client to the clients pool with the given user impersonation
func (cp *clientsPool) Add(client client.Client, cluster cluster.Cluster) error {
	cp.mutex.Lock()
	cp.clients[cluster.GetName()] = client
	cp.mutex.Unlock()

	return nil
}

// Clients returns the clusters clients
func (cp *clientsPool) Clients() map[string]client.Client {
	return cp.clients
}

// Client returns the client for the given cluster
func (cp *clientsPool) Client(name string) (client.Client, error) {
	if c, found := cp.clients[name]; found && c != nil {
		return c, nil
	}

	return nil, ClusterNotFoundError{Cluster: name}
}
