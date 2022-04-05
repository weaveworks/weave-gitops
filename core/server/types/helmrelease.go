package types

import (
	"github.com/fluxcd/helm-controller/api/v2beta1"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

type HelmReleaseStorage struct {
	Name     string `json:"name,omitempty"`
	Manifest string `json:"manifest,omitempty"`
}

func HelmReleaseToProto(helmrelease *v2beta1.HelmRelease, clusterName string, inventory []*pb.GroupVersionKind) *pb.HelmRelease {
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
		Inventory:     inventory,
		Conditions:    mapConditions(helmrelease.Status.Conditions),
		Suspended:     helmrelease.Spec.Suspend,
		ClusterName:   clusterName,
		LastUpdatedAt: lastUpdatedAt(helmrelease),
	}
}
