package types

import (
	"github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/api/v1alpha2"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AppCustomResourceToProto(a *v1alpha2.Application) *pb.App {
	return &pb.App{
		Name:        a.ObjectMeta.Name,
		Namespace:   a.ObjectMeta.Namespace,
		DisplayName: a.Spec.DisplayName,
		Description: a.Spec.Description,
	}
}

func AppAddProtoToCustomResource(msg *pb.AddAppRequest) *v1alpha2.Application {
	return &v1alpha2.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.ApplicationKind,
			APIVersion: "wego.weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      msg.Name,
			Namespace: msg.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/part-of":    msg.Name,
				"app.kubernetes.io/managed-by": "weave-gitops",
				"app.kubernetes.io/created-by": "kustomize-controller",
			},
		},
		Spec: v1alpha2.ApplicationSpec{
			Description: msg.Description,
			DisplayName: msg.DisplayName,
		},
		Status: v1alpha2.ApplicationStatus{},
	}
}
