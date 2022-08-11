package types

import (
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

func OCIRepositoryToProto(ociRepository *sourcev1.OCIRepository, clusterName string, tenant string) *pb.OCIRepository {
	return &pb.OCIRepository{
		Name:          ociRepository.Name,
		Namespace:     ociRepository.Namespace,
		Url:           ociRepository.Spec.URL,
		Interval:      durationToInterval(ociRepository.Spec.Interval),
		Conditions:    mapConditions(ociRepository.Status.Conditions),
		Suspended:     ociRepository.Spec.Suspend,
		LastUpdatedAt: lastUpdatedAt(ociRepository),
		ClusterName:   clusterName,
		ApiVersion:    ociRepository.APIVersion,
		Tenant:        tenant,
	}
}
