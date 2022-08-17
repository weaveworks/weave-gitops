package clustersmngr

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/cli-utils/pkg/flowcontrol"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type key int

const (
	// Clusters Client context key
	ClustersClientCtxKey key = iota
	// DefaultCluster name
	DefaultCluster = "Default"
	// ClientQPS is the QPS to use while creating the k8s clients (actually a float32)
	ClientQPS = 1000
	// ClientBurst is the burst to use while creating the k8s clients
	ClientBurst = 2000
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
//counterfeiter:generate . ClientsPool
type ClientsPool interface {
	Add(cfg ClusterClientConfigFunc, cluster Cluster) error
	Clients() map[string]client.Client
	Client(cluster string) (client.Client, error)
}

type clientsPool struct {
	clients map[string]client.Client
	scheme  *apiruntime.Scheme
	mutex   sync.Mutex
}

type ClusterClientConfigFunc func(Cluster) (*rest.Config, error)

// ClientConfigWithUser returns a function that returns a *rest.Config with the relevant
// user authentication details pre-defined for a given cluster.
func ClientConfigWithUser(user *auth.UserPrincipal) ClusterClientConfigFunc {
	return func(cluster Cluster) (*rest.Config, error) {
		config := &rest.Config{
			Host:            cluster.Server,
			TLSClientConfig: cluster.TLSConfig,
			Timeout:         kubeClientTimeout,
			Dial: (&net.Dialer{
				Timeout: kubeClientDialTimeout,
				// KeepAlive is default to 30s within client-go.
				KeepAlive: kubeClientDialKeepAlive,
			}).DialContext,
		}

		if !user.Valid() {
			return nil, fmt.Errorf("No user ID or Token found in UserPrincipal.")
		} else if user.Token != "" {
			config.BearerToken = user.Token
		} else {
			config.BearerToken = cluster.BearerToken
			config.Impersonate = rest.ImpersonationConfig{
				UserName: user.ID,
				Groups:   user.Groups,
			}
		}

		// flowcontrol.IsEnabled makes a request to the K8s API of the cluster stored in the config.
		// It does a HEAD request to /livez/ping which uses the config.Dial timeout. We can use this
		// function to error early rather than wait to call client.New.
		enabled, err := flowcontrol.IsEnabled(context.Background(), config)
		if err != nil {
			return nil, fmt.Errorf("error querying cluster for flowcontrol config: %w", err)
		}

		if enabled {
			// Enabled & negative QPS and Burst indicate that the client would use the rate limit set by the server.
			// Ref: https://github.com/kubernetes/kubernetes/blob/v1.24.0/staging/src/k8s.io/client-go/rest/config.go#L354-L364
			config.QPS = -1
			config.Burst = -1

			return config, nil
		}

		config.QPS = ClientQPS
		config.Burst = ClientBurst

		return config, nil
	}
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
func (cp *clientsPool) Add(cfgFunc ClusterClientConfigFunc, cluster Cluster) error {
	config, err := cfgFunc(cluster)
	if err != nil {
		return fmt.Errorf("error building cluster client config: %w", err)
	}

	leafClient, err := client.New(config, client.Options{
		Scheme: cp.scheme,
	})
	if err != nil {
		return fmt.Errorf("failed to create leaf client: %w", err)
	}

	cp.mutex.Lock()
	cp.clients[cluster.Name] = leafClient
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
