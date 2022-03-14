package multicluster

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const (
	ClustersClientsPoolCtxKey = iota
)

const (
	ClustersConfigMapName = "leaf-clusters"
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
}

//ClusterFetcher fetches all leaf clusters
//counterfeiter:generate . ClusterFetcher
type ClusterFetcher interface {
	Fetch(ctx context.Context) ([]Cluster, error)
}

type ClientsPool interface {
	Add(user *auth.UserPrincipal, cluster Cluster) error
	Clients() map[string]client.Client
}

type clientsPool struct {
	clients map[string]client.Client
}

func NewClustersClientsPool() ClientsPool {
	return &clientsPool{
		clients: map[string]client.Client{},
	}
}

func (cp *clientsPool) Add(user *auth.UserPrincipal, cluster Cluster) error {
	config := &rest.Config{
		Host:        cluster.Server,
		BearerToken: cluster.BearerToken,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true, // TODO: proper certs loading
		},
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

func (cp *clientsPool) Clients() map[string]client.Client {
	return cp.clients
}
