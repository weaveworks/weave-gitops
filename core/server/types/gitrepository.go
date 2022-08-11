package types

import (
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

func GitRepositoryToProto(repository *sourcev1.GitRepository, clusterName string, tenant string) *pb.GitRepository {
	if repository == nil {
		return nil
	}

	gr := &pb.GitRepository{
		Name:          repository.Name,
		Namespace:     repository.Namespace,
		Url:           repository.Spec.URL,
		Interval:      durationToInterval(repository.Spec.Interval),
		Conditions:    mapConditions(repository.Status.Conditions),
		Suspended:     repository.Spec.Suspend,
		LastUpdatedAt: lastUpdatedAt(repository),
		ClusterName:   clusterName,
		ApiVersion:    repository.APIVersion,
		Tenant: 	   tenant,
	}

	if repository.Spec.Reference != nil {
		gr.Reference = &pb.GitRepositoryRef{
			Branch: repository.Spec.Reference.Branch,
			Tag:    repository.Spec.Reference.Tag,
			Semver: repository.Spec.Reference.SemVer,
			Commit: repository.Spec.Reference.Commit,
		}
	}

	if repository.Spec.SecretRef != nil {
		gr.SecretRef = repository.Spec.SecretRef.Name
	}

	return gr
}
