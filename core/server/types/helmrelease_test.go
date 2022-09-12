package types

import (
	"testing"
	"time"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHelmRelease(t *testing.T) {
	g := NewGomegaWithT(t)

	format.UseStringerRepresentation = true // Makes the representation more compact

	d123, err := time.ParseDuration("1h2m3s")
	g.Expect(err).NotTo(HaveOccurred())
	d321, err := time.ParseDuration("3h2m1s")
	g.Expect(err).NotTo(HaveOccurred())

	tests := []struct {
		name        string
		clusterName string
		state       v2beta1.HelmRelease
		result      *pb.HelmRelease
	}{
		{
			"empty",
			"Default",
			v2beta1.HelmRelease{},
			&pb.HelmRelease{
				HelmChart: &pb.HelmChart{
					Name: "-",
					SourceRef: &pb.FluxObjectRef{
						Kind: -1, // This is invalid?
					},
				},
				Interval:    &pb.Interval{},
				Inventory:   []*pb.GroupVersionKind{},
				Conditions:  []*pb.Condition{},
				ClusterName: "Default",
				DependsOn:   []*pb.NamespacedObjectReference{},
			},
		},
		{
			"same-ns",
			"Default",
			v2beta1.HelmRelease{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-chart",
					Namespace: "namespace-of-all-objects",
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "some-version",
				},
				Spec: v2beta1.HelmReleaseSpec{
					Chart: v2beta1.HelmChartTemplate{
						Spec: v2beta1.HelmChartTemplateSpec{
							SourceRef: v2beta1.CrossNamespaceObjectReference{
								Name: "source-object",
								Kind: "GitRepository",
							},
						},
					},
				},
			},
			&pb.HelmRelease{
				Name:      "some-chart",
				Namespace: "namespace-of-all-objects",
				HelmChart: &pb.HelmChart{
					Name:      "namespace-of-all-objects-some-chart",
					Namespace: "namespace-of-all-objects",
					SourceRef: &pb.FluxObjectRef{
						Name:      "source-object",
						Namespace: "namespace-of-all-objects",
						Kind:      pb.FluxObjectKind_KindGitRepository,
					},
				},
				Interval:    &pb.Interval{},
				Inventory:   []*pb.GroupVersionKind{},
				Conditions:  []*pb.Condition{},
				ClusterName: "Default",
				ApiVersion:  "some-version",
				DependsOn:   []*pb.NamespacedObjectReference{},
			},
		},
		{
			"cross-ns",
			"Default",
			v2beta1.HelmRelease{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-chart",
					Namespace: "namespace-of-object",
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "some-version",
				},
				Spec: v2beta1.HelmReleaseSpec{
					Interval: metav1.Duration{Duration: d123},
					Chart: v2beta1.HelmChartTemplate{
						Spec: v2beta1.HelmChartTemplateSpec{
							Chart:    "chart-name",
							Version:  "semver-version",
							Interval: &metav1.Duration{Duration: d321},
							SourceRef: v2beta1.CrossNamespaceObjectReference{
								Kind:      "HelmRepository",
								Name:      "some-helm-repository",
								Namespace: "namespace-of-source",
							},
						},
					},
				},
				Status: v2beta1.HelmReleaseStatus{
					LastAppliedRevision:   "1.0",
					LastAttemptedRevision: "2.0",
				},
			},
			&pb.HelmRelease{
				Name:      "some-chart",
				Namespace: "namespace-of-object",
				HelmChart: &pb.HelmChart{
					Name:      "namespace-of-object-some-chart",
					Namespace: "namespace-of-source",
					SourceRef: &pb.FluxObjectRef{
						Name:      "some-helm-repository",
						Namespace: "namespace-of-source",
						Kind:      pb.FluxObjectKind_KindHelmRepository,
					},
					Chart:    "chart-name",
					Version:  "semver-version",
					Interval: &pb.Interval{Hours: 3, Minutes: 2, Seconds: 1},
				},
				Interval:              &pb.Interval{Hours: 1, Minutes: 2, Seconds: 3},
				Inventory:             []*pb.GroupVersionKind{},
				Conditions:            []*pb.Condition{},
				LastAppliedRevision:   "1.0",
				LastAttemptedRevision: "2.0",
				ClusterName:           "Default",
				ApiVersion:            "some-version",
				DependsOn:             []*pb.NamespacedObjectReference{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := HelmReleaseToProto(&tt.state, tt.clusterName, []*pb.GroupVersionKind{}, "")

			g.Expect(res).To(Equal(tt.result))
		})
	}
}
