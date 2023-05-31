package server_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestListPolicies(t *testing.T) {

	g := NewGomegaWithT(t)

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	policy1 := &pacv2beta2.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "weave.policies.missing-app-label",
		},
		Spec: pacv2beta2.PolicySpec{
			Name:     "Missing app Label",
			ID:       "weave.policies.missing-app-label",
			Severity: "medium",
			Targets: pacv2beta2.PolicyTargets{
				Labels: []map[string]string{
					{"my-label": "my-value"},
				},
			},
		},
	}

	policy2 := &pacv2beta2.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "weave.policies.missing-owner-label",
		},
		Spec: pacv2beta2.PolicySpec{
			Name:     "Missing app Label",
			ID:       "weave.policies.missing-app-label",
			Severity: "medium",
			Targets: pacv2beta2.PolicyTargets{
				Labels: []map[string]string{
					{"my-label": "my-value"},
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(policy1, policy2).Build()
	cfg := makeServerConfig(client, t, "")
	c := makeServer(cfg, t)

	res, err := c.ListPolicies(ctx, &pb.ListPoliciesRequest{})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(err).To(BeNil())
	g.Expect(res).NotTo(BeNil())
	g.Expect(len(res.Policies)).To(Equal(2))
	g.Expect(res.Policies[0].Id).To(Equal(policy1.Spec.ID))
	g.Expect(res.Policies[0].Severity).To(Equal(policy1.Spec.Severity))

}

func TestGetPolicy(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	policy := &pacv2beta2.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "weave.policies.missing-owner-label",
		},
		Spec: pacv2beta2.PolicySpec{
			Name:     "Missing Owner Label",
			ID:       "weave.policies.missing-owner-label",
			Severity: "high",
			Code:     "foo",
			Targets: pacv2beta2.PolicyTargets{
				Kinds:      []string{"Deployment"},
				Namespaces: []string{"default"},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(policy).Build()
	cfg := makeServerConfig(client, t, "")
	c := makeServer(cfg, t)

	res, err := c.GetPolicy(ctx, &pb.GetPolicyRequest{
		PolicyName: policy.Spec.ID,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Policy).NotTo(BeNil())
	g.Expect(res.Policy.Id).To(Equal(policy.Spec.ID))
	g.Expect(res.Policy.Targets.Kinds).To(Equal(policy.Spec.Targets.Kinds))
	g.Expect(res.Policy.Targets.Namespaces).To(Equal(policy.Spec.Targets.Namespaces))
	g.Expect(res.Policy.Name).To(Equal(policy.Spec.Name))
	g.Expect(res.Policy.Severity).To(Equal(policy.Spec.Severity))
	g.Expect(res.Policy.Code).To(Equal(policy.Spec.Code))

	//Test non existing policy
	res, err = c.GetPolicy(ctx, &pb.GetPolicyRequest{
		PolicyName: "foo",
	})
	g.Expect(err).To(HaveOccurred())

}
