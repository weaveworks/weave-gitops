package server

import (
	"context"
	"fmt"
	"strings"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

func getHelmReleaseInventory(ctx context.Context, helmRelease helmv2.HelmRelease, c clustersmngr.Client, cluster string) ([]*pb.GroupVersionKind, error) {
	k8sClient, err := c.Scoped(cluster)
	if err != nil {
		return nil, fmt.Errorf("error getting scoped client for cluster=%s: %w", cluster, err)
	}

	objects, err := getHelmReleaseObjects(ctx, k8sClient, &helmRelease)
	if err != nil {
		return nil, err
	}

	var gvk []*pb.GroupVersionKind

	found := map[string]bool{}

	for _, entry := range objects {
		idstr := strings.Join([]string{entry.GetAPIVersion(), entry.GetKind()}, "_")

		if !found[idstr] {
			found[idstr] = true

			gvk = append(gvk, &pb.GroupVersionKind{
				Group:   entry.GroupVersionKind().Group,
				Version: entry.GroupVersionKind().Version,
				Kind:    entry.GroupVersionKind().Kind,
			})
		}
	}

	return gvk, nil
}
