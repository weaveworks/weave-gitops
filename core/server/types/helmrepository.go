package types

import (
	"time"

	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ProtoToHelmRepository(repositoryReq *pb.AddHelmRepositoryReq) v1beta1.HelmRepository {

	labels := map[string]string{
		"app.kubernetes.io/managed-by": managedByWeaveGitops,
		"app.kubernetes.io/created-by": createdBySourceController,
	}

	if repositoryReq.AppName != "" {
		labels["app.kubernetes.io/part-of"] = repositoryReq.AppName
	}

	return v1beta1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1beta1.HelmRepositoryKind,
			APIVersion: v1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      repositoryReq.Name,
			Namespace: repositoryReq.Namespace,
			Labels:    labels,
		},
		Spec: v1beta1.HelmRepositorySpec{
			URL:      repositoryReq.Url,
			Interval: metav1.Duration{Duration: time.Minute * 1},
			Timeout:  &metav1.Duration{Duration: time.Minute * 1},
		},
		Status: v1beta1.HelmRepositoryStatus{},
	}
}

func HelmRepositoryToProto(repository *v1beta1.HelmRepository) *pb.HelmRepository {
	hr := &pb.HelmRepository{
		Name:      repository.Name,
		Namespace: repository.Namespace,
		Url:       repository.Spec.URL,
		Interval: &pb.Interval{
			Minutes: 1,
		},
	}

	return hr
}
