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
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cs *coreServer) ListHelmReleases(ctx context.Context, msg *pb.ListHelmReleasesRequest) (*pb.ListHelmReleasesResponse, error) {
	k8s, err := cs.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	l := &helmv2.HelmReleaseList{}

	if err := list(ctx, k8s, temporarilyEmptyAppName, msg.Namespace, l); err != nil {
		return nil, err
	}

	var results []*pb.HelmRelease
	for _, helmRelease := range l.Items {
		results = append(results, types.HelmReleaseToProto(&helmRelease, []*pb.GroupVersionKind{}))
	}

	return &pb.ListHelmReleasesResponse{
		HelmReleases: results,
	}, nil
}

func (cs *coreServer) GetHelmRelease(ctx context.Context, msg *pb.GetHelmReleaseRequest) (*pb.GetHelmReleaseResponse, error) {
	k8s, err := cs.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	helmRelease := helmv2.HelmRelease{}

	if err = get(ctx, k8s, msg.Name, msg.Namespace, &helmRelease); err != nil {
		return nil, err
	}

	inventory, err := getHelmReleaseInventory(ctx, helmRelease, k8s)
	if err != nil {
		return nil, err
	}

	return &pb.GetHelmReleaseResponse{
		HelmRelease: types.HelmReleaseToProto(&helmRelease, inventory),
	}, err
}

func getHelmReleaseInventory(ctx context.Context, helmRelease v2beta1.HelmRelease, k8s client.Client) ([]*pb.GroupVersionKind, error) {
	storageNamespace := helmRelease.GetNamespace()
	if helmRelease.Spec.StorageNamespace != "" {
		storageNamespace = helmRelease.Spec.StorageNamespace
	}

	storageName := helmRelease.GetName()
	if helmRelease.Spec.ReleaseName != "" {
		storageName = helmRelease.Spec.ReleaseName
	}

	storageVersion := helmRelease.Status.LastReleaseRevision
	if storageVersion < 1 {
		// skip release if it failed to install
		return nil, nil
	}

	storageSecret := v1.Secret{}
	secretName := fmt.Sprintf("sh.helm.release.v1.%s.v%v", storageName, storageVersion)

	if err := get(ctx, k8s, secretName, storageNamespace, &storageSecret); err != nil {
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
