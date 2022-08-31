package types

import (
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

func HelmChartToProto(helmchart *sourcev1.HelmChart, clusterName string, tenant string) *pb.HelmChart {
	return &pb.HelmChart{
		Name:      helmchart.Name,
		Namespace: helmchart.Namespace,
		SourceRef: &pb.FluxObjectRef{
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
		ApiVersion:    helmchart.APIVersion,
		Tenant:        tenant,
		Uid:           string(helmchart.GetUID()),
	}
}
