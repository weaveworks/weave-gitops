package server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"google.golang.org/protobuf/testing/protocmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetPolicyViolation(t *testing.T) {
	tests := []struct {
		name         string
		ViolationId  string
		clusterState []runtime.Object
		clusterName  string
		err          error
		expected     *pb.GetPolicyValidationResponse
	}{
		{
			name:        "get policy violation",
			ViolationId: "66101548-12c1-4f79-a09a-a12979903fba",
			clusterState: []runtime.Object{
				makeEvent(t),
			},
			clusterName: "Default",
			expected: &pb.GetPolicyValidationResponse{
				Validation: &pb.PolicyValidation{
					Id:              "66101548-12c1-4f79-a09a-a12979903fba",
					Name:            "Missing app Label",
					PolicyId:        "weave.policies.missing-app-label",
					ClusterId:       "cluster-1",
					Category:        "Access Control",
					Severity:        "high",
					CreatedAt:       "0001-01-01T00:00:00Z",
					Message:         "Policy event",
					Entity:          "my-deployment",
					Namespace:       "default",
					Description:     "Missing app label",
					HowToSolve:      "how_to_solve",
					ViolatingEntity: `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"nginx-deployment","namespace":"default","uid":"af912668-957b-46d4-bc7a-51e6994cba56"},"spec":{"template":{"spec":{"containers":[{"image":"nginx:latest","imagePullPolicy":"Always","name":"nginx","ports":[{"containerPort":80,"protocol":"TCP"}]}]}}}}`,
					ClusterName:     "Default",
					Occurrences: []*pb.PolicyValidationOccurrence{
						{
							Message: "occurrence details",
						},
					},
				},
			},
			err: nil,
		},
		{
			name:        "policy violation doesn't exist",
			ViolationId: "invalid-id",
			clusterState: []runtime.Object{
				makeEvent(t),
			},
			clusterName: "Default",
			err:         errors.New("no policy violation found with id invalid-id and cluster: Default"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fakeCl := fake.NewClientBuilder().WithRuntimeObjects(tt.clusterState...).Build()
			cfg := makeServerConfig(fakeCl, t, "")
			s := makeServer(cfg, t)

			policyViolation, err := s.GetPolicyValidation(context.Background(), &pb.GetPolicyValidationRequest{
				ValidationId: tt.ViolationId,
				ClusterName:  tt.clusterName,
			})
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to get policy violation:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("unexpected error while getting policy:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, policyViolation, protocmp.Transform()); diff != "" {
					t.Fatalf("policy violation didn't match expected:\n%s", diff)
				}
			}
		})
	}
}

