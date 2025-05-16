package fluxsync

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetLastHandledReconcileRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	obj := &UnstructuredAdapter{
		Unstructured: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"status": map[string]interface{}{
					"lastHandledReconcileAt": "2023-10-20T10:10:10Z",
				},
			},
		},
	}

	expected := "2023-10-20T10:10:10Z"
	got := obj.GetLastHandledReconcileRequest()
	g.Expect(got).To(Equal(expected))
}

func TestGetConditions(t *testing.T) {
	g := NewGomegaWithT(t)

	condition := v1.Condition{
		Type:   "Ready",
		Status: "True",
	}
	unstructuredCondition, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(&condition)

	obj := &UnstructuredAdapter{
		Unstructured: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"status": map[string]interface{}{
					"conditions": []interface{}{unstructuredCondition},
				},
			},
		},
	}

	conditions := obj.GetConditions()
	g.Expect(conditions).To(HaveLen(1))
	g.Expect(conditions[0].Type).To(Equal(condition.Type))
	g.Expect(conditions[0].Status).To(Equal(condition.Status))
}

func TestSetSuspended(t *testing.T) {
	g := NewGomegaWithT(t)

	obj := &UnstructuredAdapter{
		Unstructured: &unstructured.Unstructured{
			Object: make(map[string]interface{}),
		},
	}

	err := obj.SetSuspended(true)
	g.Expect(err).NotTo(HaveOccurred())
	suspend, _, _ := unstructured.NestedBool(obj.Object, "spec", "suspend")
	g.Expect(suspend).To(BeTrue())
}

func TestDeepCopyClientObject(t *testing.T) {
	g := NewGomegaWithT(t)

	obj := &UnstructuredAdapter{
		Unstructured: &unstructured.Unstructured{
			Object: map[string]interface{}{"key": "value"},
		},
	}

	objCopy := obj.DeepCopyClientObject().(*unstructured.Unstructured)
	g.Expect(objCopy.Object).To(Equal(obj.Object))
	g.Expect(objCopy).ToNot(BeIdenticalTo(obj))
}

func TestAsClientObjectCompatibilityWithTestClient(t *testing.T) {
	g := NewGomegaWithT(t)

	scheme := runtime.NewScheme()

	cl := fake.NewClientBuilder().WithScheme(scheme).Build()

	obj := &UnstructuredAdapter{
		Unstructured: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "test-cm",
					"namespace": "default",
				},
				"data": map[string]interface{}{"key": "value"},
			},
		},
	}

	err := cl.Create(t.Context(), obj.AsClientObject())
	g.Expect(err).NotTo(HaveOccurred())

	retrieved := &UnstructuredAdapter{
		Unstructured: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
			},
		},
	}
	err = cl.Get(t.Context(), client.ObjectKey{Namespace: "default", Name: "test-cm"}, retrieved.AsClientObject())
	g.Expect(err).NotTo(HaveOccurred())

	// check the data key
	data, _, _ := unstructured.NestedStringMap(retrieved.Object, "data")
	g.Expect(data).To(Equal(map[string]string{"key": "value"}))
}
