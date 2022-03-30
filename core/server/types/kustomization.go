package types

import (
	"fmt"
	"strings"

	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cli-utils/pkg/object"
)

func ProtoToKustomization(kustomization *pb.Kustomization) v1beta2.Kustomization {
	labels := getGitopsLabelMap(kustomization.Name)

	return v1beta2.Kustomization{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1beta2.KustomizationKind,
			APIVersion: v1beta2.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kustomization.Name,
			Namespace: kustomization.Namespace,
			Labels:    labels,
		},
		Spec: v1beta2.KustomizationSpec{
			Path: kustomization.Path,
			//Interval: intervalDuration(kustomization.Interval),
			SourceRef: v1beta2.CrossNamespaceSourceReference{
				Kind: kustomization.SourceRef.Kind.String(),
				Name: kustomization.SourceRef.Name,
			},
		},
		Status: v1beta2.KustomizationStatus{},
	}
}

func KustomizationToProto(kustomization *v1beta2.Kustomization) (*pb.Kustomization, error) {
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
