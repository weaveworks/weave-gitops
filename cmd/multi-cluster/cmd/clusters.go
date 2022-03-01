package multi_cluster

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type cluster struct {
	Name                     string `yaml:"name"`
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
	SecretRef                string `yaml:"secretRef"`

	bearerToken string
}

type ClustersFetcher interface {
	Fetch(ctx context.Context) ([]cluster, error)
}

type clustersClient struct {
	k8sClient client.Client
}

func NewConfigMapClustersFetcher() (ClustersFetcher, error) {
	restConfig, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	k8sClient, err := client.New(restConfig, client.Options{})
	if err != nil {
		return nil, err
	}

	return clustersClient{
		k8sClient: k8sClient,
	}, nil
}

func (cc clustersClient) Fetch(ctx context.Context) ([]cluster, error) {
	clustersCm := &corev1.ConfigMap{}
	clusterCmKey := types.NamespacedName{
		Name:      "clusters", // TODO: use contant
		Namespace: "default",  // TODO: allow custom ns or stick with flux-system
	}

	err := cc.k8sClient.Get(ctx, clusterCmKey, clustersCm)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace default: %w", err)
	}

	clusters := []cluster{}

	err = yaml.Unmarshal([]byte(clustersCm.Data["clusters"]), &clusters)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshaling clusters: %w", err)
	}

	for i, c := range clusters {
		clusterSecret := &corev1.Secret{}
		clusterSecretKey := types.NamespacedName{
			Name:      c.SecretRef,
			Namespace: "default", // TODO: allow custom ns or stick with flux-system
		}

		err = cc.k8sClient.Get(context.Background(), clusterSecretKey, clusterSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch cluster=%s secret in namespace default: %w", c.Name, err)
		}

		clusters[i].bearerToken = string(clusterSecret.Data["token"])
	}

	return clusters, nil
}

type clientsPool struct {
	clients map[string]client.Client
}

func NewClustersClientsPool() *clientsPool {
	return &clientsPool{
		clients: map[string]client.Client{},
	}
}

func (cp *clientsPool) Add(user User, cluster cluster) error {
	leafConfig := &rest.Config{
		Host:        cluster.Server,
		BearerToken: cluster.bearerToken,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true, // TODO: proper certs loading
		},
		Impersonate: rest.ImpersonationConfig{
			UserName: user.Email,
			Groups:   []string{user.Team},
		},
	}

	leafClient, err := client.New(leafConfig, client.Options{})
	if err != nil {
		return fmt.Errorf("failed to create leaf client: %w", err)
	}

	cp.clients[cluster.Name] = leafClient

	return nil
}

func (cp *clientsPool) Clients() map[string]client.Client {
	return cp.clients
}
