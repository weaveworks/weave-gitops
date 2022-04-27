package types

import (
	pb "github.com/weaveworks/weave-gitops/gitops-server/pkg/api/core"
	corev1 "k8s.io/api/core/v1"
)

func NamespaceToProto(ns corev1.Namespace, clusterName string) *pb.Namespace {
	return &pb.Namespace{
		ClusterName: clusterName,
		Name:        ns.GetName(),
		Status:      ns.Status.String(),
		Annotations: ns.GetAnnotations(),
		Labels:      ns.GetLabels(),
	}
}
