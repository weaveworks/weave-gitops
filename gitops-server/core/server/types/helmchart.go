package types

import (
	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/gitops-server/pkg/api/core"
)

func HelmChartToProto(helmchart *v1beta1.HelmChart, clusterName string) *pb.HelmChart {
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
		ClusterName:   clusterName,
	}
}
