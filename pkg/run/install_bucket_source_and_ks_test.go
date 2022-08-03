package run_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/pkg/run"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// mock controller-runtime client
type clientMock struct {
	client.Client
}

// mock client.List
func (c *clientMock) List(_ context.Context, list client.ObjectList, _ ...client.ListOption) error { // m
	list.(*unstructured.UnstructuredList).Items = []unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      "deployment",
					"namespace": "default",
				},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{
							"type":    "Ready",
							"status":  "False",
							"message": "This is message",
						},
						map[string]interface{}{
							"type":    "Healthy",
							"status":  "True",
							"message": "no error",
						},
					},
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      "app2",
					"namespace": "default",
				},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{
							"type":    "Ready",
							"status":  "True",
							"message": "no error",
						},
						map[string]interface{}{
							"type":    "Healthy",
							"status":  "True",
							"message": "no error",
						},
					},
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      "app3",
					"namespace": "default",
				},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{
							"type":    "Ready",
							"status":  "False",
							"message": "app 3 error",
						},
						map[string]interface{}{
							"type":    "Healthy",
							"status":  "False",
							"message": "time out",
						},
					},
				},
			},
		},
	}

	return nil
}

var _ = Describe("FindConditionMessages", func() {
	It("returns the condition messages", func() {
		client := &clientMock{}
		ks := &kustomizev1.Kustomization{
			Spec: kustomizev1.KustomizationSpec{},
			Status: kustomizev1.KustomizationStatus{
				Inventory: &kustomizev1.ResourceInventory{
					Entries: []kustomizev1.ResourceRef{
						{
							ID:      "default_deployment_apps_Deployment",
							Version: "v1",
						},
						{
							ID:      "default_app2_apps_Deployment",
							Version: "v1",
						},
						{
							ID:      "default_app3_apps_Deployment",
							Version: "v1",
						},
					},
				},
			},
		}
		messages, err := run.FindConditionMessages(client, ks)
		Expect(err).ToNot(HaveOccurred())
		Expect(messages).To(Equal([]string{
			"Deployment default/deployment: This is message",
			"Deployment default/app3: app 3 error",
			"Deployment default/app3: time out",
		}))
	})
})
