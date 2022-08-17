package types_test

import (
	"testing"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/weaveworks/weave-gitops/core/server/types"
	api "github.com/weaveworks/weave-gitops/pkg/api/core"
)

func TestGitRepositoryToProto(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		name        string
		clusterName string
		in          *sourcev1.GitRepository
		expected    *api.GitRepository
	}{
		{
			name:        "nil in, nil out",
			clusterName: "foo",
		},
		{
			name:        "empty in, empty out",
			clusterName: "foo",
			in:          &sourcev1.GitRepository{},
			expected: &api.GitRepository{
				ClusterName: "foo",
				Conditions:  []*api.Condition{},
				Interval:    &api.Interval{},
			},
		},
		{
			name:        "complete in, complete out",
			clusterName: "foo",
			in: &sourcev1.GitRepository{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "api-version",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gr01",
					Namespace: "ns01",
				},
				Spec: sourcev1.GitRepositorySpec{
					URL: "some-url",
					Interval: metav1.Duration{
						Duration: time.Second * 13942,
					},
					Suspend: true,
					Reference: &sourcev1.GitRepositoryRef{
						Branch: "branch",
						Tag:    "tag",
						SemVer: "semver",
						Commit: "commit",
					},
					SecretRef: &meta.LocalObjectReference{
						Name: "oh-so-secret",
					},
				},
				Status: sourcev1.GitRepositoryStatus{
					Conditions: []metav1.Condition{{
						Type:               sourcev1.StorageOperationFailedCondition,
						Status:             metav1.ConditionTrue,
						Reason:             sourcev1.StatOperationFailedReason,
						Message:            "foobar",
						LastTransitionTime: metav1.NewTime(time.Date(2021, time.January, 14, 13, 12, 11, 10, time.FixedZone("fz", 5*3600))),
					}},
					Artifact: &sourcev1.Artifact{
						LastUpdateTime: metav1.NewTime(time.Date(2022, time.April, 9, 8, 7, 6, 5, time.FixedZone("fz", 8*3600))),
					},
				},
			},
			expected: &api.GitRepository{
				Name:      "gr01",
				Namespace: "ns01",
				Url:       "some-url",
				Interval: &api.Interval{
					Hours:   3,
					Minutes: 52,
					Seconds: 22,
				},
				Conditions: []*api.Condition{{
					Type:      sourcev1.StorageOperationFailedCondition,
					Status:    string(metav1.ConditionTrue),
					Reason:    sourcev1.StatOperationFailedReason,
					Message:   "foobar",
					Timestamp: "2021-01-14T13:12:11+05:00",
				}},
				Suspended:     true,
				LastUpdatedAt: "2022-04-09T08:07:06+08:00",
				ClusterName:   "foo",
				ApiVersion:    "api-version",
				Reference: &api.GitRepositoryRef{
					Branch: "branch",
					Tag:    "tag",
					Semver: "semver",
					Commit: "commit",
				},
				SecretRef: "oh-so-secret",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := types.GitRepositoryToProto(tt.in, tt.clusterName, "")

			g.Expect(out).To(Equal(tt.expected))
		})
	}
}
