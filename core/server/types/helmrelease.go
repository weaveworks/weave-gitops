package types

import (
	"time"

	"github.com/fluxcd/helm-controller/api/v2beta1"

	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HelmReleaseStorage struct {
	Name     string `json:"name,omitempty"`
	Manifest string `json:"manifest,omitempty"`
}

func ProtoToHelmRelease(hr *pb.HelmRelease) v2beta1.HelmRelease {
	return v2beta1.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       v2beta1.HelmReleaseKind,
			APIVersion: v1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hr.Name,
			Namespace: hr.Namespace,
			Labels:    getGitopsLabelMap(hr.Name),
		},
		Spec: v2beta1.HelmReleaseSpec{
			Chart: v2beta1.HelmChartTemplate{
				Spec: v2beta1.HelmChartTemplateSpec{
					Chart:   hr.HelmChart.Chart,
					Version: hr.HelmChart.Version,
					SourceRef: v2beta1.CrossNamespaceObjectReference{
						Kind: hr.HelmChart.SourceRef.Kind.String(),
						Name: hr.HelmChart.SourceRef.Name,
					},
					Interval: &metav1.Duration{Duration: time.Minute * 1},
				},
			},
		},
		Status: v2beta1.HelmReleaseStatus{
			ObservedGeneration: -1,
		},
	}
}

func HelmReleaseToProto(helmrelease *v2beta1.HelmRelease, inventory []*pb.GroupVersionKind) *pb.HelmRelease {
	var chartInterval *pb.Interval

	if helmrelease.Spec.Chart.Spec.Interval != nil {
		chartInterval = durationToInterval(*helmrelease.Spec.Chart.Spec.Interval)
	}

	return &pb.HelmRelease{
		Name:        helmrelease.Name,
		ReleaseName: helmrelease.Spec.ReleaseName,
		Namespace:   helmrelease.Namespace,
		Interval:    durationToInterval(helmrelease.Spec.Interval),
		HelmChart: &pb.HelmChart{
			Chart:     helmrelease.Spec.Chart.Spec.Chart,
			Namespace: helmrelease.Spec.Chart.Spec.SourceRef.Namespace,
			Name:      helmrelease.Spec.Chart.Spec.SourceRef.Name,
			Version:   helmrelease.Spec.Chart.Spec.Version,
			Interval:  chartInterval,
			SourceRef: &pb.SourceRef{
				Kind: getSourceKind(helmrelease.Spec.Chart.Spec.SourceRef.Kind),
			},
		},
		Inventory:  inventory,
		Conditions: mapConditions(helmrelease.Status.Conditions),
		Suspended:  helmrelease.Spec.Suspend,
	}
}
