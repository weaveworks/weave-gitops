package clustersmngr

import (
	"context"
	"fmt"
	"sync"

	"github.com/weaveworks/weave-gitops/core/clustersmngr/clusters"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type key int

const (
	// Clusters Client context key
	ClustersClientCtxKey key = iota
)

// Cluster defines a leaf cluster
type cluster struct {
	// Name defines the cluster name
	Name string `yaml:"name"`
	// Server defines cluster api address
	Server string `yaml:"server"`

	// SecretRef defines secret name that holds the cluster Bearer Token
	SecretRef string `yaml:"secretRef"`
	// BearerToken cluster access token read from SecretRef
	BearerToken string

	// TLSConfig holds configuration for TLS connection with the cluster values read from SecretRef
	TLSConfig rest.TLSClientConfig
}

// ClusterNotFoundError cluster client can be found in the pool
type ClusterNotFoundError struct {
	Cluster string
}

func (e ClusterNotFoundError) Error() string {
	return fmt.Sprintf("cluster=%s not found", e.Cluster)
}

// ClusterFetcher fetches all leaf clusters
//
//counterfeiter:generate . ClusterFetcher
type ClusterFetcher interface {
	Fetch(ctx context.Context) ([]clusters.Cluster, error)
}

// ClientsPool stores all clients to the leaf clusters
//
//counterfeiter:generate . ClientsPool
type ClientsPool interface {
	Add(c client.Client, cluster clusters.Cluster) error
	Clients() map[string]client.Client
	Client(cluster string) (client.Client, error)
}

type clientsPool struct {
	clients map[string]client.Client
	scheme  *apiruntime.Scheme
	mutex   sync.Mutex
}

// NewClustersClientsPool initializes a new ClientsPool
func NewClustersClientsPool(scheme *apiruntime.Scheme) ClientsPool {
	return &clientsPool{
		clients: map[string]client.Client{},
		scheme:  scheme,
		mutex:   sync.Mutex{},
	}
}

// Add adds a cluster client to the clients pool with the given user impersonation
func (cp *clientsPool) Add(client client.Client, cluster clusters.Cluster) error {
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