func TestListPolicyValidations(t *testing.T) {
	tests := []struct {
		name         string
		clusterState []runtime.Object
		events       []*corev1.Event
		err          error
		expected     *pb.ListPolicyValidationsResponse
		clusterName  string
		appName      string
		appKind      string
		namespace    string
	}{
		{
			name: "list policy violations",
			clusterState: []runtime.Object{
				makeEvent(t),
				makeEvent(t, func(e *corev1.Event) {
					e.ObjectMeta.Name = "Missing Owner Label - fake-event-2"
					e.InvolvedObject.Namespace = "weave-system"
					e.ObjectMeta.Namespace = "weave-system"
					e.Annotations["policy_name"] = "Missing Owner Label"
					e.Annotations["policy_id"] = "weave.policies.missing-app-label"
					e.Labels["pac.weave.works/id"] = "56701548-12c1-4f79-a09a-a12979903"
				}),
			},
			expected: &pb.ListPolicyValidationsResponse{
				Violations: []*pb.PolicyValidation{
					{
						Id:          "66101548-12c1-4f79-a09a-a12979903fba",
						Name:        "Missing app Label",
						PolicyId:    "weave.policies.missing-app-label",
						ClusterId:   "cluster-1",
						Category:    "Access Control",
						Severity:    "high",
						CreatedAt:   "0001-01-01T00:00:00Z",
						Message:     "Policy event",
						Entity:      "my-deployment",
						Namespace:   "default",
						ClusterName: "Default",
					},
					{
						Id:          "56701548-12c1-4f79-a09a-a12979903",
						Name:        "Missing Owner Label",
						PolicyId:    "weave.policies.missing-app-label",
						ClusterId:   "cluster-1",
						Category:    "Access Control",
						Severity:    "high",
						CreatedAt:   "0001-01-01T00:00:00Z",
						Message:     "Policy event",
						Entity:      "my-deployment",
						Namespace:   "weave-system",
						ClusterName: "Default",
					},
				},
				Total: int32(2),
			},
		},
		{
			name: "list application policy violations",
			clusterState: []runtime.Object{
				makeEvent(t, func(e *corev1.Event) {
					e.ObjectMeta.Name = "Missing Owner Label - fake-event-2"
					e.InvolvedObject.Namespace = "weave-system"
					e.ObjectMeta.Namespace = "weave-system"
					e.InvolvedObject.Name = "app1"
					e.InvolvedObject.Kind = "HelmRelease"
					e.Annotations["policy_name"] = "Missing Owner Label"
					e.Annotations["policy_id"] = "weave.policies.missing-app-label"
					e.Labels["pac.weave.works/id"] = "56701548-12c1-4f79-a09a-a12979904"
				}),
			},
			expected: &pb.ListPolicyValidationsResponse{
				Violations: []*pb.PolicyValidation{
					{
						Id:          "56701548-12c1-4f79-a09a-a12979904",
						Name:        "Missing Owner Label",
						PolicyId:    "weave.policies.missing-app-label",
						ClusterId:   "cluster-1",
						Category:    "Access Control",
						Severity:    "high",
						CreatedAt:   "0001-01-01T00:00:00Z",
						Message:     "Policy event",
						Entity:      "app1",
						Namespace:   "weave-system",
						ClusterName: "Default",
					},
				},
				Total: int32(1),
			},
			appName:   "app1",
			appKind:   "HelmRelease",
			namespace: "weave-system",
		},
		{
			name: "list policy violations with cluster filtering",
			clusterState: []runtime.Object{
				makeEvent(t),
			},
			expected:    &pb.ListPolicyValidationsResponse{},
			clusterName: "wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeCl := fake.NewClientBuilder().WithRuntimeObjects(tt.clusterState...).Build()

			cfg := makeServerConfig(fakeCl, t, "")
			s := makeServer(cfg, t)

			policyViolation, err := s.ListPolicyValidations(context.Background(), &pb.ListPolicyValidationsRequest{
				ClusterName: tt.clusterName,
				Application: tt.appName,
				Kind:        tt.appKind,
				Namespace:   tt.namespace,
			})
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to list policy violation:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("unexpected error while getting policy:\n%s", diff)
				}
			} else {
				if policyViolation.Total != tt.expected.Total {
					t.Fatalf("total policy violation didn't match expected:\n%s", cmp.Diff(tt.expected.Total, policyViolation.Total))
				}
				if diff := cmp.Diff(tt.expected.Violations, policyViolation.Violations, protocmp.Transform()); diff != "" {
					t.Fatalf("policy violation didn't match expected:\n%s", diff)
				}
			}
		})
	}
}

func makeEvent(t *testing.T, opts ...func(e *corev1.Event)) *corev1.Event {
	t.Helper()
	event := &corev1.Event{
		InvolvedObject: corev1.ObjectReference{
			APIVersion:      "v1",
			Kind:            "Deployment",
			Name:            "my-deployment",
			Namespace:       "default",
			ResourceVersion: "1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"policy_name":     "Missing app Label",
				"policy_id":       "weave.policies.missing-app-label",
				"cluster_id":      "cluster-1",
				"category":        "Access Control",
				"severity":        "high",
				"description":     "Missing app label",
				"how_to_solve":    "how_to_solve",
				"entity_manifest": `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"nginx-deployment","namespace":"default","uid":"af912668-957b-46d4-bc7a-51e6994cba56"},"spec":{"template":{"spec":{"containers":[{"image":"nginx:latest","imagePullPolicy":"Always","name":"nginx","ports":[{"containerPort":80,"protocol":"TCP"}]}]}}}}`,
				"occurrences":     `[{"message": "occurrence details"}]`,
			},
			Labels: map[string]string{
				"pac.weave.works/type": "Admission",
				"pac.weave.works/id":   "66101548-12c1-4f79-a09a-a12979903fba",
			},
			Name:      "Missing app Label - fake-event-1",
			Namespace: "default",
		},
		Message: "Policy event",
		Reason:  "PolicyViolation",
		Type:    "Warning",
	}
	for _, o := range opts {
		o(event)
	}
	return event
}
