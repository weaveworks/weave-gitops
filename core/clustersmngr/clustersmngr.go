package clustersmngr

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type key int

const (
	// Clusters Client context key
	ClustersClientCtxKey key = iota
)

var (
	scheme = kube.CreateScheme()
)

// Cluster defines a leaf cluster
type Cluster struct {
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

//ClusterFetcher fetches all leaf clusters
//counterfeiter:generate . ClusterFetcher
type ClusterFetcher interface {
	Fetch(ctx context.Context) ([]Cluster, error)
}

// ClientsPool stores all clients to the leaf clusters
type ClientsPool interface {
	Add(user *auth.UserPrincipal, cluster Cluster) error
	Clients() map[string]client.Client
	Client(clsuter string) (client.Client, error)
}

type clientsPool struct {
	clients map[string]client.Client
}

// NewClustersClientsPool initializes a new ClientsPool
func NewClustersClientsPool() ClientsPool {
	return &clientsPool{
		clients: map[string]client.Client{},
	}
}

// Add adds a cluster client to the clients pool with the given user impersonation
func (cp *clientsPool) Add(user *auth.UserPrincipal, cluster Cluster) error {
	config := &rest.Config{
		Host:            cluster.Server,
		BearerToken:     cluster.BearerToken,
		TLSClientConfig: cluster.TLSConfig,
		Impersonate: rest.ImpersonationConfig{
			UserName: user.ID,
			Groups:   user.Groups,
		},
	}

	leafClient, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return fmt.Errorf("failed to create leaf client: %w", err)
	}

	cp.clients[cluster.Name] = leafClient

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
