package types

import (
	"time"

	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ProtoToHelmRepository(hr *pb.HelmRepository) v1beta1.HelmRepository {
	return v1beta1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1beta1.HelmRepositoryKind,
			APIVersion: v1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hr.Name,
			Namespace: hr.Namespace,
			Labels:    getGitopsLabelMap(hr.Name),
		},
		Spec: v1beta1.HelmRepositorySpec{
			URL:      hr.Url,
			Interval: metav1.Duration{Duration: time.Minute * 1},
			Timeout:  &metav1.Duration{Duration: time.Minute * 1},
		},
		Status: v1beta1.HelmRepositoryStatus{},
	}
}

func HelmRepositoryToProto(helmRepository *v1beta1.HelmRepository) *pb.HelmRepository {
	return &pb.HelmRepository{
		Name:       helmRepository.Name,
		Namespace:  helmRepository.Namespace,
		Url:        helmRepository.Spec.URL,
		Interval:   durationToInterval(helmRepository.Spec.Interval),
		Conditions: mapConditions(helmRepository.Status.Conditions),
		Suspended:  helmRepository.Spec.Suspend,
	}
}
