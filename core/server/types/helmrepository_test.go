package types

import (
	"testing"
	"time"

	"github.com/fluxcd/source-controller/api/v1beta2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHelmRepository(t *testing.T) {
	g := NewGomegaWithT(t)

	format.UseStringerRepresentation = true // Makes the representation more compact

	d, err := time.ParseDuration("1h2m3s")
	g.Expect(err).NotTo(HaveOccurred())

	tests := []struct {
		name        string
		clusterName string
		state       v1beta2.HelmRepository
		result      *pb.HelmRepository
	}{
		{
			"empty",
			"Default",
			v1beta2.HelmRepository{},
			&pb.HelmRepository{
				Interval:       &pb.Interval{},
				Conditions:     []*pb.Condition{},
				ClusterName:    "Default",
				RepositoryType: pb.HelmRepositoryType_Default,
			},
		},
		{
			"chart-repository",
			"Default",
			v1beta2.HelmRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-chart",
					Namespace: "namespace-of-all-objects",
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "some-version",
				},
				Spec: v1beta2.HelmRepositorySpec{
					URL:      "http://some-domain.example",
					Type:     "Default",
					Interval: metav1.Duration{Duration: d},
					Suspend:  true,
				},
				Status: v1beta2.HelmRepositoryStatus{
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionTrue,
							Reason:             "InstallSucceeded",
							Message:            "Release reconciliation succeeded",
							LastTransitionTime: metav1.Time{Time: time.Date(2022, time.January, 1, 1, 1, 1, 0, time.UTC)},
						},
					},
					Artifact: &v1beta2.Artifact{
						Checksum:       "1234",
						LastUpdateTime: metav1.Time{Time: time.Date(2023, time.January, 1, 1, 1, 1, 0, time.UTC)},
					},
				},
			},
			&pb.HelmRepository{
				Name:      "some-chart",
				Namespace: "namespace-of-all-objects",
				Url:       "http://some-domain.example",
				Interval:  &pb.Interval{Hours: 1, Minutes: 2, Seconds: 3},
				Conditions: []*pb.Condition{
					{
						Type:      "Ready",
						Status:    "True",
						Reason:    "InstallSucceeded",
						Message:   "Release reconciliation succeeded",
						Timestamp: "2022-01-01T01:01:01Z",
					},
				},
				ClusterName:    "Default",
				LastUpdatedAt:  "2023-01-01T01:01:01Z",
				ApiVersion:     "some-version",
				Suspended:      true,
				RepositoryType: pb.HelmRepositoryType_Default,
			},
		},
		{
			"oci-repository",
			"Default",
			v1beta2.HelmRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-chart",
					Namespace: "namespace-of-all-objects",
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "some-version",
				},
				Spec: v1beta2.HelmRepositorySpec{
					URL:      "oci://some-domain.example",
					Type:     "oci",
					Interval: metav1.Duration{Duration: d},
					Suspend:  true,
				},
				Status: v1beta2.HelmRepositoryStatus{
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionTrue,
							Reason:             "InstallSucceeded",
							Message:            "Release reconciliation succeeded",
							LastTransitionTime: metav1.Time{Time: time.Date(2022, time.January, 1, 1, 1, 1, 0, time.UTC)},
						},
					},
					// OCI repositories don't have artifacts - the artifact is per-chart
				},
			},
			&pb.HelmRepository{
				Name:      "some-chart",
				Namespace: "namespace-of-all-objects",
				Url:       "oci://some-domain.example",
				Interval:  &pb.Interval{Hours: 1, Minutes: 2, Seconds: 3},
				Conditions: []*pb.Condition{
					{
						Type:      "Ready",
						Status:    "True",
						Reason:    "InstallSucceeded",
						Message:   "Release reconciliation succeeded",
						Timestamp: "2022-01-01T01:01:01Z",
					},
				},
				ClusterName:    "Default",
				ApiVersion:     "some-version",
				Suspended:      true,
				RepositoryType: pb.HelmRepositoryType_OCI,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := HelmRepositoryToProto(&tt.state, tt.clusterName, "")

			g.Expect(res).To(Equal(tt.result))
		})
	}
}
