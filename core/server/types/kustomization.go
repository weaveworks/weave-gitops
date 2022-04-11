package types

import (
	"fmt"
	"strings"

	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"sigs.k8s.io/cli-utils/pkg/object"
)

func KustomizationToProto(kustomization *v1beta2.Kustomization, clusterName string) (*pb.Kustomization, error) {
	var kind pb.SourceRef_SourceKind

	switch kustomization.Spec.SourceRef.Kind {
	case v1beta1.GitRepositoryKind:
		kind = pb.SourceRef_GitRepository
	case v1beta1.HelmRepositoryKind:
		kind = pb.SourceRef_HelmRepository
	case v1beta1.BucketKind:
		kind = pb.SourceRef_Bucket
	}

	inv, err := getKustomizeInventory(kustomization)
	if err != nil {
		return nil, fmt.Errorf("coverting kustomization to proto: %w", err)
	}

	return &pb.Kustomization{
		Name:      kustomization.Name,
		Namespace: kustomization.Namespace,
		Path:      kustomization.Spec.Path,
		SourceRef: &pb.SourceRef{
			Kind: kind,
			Name: kustomization.Spec.SourceRef.Name,
		},
		Interval:                durationToInterval(kustomization.Spec.Interval),
		Conditions:              mapConditions(kustomization.Status.Conditions),
		LastAppliedRevision:     kustomization.Status.LastAppliedRevision,
		LastAttemptedRevision:   kustomization.Status.LastAttemptedRevision,
		LastHandledReconciledAt: kustomization.Status.LastHandledReconcileAt,
		Inventory:               inv,
		Suspended:               kustomization.Spec.Suspend,
		ClusterName:             clusterName,
	}, nil
}

func getKustomizeInventory(kustomization *v1beta2.Kustomization) ([]*pb.GroupVersionKind, error) {
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
