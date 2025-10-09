package server

import (
	"testing"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/weaveworks/weave-gitops/pkg/kube"
)

func TestGetFluxLikeInventory(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := t.Context()

	ks := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-kustomization",
			Namespace: "my-namespace",
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: sourcev1.GitRepositoryKind,
			},
		},
		Status: kustomizev1.KustomizationStatus{
			Inventory: &kustomizev1.ResourceInventory{
				Entries: []kustomizev1.ResourceRef{
					{
						ID:      "my-namespace_my-deployment_apps_Deployment",
						Version: "v1",
					},
				},
			},
		},
	}

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ks).Build()

	gvk := kustomizev1.GroupVersion.WithKind("Kustomization")
	entries, err := getFluxLikeInventory(ctx, k8sClient, ks.Name, ks.Namespace, gvk)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(entries).To(HaveLen(1))

	expected := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "my-deployment",
				"namespace": "my-namespace",
			},
		},
	}

	g.Expect(entries[0]).To(Equal(expected))
}

func TestParseInventoryFromUnstructured(t *testing.T) {
	// inv lives at status.inventory.entries
	stdErr := "status.inventory not found in object my-namespace/my-resource"
	testCases := []struct {
		name        string
		obj         *unstructured.Unstructured
		expected    []*unstructured.Unstructured
		expectedErr string
	}{
		{
			name: "no status field",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					// include name to make sure its included in the error
					"metadata": map[string]interface{}{
						"name":      "my-resource",
						"namespace": "my-namespace",
					},
				},
			},
			expected:    nil,
			expectedErr: stdErr,
		},
		{
			name: "empty status",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "my-resource",
						"namespace": "my-namespace",
					},
					"status": map[string]interface{}{},
				},
			},
			expected:    nil,
			expectedErr: stdErr,
		},
		{
			name: "empty inventory",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"inventory": map[string]interface{}{},
					},
				},
			},
			expected: nil,
		},
		{
			name: "mallformed inventory",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"inventory": "hi there",
					},
				},
			},
			expectedErr: ".status.inventory accessor error: hi there is of the type string, expected map[string]interface{}",
		},
		{
			name: "empty entry item",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"inventory": map[string]interface{}{
							"entries": []interface{}{
								map[string]interface{}{},
							},
						},
					},
				},
			},
			expected:    nil,
			expectedErr: "unable to parse stored object metadata: ",
		},
		{
			name: "invalid inventory",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"inventory": map[string]interface{}{
							"entries": []interface{}{
								map[string]interface{}{
									"v":  "v1",
									"id": "foo",
								},
							},
						},
					},
				},
			},
			expected:    nil,
			expectedErr: "unable to parse stored object metadata: foo",
		},
		{
			name: "valid inventory",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"inventory": map[string]interface{}{
							"entries": []interface{}{
								map[string]interface{}{
									"v":  "v1",
									"id": "my-namespace_my-deployment_apps_Deployment",
								},
								map[string]interface{}{
									"v":  "v1",
									"id": "my-other-namespace_my-configmap__ConfigMap",
								},
							},
						},
					},
				},
			},
			expected: []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "apps/v1",
						"kind":       "Deployment",
						"metadata": map[string]interface{}{
							"name":      "my-deployment",
							"namespace": "my-namespace",
						},
					},
				},
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "my-configmap",
							"namespace": "my-other-namespace",
						},
					},
				},
			},
		},
	}

	for _, tt := range testCases {
		// subtests...
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			// Parse inventory from unstructured
			entries, err := parseInventoryFromUnstructured(tt.obj)

			if err != nil || tt.expectedErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tt.expectedErr)))
			}

			g.Expect(entries).To(ConsistOf(tt.expected))
		})
	}
}

func TestSanitizeUnstructuredSecret(t *testing.T) {
	unstructuredSecret := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name":      "my-secret",
				"namespace": "my-namespace",
			},
			"type": "Opaque",
			"data": map[string]interface{}{
				"key": "dGVzdA==",
			},
		},
	}

	expected := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name":      "my-secret",
				"namespace": "my-namespace",
			},
			"type": "Opaque",
			"data": map[string]interface{}{
				"redacted": nil,
			},
		},
	}

	secret, err := sanitizeUnstructuredSecret(unstructuredSecret)

	g := NewGomegaWithT(t)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(&secret).To(Equal(expected))
}
