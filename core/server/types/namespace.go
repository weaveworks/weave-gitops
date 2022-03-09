package types

import (
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	corev1 "k8s.io/api/core/v1"
)

func NamespaceToProto(ns corev1.Namespace) *pb.Namespace {
	return &pb.Namespace{
		Name:        ns.GetName(),
		Status:      ns.Status.String(),
		Annotations: ns.GetAnnotations(),
		Labels:      ns.GetLabels(),
	}
}
