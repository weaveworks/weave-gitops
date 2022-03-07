package multicluster

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ClustersClientsPoolCtxKey = iota
)

const (
	ClustersConfigMapName = "leaf-clusters"
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

// ClustersFetcher fetches all the leaf clusters
type ClustersFetcher interface {
	Fetch(ctx context.Context) ([]Cluster, error)
}

type confimapClusterFetcher struct {
	hubClient client.Client
	namespace string
}

func NewConfigMapClustersFetcher(config *rest.Config, namespace string) (ClustersFetcher, error) {
	client, err := client.New(config, client.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed creating hub client: %w", err)
	}

	return confimapClusterFetcher{
		hubClient: client,
		namespace: namespace,
	}, nil
}

func (cc confimapClusterFetcher) Fetch(ctx context.Context) ([]Cluster, error) {
	clustersCm := &corev1.ConfigMap{}
	clusterCmKey := types.NamespacedName{
		Name:      ClustersConfigMapName,
		Namespace: cc.namespace,
	}

	err := cc.hubClient.Get(ctx, clusterCmKey, clustersCm)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace default: %w", err)
	}

	clusters := []Cluster{}

	err = yaml.Unmarshal([]byte(clustersCm.Data["clusters"]), &clusters)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshaling clusters: %w", err)
	}

	for i, c := range clusters {
		clusterSecret := &corev1.Secret{}
		clusterSecretKey := types.NamespacedName{
			Name:      c.SecretRef,
			Namespace: cc.namespace,
		}

		err = cc.hubClient.Get(context.Background(), clusterSecretKey, clusterSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch cluster=%s secret in namespace default: %w", c.Name, err)
		}

		clusters[i].BearerToken = string(clusterSecret.Data["token"])
	}

	return clusters, nil
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

	leafClient, err := client.New(config, client.Options{})
	if err != nil {
		return fmt.Errorf("failed to create leaf client: %w", err)
	}

	cp.clients[cluster.Name] = leafClient

	return nil
}

func (cp *clientsPool) Clients() map[string]client.Client {
	return cp.clients
}
