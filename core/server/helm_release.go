package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/pkg/ssa"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cs *coreServer) ListHelmReleases(ctx context.Context, msg *pb.ListHelmReleasesRequest) (*pb.ListHelmReleasesResponse, error) {
	clustersClient := clustersmngr.ClientFromCtx(ctx)

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &helmv2.HelmReleaseList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, true); err != nil {
		return nil, err
	}

	var results []*pb.HelmRelease

	for clusterName, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*helmv2.HelmReleaseList)
			if !ok {
				continue
			}

			for _, helmrelease := range list.Items {
				inv, err := getHelmReleaseInventory(ctx, helmrelease, clustersClient, clusterName)
				if err != nil {
					return nil, err
				}

				results = append(results, types.HelmReleaseToProto(&helmrelease, clusterName, inv))
			}
		}
	}

	return &pb.ListHelmReleasesResponse{
		HelmReleases: results,
	}, nil
}

func (cs *coreServer) GetHelmRelease(ctx context.Context, msg *pb.GetHelmReleaseRequest) (*pb.GetHelmReleaseResponse, error) {
	clustersClient := clustersmngr.ClientFromCtx(ctx)

	helmRelease := helmv2.HelmRelease{}
	key := client.ObjectKey{
		Name:      msg.Name,
		Namespace: msg.Namespace,
	}

	if err := clustersClient.Get(ctx, msg.ClusterName, key, &helmRelease); err != nil {
		return nil, err
	}

	inventory, err := getHelmReleaseInventory(ctx, helmRelease, clustersClient, msg.ClusterName)
	if err != nil {
		return nil, err
	}

	return &pb.GetHelmReleaseResponse{
		HelmRelease: types.HelmReleaseToProto(&helmRelease, msg.ClusterName, inventory),
	}, err
}

func getHelmReleaseInventory(ctx context.Context, helmRelease v2beta1.HelmRelease, c clustersmngr.Client, cluster string) ([]*pb.GroupVersionKind, error) {
	storageNamespace := helmRelease.GetStorageNamespace()

	storageName := helmRelease.GetReleaseName()

	storageVersion := helmRelease.Status.LastReleaseRevision
	if storageVersion < 1 {
		// skip release if it failed to install
		return nil, nil
	}

	storageSecret := &v1.Secret{}
	secretName := fmt.Sprintf("sh.helm.release.v1.%s.v%v", storageName, storageVersion)
	key := client.ObjectKey{
		Name:      secretName,
		Namespace: storageNamespace,
	}

	if err := c.Get(ctx, cluster, key, storageSecret); err != nil {
		return nil, err
	}

	releaseData, releaseFound := storageSecret.Data["release"]
	if !releaseFound {
		return nil, fmt.Errorf("failed to decode the Helm storage object for HelmRelease '%s'", helmRelease.Name)
	}

	byteData, err := base64.StdEncoding.DecodeString(string(releaseData))
	if err != nil {
		return nil, err
	}

	var magicGzip = []byte{0x1f, 0x8b, 0x08}
	if bytes.Equal(byteData[0:3], magicGzip) {
		r, err := gzip.NewReader(bytes.NewReader(byteData))
		if err != nil {
			return nil, err
		}

		defer r.Close()

		uncompressedByteData, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		byteData = uncompressedByteData
	}

	storage := types.HelmReleaseStorage{}
	if err := json.Unmarshal(byteData, &storage); err != nil {
		return nil, fmt.Errorf("failed to decode the Helm storage object for HelmRelease '%s': %w", helmRelease.Name, err)
	}

	objects, err := ssa.ReadObjects(strings.NewReader(storage.Manifest))
	if err != nil {
		return nil, fmt.Errorf("failed to read the Helm storage object for HelmRelease '%s': %w", helmRelease.Name, err)
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
