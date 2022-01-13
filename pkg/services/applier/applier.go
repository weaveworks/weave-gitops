package applier

import (
	"bytes"
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterApplier interface {
	ApplyManifests(ctx context.Context, cluster models.Cluster, namespace string, manifests []models.Manifest) error
}

type ClusterApplySvc struct {
	Client client.Client
	Kube   kube.Kube
	Logger logger.Logger
}

var _ ClusterApplier = &ClusterApplySvc{}

func NewClusterApplier(kubeClient kube.Kube) ClusterApplier {
	return &ClusterApplySvc{
		Kube: kubeClient,
	}
}

func (a *ClusterApplySvc) ApplyManifests(ctx context.Context, cluster models.Cluster, namespace string, manifests []models.Manifest) error {
	for _, manifest := range manifests {
		ms := bytes.Split(manifest.Content, []byte("---\n"))

		for _, m := range ms {
			if len(bytes.Trim(m, " \t\n")) == 0 {
				continue
			}

			if err := a.Kube.Apply(ctx, m, namespace); err != nil {
				return fmt.Errorf("error applying manifest %s: %w", manifest.Path, err)
			}
		}
	}

	return nil
}
