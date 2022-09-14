package types

import (
	"fmt"
	"strings"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"sigs.k8s.io/cli-utils/pkg/object"
)

func KustomizationToProto(kustomization *kustomizev1.Kustomization, clusterName string, tenant string) (*pb.Kustomization, error) {
	var kind pb.FluxObjectKind

	switch kustomization.Spec.SourceRef.Kind {
	case sourcev1.GitRepositoryKind:
		kind = pb.FluxObjectKind_KindGitRepository
	case sourcev1.HelmRepositoryKind:
		kind = pb.FluxObjectKind_KindHelmRepository
	case sourcev1.BucketKind:
		kind = pb.FluxObjectKind_KindBucket
	case sourcev1.OCIRepositoryKind:
		kind = pb.FluxObjectKind_KindOCIRepository
	}

	inv, err := getKustomizeInventory(kustomization)
	if err != nil {
		return nil, fmt.Errorf("coverting kustomization to proto: %w", err)
	}

	var sourceNamespace string
	if kustomization.Spec.SourceRef.Namespace != "" {
		sourceNamespace = kustomization.Spec.SourceRef.Namespace
	} else {
		sourceNamespace = kustomization.Namespace
	}

	version, _ := kustomization.GroupVersionKind().ToAPIVersionAndKind()

	dependsOn := []*pb.NamespacedObjectReference{}

	for _, v := range kustomization.Spec.DependsOn {
		dependsOn = append(dependsOn, &pb.NamespacedObjectReference{
			Name:      v.Name,
			Namespace: v.Namespace,
		})
	}

	return &pb.Kustomization{
		Name:      kustomization.Name,
		Namespace: kustomization.Namespace,
		Path:      kustomization.Spec.Path,
		SourceRef: &pb.FluxObjectRef{
			Kind:      kind,
			Name:      kustomization.Spec.SourceRef.Name,
			Namespace: sourceNamespace,
		},
		Interval:              durationToInterval(kustomization.Spec.Interval),
		Conditions:            mapConditions(kustomization.Status.Conditions),
		LastAppliedRevision:   kustomization.Status.LastAppliedRevision,
		LastAttemptedRevision: kustomization.Status.LastAttemptedRevision,
		Inventory:             inv,
		Suspended:             kustomization.Spec.Suspend,
		ClusterName:           clusterName,
		ApiVersion:            version,
		Tenant:                tenant,
		Uid:                   string(kustomization.GetUID()),
		DependsOn:             dependsOn,
	}, nil
}

func getKustomizeInventory(kustomization *kustomizev1.Kustomization) ([]*pb.GroupVersionKind, error) {
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

		gvkId := strings.Join([]string{objMeta.GroupKind.Group, entry.Version, objMeta.GroupKind.Kind}, "_")

		if !found[gvkId] {
			found[gvkId] = true

			gvk = append(gvk, &pb.GroupVersionKind{
				Group:   objMeta.GroupKind.Group,
				Version: entry.Version,
				Kind:    objMeta.GroupKind.Kind,
			})
		}
	}

	return gvk, nil
}
