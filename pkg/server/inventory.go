package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/ssa"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cli-utils/pkg/object"
)

func getKustomizeInventory(kustomization *kustomizev2.Kustomization) ([]*pb.GroupVersionKind, error) {
	if kustomization.Status.Inventory == nil {
		return nil, nil
	}

	var gvk []*pb.GroupVersionKind

	found := map[string]bool{}

	for _, entry := range kustomization.Status.Inventory.Entries {
		objMeta, err := object.ParseObjMetadata(entry.ID)
		if err != nil {
			return gvk, fmt.Errorf("invalid inventory item '%s', error: %w", entry.ID, err)
		}

		idstr := strings.Join([]string{objMeta.GroupKind.Group, entry.Version, objMeta.GroupKind.Kind}, "_")

		if !found[idstr] {
			found[idstr] = true

			gvk = append(gvk, &pb.GroupVersionKind{
				Group:   objMeta.GroupKind.Group,
				Version: entry.Version,
				Kind:    objMeta.GroupKind.Kind,
			})
		}
	}

	return gvk, nil
}

type hrStorage struct {
	Name     string `json:"name,omitempty"`
	Manifest string `json:"manifest,omitempty"`
}

func getHelmInventory(hr *helmv2.HelmRelease, kubeClient kube.Kube) ([]*pb.GroupVersionKind, error) {
	storageNamespace := hr.GetNamespace()
	if hr.Spec.StorageNamespace != "" {
		storageNamespace = hr.Spec.StorageNamespace
	}

	storageName := hr.GetName()
	if hr.Spec.ReleaseName != "" {
		storageName = hr.Spec.ReleaseName
	} else if hr.Spec.TargetNamespace != "" {
		storageName = strings.Join([]string{hr.Spec.TargetNamespace, hr.Name}, "-")
	}

	storageVersion := hr.Status.LastReleaseRevision
	// skip release if it failed to install
	if storageVersion < 1 {
		return nil, nil
	}

	storageSecret, err := kubeClient.GetSecret(context.TODO(), types.NamespacedName{
		Namespace: storageNamespace,
		Name:      fmt.Sprintf("sh.helm.release.v1.%s.v%v", storageName, storageVersion),
	})

	if err != nil {
		return nil, err
	}

	releaseData, releaseFound := storageSecret.Data["release"]
	if !releaseFound {
		return nil, fmt.Errorf("failed to decode the Helm storage object for HelmRelease '%s'", hr.Name)
	}

	// adapted from https://github.com/helm/helm/blob/02685e94bd3862afcb44f6cd7716dbeb69743567/pkg/storage/driver/util.go
	var b64 = base64.StdEncoding

	b, err := b64.DecodeString(string(releaseData))
	if err != nil {
		return nil, err
	}

	var magicGzip = []byte{0x1f, 0x8b, 0x08}
	if bytes.Equal(b[0:3], magicGzip) {
		r, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		defer r.Close()

		b2, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}

		b = b2
	}

	var storage hrStorage
	if err := json.Unmarshal(b, &storage); err != nil {
		return nil, fmt.Errorf("failed to decode the Helm storage object for HelmRelease '%s': %w", hr.Name, err)
	}

	objects, err := ssa.ReadObjects(strings.NewReader(storage.Manifest))
	if err != nil {
		return nil, fmt.Errorf("failed to read the Helm storage object for HelmRelease '%s': %w", hr.Name, err)
	}

	var gvk []*pb.GroupVersionKind

	found := map[string]bool{}

	for _, entry := range objects {
		entry.GetAPIVersion()
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
