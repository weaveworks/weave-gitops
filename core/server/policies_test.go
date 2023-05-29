package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListPolicies(t *testing.T) {
	tests := []struct {
		name         string
		clusterState []runtime.Object
		clusterName  string
		expected     *pb.ListPoliciesResponse
		err          error
	}{
		{
			name: "list policies",
			clusterState: []runtime.Object{
				makePolicy(t),
				makePolicy(t, func(p *pacv2beta2.Policy) {
					p.ObjectMeta.Name = "weave.policies.missing-app-label"
					p.Spec.Name = "Missing app Label"
					p.Spec.Severity = "medium"
				}),
			},
			expected: &pb.ListPoliciesResponse{
				Policies: []*pb.Policy{
					{
						Name:      "Missing app Label",
						Severity:  "medium",
						Code:      "foo",
						CreatedAt: "0001-01-01T00:00:00Z",
						Targets: &pb.PolicyTargets{
							Labels: []*pb.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						ClusterName: "Default",
					},
					{
						Name:      "Missing Owner Label",
						Severity:  "high",
						Code:      "foo",
						CreatedAt: "0001-01-01T00:00:00Z",
						Targets: &pb.PolicyTargets{
							Labels: []*pb.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						ClusterName: "Default",
					},
				},
				Total: int32(2),
			},
		},
		{
			name: "list policies with parameter type string",
			clusterState: []runtime.Object{
				makePolicy(t, func(p *pacv2beta2.Policy) {
					strBytes, err := json.Marshal("value")
					if err != nil {
						t.Fatal(err)
					}
					p.Spec.Parameters = append(p.Spec.Parameters, pacv2beta2.PolicyParameters{
						Name:  "key",
						Type:  "string",
						Value: &apiextensionsv1.JSON{Raw: strBytes},
					})
				}),
			},
			expected: &pb.ListPoliciesResponse{
				Policies: []*pb.Policy{
					{
						Name:     "Missing Owner Label",
						Severity: "high",
						Code:     "foo",
						Targets: &pb.PolicyTargets{
							Labels: []*pb.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						CreatedAt: "0001-01-01T00:00:00Z",
						Parameters: []*pb.PolicyParam{
							{
								Name:  "key",
								Type:  "string",
								Value: getAnyValue(t, "string", "value"),
							},
						},
						ClusterName: "Default",
					},
				},
				Total: int32(1),
			},
		},
		{
			name: "list policies with parameter type integer",
			clusterState: []runtime.Object{
				makePolicy(t, func(p *pacv2beta2.Policy) {
					intBytes, err := json.Marshal(1)
					if err != nil {
						t.Fatal(err)
					}
					p.Spec.Parameters = append(p.Spec.Parameters, pacv2beta2.PolicyParameters{
						Name:  "key",
						Type:  "integer",
						Value: &apiextensionsv1.JSON{Raw: intBytes},
					})
				}),
			},
			expected: &pb.ListPoliciesResponse{
				Policies: []*pb.Policy{
					{
						Name:     "Missing Owner Label",
						Severity: "high",
						Code:     "foo",
						Targets: &pb.PolicyTargets{
							Labels: []*pb.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						CreatedAt: "0001-01-01T00:00:00Z",
						Parameters: []*pb.PolicyParam{
							{
								Name:  "key",
								Type:  "integer",
								Value: getAnyValue(t, "integer", int32(1)),
							},
						},
						ClusterName: "Default",
					},
				},
				Total: int32(1),
			},
		},
		{
			name: "list policies with parameter type boolean",
			clusterState: []runtime.Object{
				makePolicy(t, func(p *pacv2beta2.Policy) {
					boolBytes, err := json.Marshal(false)
					if err != nil {
						t.Fatal(err)
					}
					p.Spec.Parameters = append(p.Spec.Parameters, pacv2beta2.PolicyParameters{
						Name:  "key",
						Type:  "boolean",
						Value: &apiextensionsv1.JSON{Raw: boolBytes},
					})
				}),
			},
			expected: &pb.ListPoliciesResponse{
				Policies: []*pb.Policy{
					{
						Name:     "Missing Owner Label",
						Severity: "high",
						Code:     "foo",
						Targets: &pb.PolicyTargets{
							Labels: []*pb.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						CreatedAt: "0001-01-01T00:00:00Z",
						Parameters: []*pb.PolicyParam{
							{
								Name:  "key",
								Type:  "boolean",
								Value: getAnyValue(t, "boolean", false),
							},
						},
						ClusterName: "Default",
					},
				},
				Total: int32(1),
			},
		},
		{
			name: "list policies with parameter type array",
			clusterState: []runtime.Object{
				makePolicy(t, func(p *pacv2beta2.Policy) {
					sliceBytes, err := json.Marshal([]string{"value"})
					if err != nil {
						t.Fatal(err)
					}
					p.Spec.Parameters = append(p.Spec.Parameters, pacv2beta2.PolicyParameters{
						Name:  "key",
						Type:  "array",
						Value: &apiextensionsv1.JSON{Raw: sliceBytes},
					})
				}),
			},
			expected: &pb.ListPoliciesResponse{
				Policies: []*pb.Policy{
					{
						Name:     "Missing Owner Label",
						Severity: "high",
						Code:     "foo",
						Targets: &pb.PolicyTargets{
							Labels: []*pb.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						CreatedAt: "0001-01-01T00:00:00Z",
						Parameters: []*pb.PolicyParam{
							{
								Name:  "key",
								Type:  "array",
								Value: getAnyValue(t, "array", []string{"value"}),
							},
						},
						ClusterName: "Default",
					},
				},
				Total: int32(1),
			},
		},
		{
			name: "list policies with cluster filtering",
			clusterState: []runtime.Object{
				makePolicy(t),
			},
			expected: &pb.ListPoliciesResponse{
				Policies: []*pb.Policy{
					{
						Name:      "Missing Owner Label",
						Severity:  "high",
						Code:      "foo",
						CreatedAt: "0001-01-01T00:00:00Z",
						Targets: &pb.PolicyTargets{
							Labels: []*pb.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						ClusterName: "Default",
					},
				},
				Total: int32(1),
			},
			clusterName: "Default",
		},
		{
			name: "list policies with invalid cluster filtering",
			clusterState: []runtime.Object{
				makePolicy(t),
			},
			err:         errors.New("error while listing policies for cluster wrong: cluster wrong not found"),
			clusterName: "wrong",
		},
		{
			name: "list policies with invalid parameter type",
			clusterState: []runtime.Object{
				makePolicy(t, func(p *pacv2beta2.Policy) {
					strBytes, err := json.Marshal("value")
					if err != nil {
						t.Fatal(err)
					}
					p.Spec.Parameters = append(p.Spec.Parameters, pacv2beta2.PolicyParameters{
						Name:  "key",
						Type:  "invalid",
						Value: &apiextensionsv1.JSON{Raw: strBytes},
					})
				}),
			},
			err: errors.New("found unsupported policy parameter type invalid in policy "),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientsPool := &clustersmngrfakes.FakeClientsPool{}
			fakeCl := createClient(t, tt.clusterState...)
			clients := map[string]client.Client{"Default": fakeCl}
			clientsPool.ClientsReturns(clients)
			clientsPool.ClientReturns(fakeCl, nil)
			clientsPool.ClientStub = func(name string) (client.Client, error) {
				if c, found := clients[name]; found && c != nil {
					return c, nil
				}

				return nil, fmt.Errorf("cluster %s not found", name)
			}
			clustersClient := clustersmngr.NewClient(clientsPool, map[string][]v1.Namespace{}, logr.Discard())

			fakeFactory := &clustersmngrfakes.FakeClustersManager{}
			fakeFactory.GetImpersonatedClientReturns(clustersClient, nil)

			s, err := createServer(t, serverOptions{
				clustersManager: fakeFactory,
			})
			if err != nil {

				req := pb.ListPoliciesRequest{ClusterName: tt.clusterName}
				gotResponse, err := s.ListPolicies(context.Background(), &req)
				if err != nil {
					if tt.err == nil {
						t.Fatalf("failed to list policies:\n%s", err)
					}
				} else {
					if !cmpPoliciesResp(t, tt.expected, gotResponse) {
						t.Fatalf("policies didn't match expected:\n%+v\n%+v", tt.expected, gotResponse)
					}
				}
			}
		})
	}
}

func getAnyValue(t *testing.T, kind string, o interface{}) *anypb.Any {
	t.Helper()
	var src proto.Message
	switch kind {
	case "string":
		src = wrapperspb.String(o.(string))
	case "integer":
		src = wrapperspb.Int32(o.(int32))
	case "boolean":
		src = wrapperspb.Bool(o.(bool))
	case "array":
		src = &pb.PolicyParamRepeatedString{Value: o.([]string)}
	}
	defaultAny, err := anypb.New(src)
	if err != nil {
		t.Fatal(err)
	}
	return defaultAny
}

func TestGetPolicy(t *testing.T) {
	tests := []struct {
		name         string
		policyName   string
		clusterName  string
		clusterState []runtime.Object
		err          error
		expected     *pb.GetPolicyResponse
	}{
		{
			name:        "get policy",
			policyName:  "weave.policies.missing-owner-label",
			clusterName: "Default",
			clusterState: []runtime.Object{
				makePolicy(t),
			},
			expected: &pb.GetPolicyResponse{
				Policy: &pb.Policy{
					Name:     "Missing Owner Label",
					Severity: "high",
					Code:     "foo",
					Targets: &pb.PolicyTargets{
						Labels: []*pb.PolicyTargetLabel{
							{
								Values: map[string]string{"my-label": "my-value"},
							},
						},
					},
					CreatedAt:   "0001-01-01T00:00:00Z",
					ClusterName: "Default",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientsPool := &clustersmngrfakes.FakeClientsPool{}
			fakeCl := createClient(t, tt.clusterState...)
			clientsPool.ClientsReturns(map[string]client.Client{tt.clusterName: fakeCl})
			clientsPool.ClientReturns(fakeCl, nil)
			clustersClient := clustersmngr.NewClient(clientsPool, map[string][]v1.Namespace{}, logr.Discard())

			fakeFactory := &clustersmngrfakes.FakeClustersManager{}
			fakeFactory.GetImpersonatedClientForClusterReturns(clustersClient, nil)

			s, err := createServer(t, serverOptions{
				clustersManager: fakeFactory,
			})
			if err != nil {
				gotResponse, err := s.GetPolicy(context.Background(), &pb.GetPolicyRequest{
					PolicyName:  tt.policyName,
					ClusterName: tt.clusterName})
				if err != nil {
					if tt.err == nil {
						t.Fatalf("failed to get policy:\n%s", err)
					}
					if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
						t.Fatalf("unexpected error while getting policy:\n%s", diff)
					}
				} else {
					if !cmpPolicy(t, tt.expected.Policy, gotResponse.Policy) {
						t.Fatalf("policies didn't match expected:\n%+v\n%+v", tt.expected, gotResponse)
					}
				}
			}
		})
	}
}

func cmpPoliciesResp(t *testing.T, pol1 *pb.ListPoliciesResponse, pol2 *pb.ListPoliciesResponse) bool {
	t.Helper()
	if len(pol1.Policies) != len(pol2.Policies) {
		return false
	}

	for i := range pol1.Policies {
		if !cmpPolicy(t, pol1.Policies[i], pol2.Policies[i]) {
			return false
		}
	}

	return cmp.Equal(pol1.Total, pol2.Total)
}
func cmpPolicy(t *testing.T, pol1 *pb.Policy, pol2 *pb.Policy) bool {
	t.Helper()

	if !cmp.Equal(pol1.Id, pol2.Id, protocmp.Transform()) {
		return false
	}
	if !cmp.Equal(pol1.Targets, pol2.Targets, protocmp.Transform()) {
		return false
	}
	if !cmp.Equal(pol1.Parameters, pol2.Parameters, protocmp.Transform()) {
		return false
	}
	if !cmp.Equal(pol1.ClusterName, pol2.ClusterName, protocmp.Transform()) {
		return false
	}
	return true
}
