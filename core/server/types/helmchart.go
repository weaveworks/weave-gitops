package types

import (
	"time"

	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ProtoToHelmChart(chart *pb.HelmChart) v1beta1.HelmChart {
	return v1beta1.HelmChart{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1beta1.HelmChartKind,
			APIVersion: v1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      chart.Name,
			Namespace: chart.Namespace,
			Labels:    getGitopsLabelMap(chart.Name),
		},
		Spec: v1beta1.HelmChartSpec{
			Chart:   chart.Chart,
			Version: chart.Version,
			SourceRef: v1beta1.LocalHelmChartSourceReference{
				Kind: chart.SourceRef.Kind.String(),
				Name: chart.SourceRef.Name,
			},
			Interval: metav1.Duration{Duration: time.Minute * 1},
		},
		Status: v1beta1.HelmChartStatus{},
	}
}

func HelmChartToProto(helmchart *v1beta1.HelmChart) *pb.HelmChart {
	return &pb.HelmChart{
		Name:      helmchart.Name,
		Namespace: helmchart.Namespace,
		SourceRef: &pb.SourceRef{
			Kind: getSourceKind(helmchart.Spec.SourceRef.Kind),
			Name: helmchart.Name,
		},
		Chart:         helmchart.Spec.Chart,
		Version:       helmchart.Spec.Version,
		Interval:      durationToInterval(helmchart.Spec.Interval),
		Conditions:    mapConditions(helmchart.Status.Conditions),
		Suspended:     helmchart.Spec.Suspend,
		LastUpdatedAt: lastUpdatedAt(helmchart),
	}
}
