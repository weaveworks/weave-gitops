package types

import (
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ProtoToKustomization(kustomization *pb.AddKustomizationReq) v1beta2.Kustomization {
	labels := getGitopsLabelMap(kustomization.AppName)

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

func KustomizationToProto(kustomization *v1beta2.Kustomization) *pb.Kustomization {
	var kind pb.SourceRef_Kind

	switch kustomization.Spec.SourceRef.Kind {
	case v1beta1.GitRepositoryKind:
		kind = pb.SourceRef_GitRepository
	case v1beta1.HelmRepositoryKind:
		kind = pb.SourceRef_HelmRepository
	case v1beta1.BucketKind:
		kind = pb.SourceRef_Bucket
	}

	return &pb.Kustomization{
		Name:      kustomization.Name,
		Namespace: kustomization.Namespace,
		Path:      kustomization.Spec.Path,
		SourceRef: &pb.SourceRef{
			Kind: kind,
			Name: kustomization.Spec.SourceRef.Name,
		},
		Interval:                nil,
		Conditions:              mapConditions(kustomization.Status.Conditions),
		LastAppliedRevision:     kustomization.Status.LastAppliedRevision,
		LastAttemptedRevision:   kustomization.Status.LastAttemptedRevision,
		LastHandledReconciledAt: kustomization.Status.LastHandledReconcileAt,
	}
}
