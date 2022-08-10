package types

import (
	"time"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getSourceKind(kind string) pb.FluxObjectKind {
	switch kind {
	case sourcev1.GitRepositoryKind:
		return pb.FluxObjectKind_KindGitRepository
	case sourcev1.HelmRepositoryKind:
		return pb.FluxObjectKind_KindHelmRepository
	case sourcev1.BucketKind:
		return pb.FluxObjectKind_KindBucket
	case sourcev1.OCIRepositoryKind:
		return pb.FluxObjectKind_KindOCIRepository
	default:
		return -1
	}
}

func mapConditions(conditions []metav1.Condition) []*pb.Condition {
	out := []*pb.Condition{}

	for _, c := range conditions {
		out = append(out, &pb.Condition{
			Type:      c.Type,
			Status:    string(c.Status),
			Reason:    c.Reason,
			Message:   c.Message,
			Timestamp: c.LastTransitionTime.Format(time.RFC3339),
		})
	}

	return out
}

func lastUpdatedAt(obj interface{}) string {
	switch s := obj.(type) {
	case *sourcev1.GitRepository:
		if s.Status.Artifact != nil {
			return s.Status.Artifact.LastUpdateTime.Format(time.RFC3339)
		}
	case *sourcev1.Bucket:
		if s.Status.Artifact != nil {
			return s.Status.Artifact.LastUpdateTime.Format(time.RFC3339)
		}
	case *sourcev1.HelmChart:
		if s.Status.Artifact != nil {
			return s.Status.Artifact.LastUpdateTime.Format(time.RFC3339)
		}
	case *sourcev1.HelmRepository:
		if s.Status.Artifact != nil {
			return s.Status.Artifact.LastUpdateTime.Format(time.RFC3339)
		}
	case *sourcev1.OCIRepository:
		if s.Status.Artifact != nil {
			return s.Status.Artifact.LastUpdateTime.Format(time.RFC3339)
		}
	}

	return ""
}

func durationToInterval(duration metav1.Duration) *pb.Interval {
	return &pb.Interval{
		Hours:   int64(duration.Hours()),
		Minutes: int64(duration.Minutes()) % 60,
		Seconds: int64(duration.Seconds()) % 60,
	}
}
