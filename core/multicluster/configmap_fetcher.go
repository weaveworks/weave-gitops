package multicluster

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type confimapClusterFetcher struct {
	hubClient client.Client
	namespace string
}

func NewConfigMapClustersFetcher(config *rest.Config, namespace string) (ClusterFetcher, error) {
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
		return nil, fmt.Errorf("failed getting clusters config map: %w", err)
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
