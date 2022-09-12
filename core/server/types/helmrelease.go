package types

import (
	"fmt"

	"github.com/fluxcd/helm-controller/api/v2beta1"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

type HelmReleaseStorage struct {
	Name     string `json:"name,omitempty"`
	Manifest string `json:"manifest,omitempty"`
}

func HelmReleaseToProto(helmrelease *v2beta1.HelmRelease, clusterName string, inventory []*pb.GroupVersionKind, tenant string) *pb.HelmRelease {
	var chartInterval *pb.Interval

	if helmrelease.Spec.Chart.Spec.Interval != nil {
		chartInterval = durationToInterval(*helmrelease.Spec.Chart.Spec.Interval)
	}

	var sourceNamespace string
	if helmrelease.Spec.Chart.Spec.SourceRef.Namespace != "" {
		sourceNamespace = helmrelease.Spec.Chart.Spec.SourceRef.Namespace
	} else {
		sourceNamespace = helmrelease.Namespace
	}

	version, _ := helmrelease.GroupVersionKind().ToAPIVersionAndKind()

	dependsOn := []*pb.NamespacedObjectReference{}

	for _, v := range helmrelease.Spec.DependsOn {
		dependsOn = append(dependsOn, &pb.NamespacedObjectReference{
			Name:      v.Name,
			Namespace: v.Namespace,
		})
	}

	return &pb.HelmRelease{
		Name:        helmrelease.Name,
		ReleaseName: helmrelease.Spec.ReleaseName,
		Namespace:   helmrelease.Namespace,
		Interval:    durationToInterval(helmrelease.Spec.Interval),
		HelmChart: &pb.HelmChart{
			Chart:     helmrelease.Spec.Chart.Spec.Chart,
			Version:   helmrelease.Spec.Chart.Spec.Version,
			Name:      fmt.Sprintf("%s-%s", helmrelease.Namespace, helmrelease.Name),
			Namespace: sourceNamespace,
			Interval:  chartInterval,
			SourceRef: &pb.FluxObjectRef{
				Namespace: sourceNamespace,
				Name:      helmrelease.Spec.Chart.Spec.SourceRef.Name,
				Kind:      getSourceKind(helmrelease.Spec.Chart.Spec.SourceRef.Kind),
			},
		},
		Inventory:             inventory,
		Conditions:            mapConditions(helmrelease.Status.Conditions),
		Suspended:             helmrelease.Spec.Suspend,
		HelmChartName:         helmrelease.Status.HelmChart,
		ClusterName:           clusterName,
		LastAppliedRevision:   helmrelease.Status.LastAppliedRevision,
		LastAttemptedRevision: helmrelease.Status.LastAttemptedRevision,
		ApiVersion:            version,
		Tenant:                tenant,
		Uid:                   string(helmrelease.GetUID()),
		DependsOn:             dependsOn,
	}
}
