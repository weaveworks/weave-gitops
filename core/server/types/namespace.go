package types

import (
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
)

func NamespaceToProto(ns corev1.Namespace, authRes *authorizationv1.SelfSubjectRulesReview) *pb.Namespace {
	p := &pb.Namespace{
		Name:          ns.GetName(),
		Status:        ns.Status.String(),
		Annotations:   ns.GetAnnotations(),
		Labels:        ns.GetLabels(),
		ResourceRules: []*pb.ResourceRule{},
	}

	for _, r := range authRes.Status.ResourceRules {
		p.ResourceRules = append(p.ResourceRules, &pb.ResourceRule{
			Verbs:         r.Verbs,
			ApiGroups:     r.APIGroups,
			Resources:     r.Resources,
			ResourceNames: r.ResourceNames,
		})
	}

	return p
}
