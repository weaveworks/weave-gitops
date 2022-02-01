package types

import (
	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
)

func getGitopsLabelMap(appName string) map[string]string {
	labels := map[string]string{
		ManagedByLabel: managedByWeaveGitops,
		CreatedByLabel: createdBySourceController,
	}

	if appName != "" {
		labels[PartOfLabel] = appName
	}

	return labels
}

func getSourceKind(kind string) pb.SourceRef_Kind {
	switch kind {
	case v1beta1.GitRepositoryKind:
		return pb.SourceRef_GitRepository
	case v1beta1.HelmRepositoryKind:
		return pb.SourceRef_HelmRepository
	case v1beta1.BucketKind:
		return pb.SourceRef_Bucket
	default:
		return -1
	}
}
