package server_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetViolation(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(makeValidationEvent(t), makeValidationEvent(t, func(e *corev1.Event) {
			e.ObjectMeta.Name = "Missing Owner Label - fake-event-2"
			e.InvolvedObject.Namespace = "weave-system"
			e.ObjectMeta.Namespace = "weave-system"
			e.InvolvedObject.FieldPath = "weave.policies.test-policy"
			e.Annotations["policy_name"] = "Test Policy"
			e.Annotations["policy_id"] = "weave.policies.test-policy"
			e.Labels["pac.weave.works/id"] = "66101548-12c1-4f79-a09a-a12979903fba"
		})).
		WithIndex(&corev1.Event{}, "type", client.IndexerFunc(func(o client.Object) []string {
			event := o.(*corev1.Event)
			return []string{event.Type}
		})).
		Build()

	cfg := makeServerConfig(client, t, "")
	c := makeServer(cfg, t)
	// existing validation
	res, err := c.GetPolicyValidation(ctx, &pb.GetPolicyValidationRequest{
		ValidationId: "66101548-12c1-4f79-a09a-a12979903fba",
		ClusterName:  "Default",
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Validation).NotTo(BeNil())
	g.Expect(res.Validation.Id).To(Equal("66101548-12c1-4f79-a09a-a12979903fba"))
	g.Expect(res.Validation.Name).To(Equal("Missing app Label"))
	g.Expect(res.Validation.PolicyId).To(Equal("weave.policies.missing-app-label"))
	g.Expect(res.Validation.ClusterId).To(Equal("cluster-1"))
	g.Expect(res.Validation.Category).To(Equal("Access Control"))
	g.Expect(res.Validation.Severity).To(Equal("high"))
	g.Expect(res.Validation.CreatedAt).To(Equal("0001-01-01T00:00:00Z"))
	g.Expect(res.Validation.Message).To(Equal("Policy event"))
	g.Expect(res.Validation.Entity).To(Equal("my-deployment"))
	g.Expect(res.Validation.Namespace).To(Equal("default"))
	g.Expect(res.Validation.Description).To(Equal("Missing app label"))
	g.Expect(res.Validation.HowToSolve).To(Equal("how_to_solve"))
	g.Expect(res.Validation.ViolatingEntity).To(Equal(`{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"nginx-deployment","namespace":"default","uid":"af912668-957b-46d4-bc7a-51e6994cba56"},"spec":{"template":{"spec":{"containers":[{"image":"nginx:latest","imagePullPolicy":"Always","name":"nginx","ports":[{"containerPort":80,"protocol":"TCP"}]}]}}}}`))
	g.Expect(res.Validation.ClusterName).To(Equal("Default"))
	g.Expect(res.Validation.Occurrences).To(Equal([]*pb.PolicyValidationOccurrence{{Message: "occurrence details"}}))

	// non existing validation
	res, err = c.GetPolicyValidation(ctx, &pb.GetPolicyValidationRequest{
		ValidationId: "invalid-id",
	})
	g.Expect(err).To(HaveOccurred())
	g.Expect(res.Validation).To(BeNil())
}

func TestListApplicationValidations(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(makeValidationEvent(t), makeValidationEvent(t, func(e *corev1.Event) {
			e.ObjectMeta.Name = "Missing Owner Label - fake-event-2"
			e.InvolvedObject.Namespace = "weave-system"
			e.ObjectMeta.Namespace = "weave-system"
			e.InvolvedObject.Name = "app1"
			e.InvolvedObject.Kind = "HelmRelease"
			e.Annotations["policy_name"] = "Missing Owner Label"
			e.Annotations["policy_id"] = "weave.policies.missing-app-label"
			e.Labels["pac.weave.works/id"] = "56701548-12c1-4f79-a09a-a12979904"
		})).
		WithIndex(&corev1.Event{}, "type", client.IndexerFunc(func(o client.Object) []string {
			event := o.(*corev1.Event)
			return []string{event.Type}
		})).
		Build()

	cfg := makeServerConfig(client, t, "")
	c := makeServer(cfg, t)
	res, err := c.ListPolicyValidations(ctx, &pb.ListPolicyValidationsRequest{
		Application: "app1",
		Kind:        "HelmRelease",
		Namespace:   "weave-system",
		ClusterName: "Default",
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(err).To(BeNil())
	g.Expect(len(res.Violations)).To(Equal(1))
	g.Expect(res.Violations[0].Id).To(Equal("66101548-12c1-4f79-a09a-a12979903fba"))
	g.Expect(res.Violations[0].Name).To(Equal("Missing Owner Label"))
	g.Expect(res.Violations[0].PolicyId).To(Equal("weave.policies.missing-app-label"))
	g.Expect(res.Violations[0].ClusterId).To(Equal("cluster-1"))
	g.Expect(res.Violations[0].Category).To(Equal("Access Control"))
	g.Expect(res.Violations[0].Severity).To(Equal("high"))
	g.Expect(res.Violations[0].CreatedAt).To(Equal("0001-01-01T00:00:00Z"))
	g.Expect(res.Violations[0].Message).To(Equal("Policy event"))
	g.Expect(res.Violations[0].Entity).To(Equal("app1"))
	g.Expect(res.Violations[0].Namespace).To(Equal("weave-system"))
	g.Expect(res.Violations[0].ClusterName).To(Equal("Default"))
}

func TestListPolicyValidations(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(makeValidationEvent(t), makeValidationEvent(t, func(e *corev1.Event) {
			e.ObjectMeta.Name = "Missing Owner Label - fake-event-2"
			e.InvolvedObject.Namespace = "weave-system"
			e.ObjectMeta.Namespace = "weave-system"
			e.InvolvedObject.FieldPath = "weave.policies.test-policy"
			e.Annotations["policy_name"] = "Test Policy"
			e.Annotations["policy_id"] = "weave.policies.test-policy"
			e.Labels["pac.weave.works/id"] = "66101548-12c1-4f79-a09a-a12979903fba"
		})).
		WithIndex(&corev1.Event{}, "type", client.IndexerFunc(func(o client.Object) []string {
			event := o.(*corev1.Event)
			return []string{event.Type}
		})).
		Build()

	cfg := makeServerConfig(client, t, "")
	c := makeServer(cfg, t)
	res, err := c.ListPolicyValidations(ctx, &pb.ListPolicyValidationsRequest{
		PolicyId:    "weave.policies.test-policy",
		ClusterName: "Default",
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(err).To(BeNil())
	g.Expect(len(res.Violations)).To(Equal(1))
	g.Expect(res.Violations[0].Id).To(Equal("66101548-12c1-4f79-a09a-a12979903fba"))
	g.Expect(res.Violations[0].Name).To(Equal("Test Policy"))
	g.Expect(res.Violations[0].PolicyId).To(Equal("weave.policies.test-policy"))
	g.Expect(res.Violations[0].ClusterId).To(Equal("cluster-1"))
	g.Expect(res.Violations[0].Category).To(Equal("Access Control"))
	g.Expect(res.Violations[0].Severity).To(Equal("high"))
	g.Expect(res.Violations[0].CreatedAt).To(Equal("0001-01-01T00:00:00Z"))
	g.Expect(res.Violations[0].Message).To(Equal("Policy event"))
	g.Expect(res.Violations[0].Entity).To(Equal("app1"))
	g.Expect(res.Violations[0].Namespace).To(Equal("weave-system"))
	g.Expect(res.Violations[0].ClusterName).To(Equal("Default"))
}

func makeValidationEvent(t *testing.T, opts ...func(e *corev1.Event)) *corev1.Event {
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
