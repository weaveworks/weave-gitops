package types

import (
	"time"

	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ProtoToHelmChart(helmChartReq *pb.AddHelmChartReq) v1beta1.HelmChart {
	labels := map[string]string{
		ManagedByLabel: managedByWeaveGitops,
		CreatedByLabel: createdBySourceController,
	}

	if helmChartReq.AppName != "" {
		labels[PartOfLabel] = helmChartReq.AppName
	}

	return v1beta1.HelmChart{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1beta1.HelmChartKind,
			APIVersion: v1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      helmChartReq.Name,
			Namespace: helmChartReq.Namespace,
			Labels:    labels,
		},
		Spec: v1beta1.HelmChartSpec{
			Chart:   helmChartReq.Chart,
			Version: helmChartReq.Version,
			SourceRef: v1beta1.LocalHelmChartSourceReference{
				Kind: helmChartReq.SourceRef.Kind.String(),
				Name: helmChartReq.SourceRef.Name,
			},
			Interval: metav1.Duration{Duration: time.Minute * 1},
		},
		Status: v1beta1.HelmChartStatus{},
	}
}

func HelmChartToProto(helmchart *v1beta1.HelmChart) *pb.HelmChart {
	var kind pb.SourceRef_Kind

	switch helmchart.Spec.SourceRef.Kind {
	case v1beta1.GitRepositoryKind:
		kind = pb.SourceRef_GitRepository
	case v1beta1.HelmRepositoryKind:
		kind = pb.SourceRef_HelmRepository
	case v1beta1.BucketKind:
		kind = pb.SourceRef_Bucket
	}

	hr := &pb.HelmChart{
		Name:      helmchart.Name,
		Namespace: helmchart.Namespace,
		SourceRef: &pb.SourceRef{
			Kind: kind,
			Name: helmchart.Name,
		},
		Chart:   helmchart.Spec.Chart,
		Version: helmchart.Spec.Version,
		Interval: &pb.Interval{
			Minutes: 1,
		},
	}

	return hr
}
