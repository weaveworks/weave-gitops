package server

import (
	. "github.com/onsi/gomega"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestVersionRank(t *testing.T) {
	g := NewGomegaWithT(t)

	testCases := []struct {
		gvk1     schema.GroupVersionKind
		gvk2     schema.GroupVersionKind
		expected int
	}{
		{schema.GroupVersionKind{Version: "v1"}, schema.GroupVersionKind{Version: "v1alpha1"}, 1},
		{schema.GroupVersionKind{Version: "v1"}, schema.GroupVersionKind{Version: "v1beta1"}, 1},
		{schema.GroupVersionKind{Version: "v2"}, schema.GroupVersionKind{Version: "v1"}, 1},
		{schema.GroupVersionKind{Version: "v2beta1"}, schema.GroupVersionKind{Version: "v1beta1"}, 1},
		// Additional test cases
		{schema.GroupVersionKind{Version: "v1alpha1"}, schema.GroupVersionKind{Version: "v1alpha2"}, -1},
		{schema.GroupVersionKind{Version: "v1beta1"}, schema.GroupVersionKind{Version: "v1alpha1"}, 1},
		{schema.GroupVersionKind{Version: "v2alpha1"}, schema.GroupVersionKind{Version: "v1beta1"}, 1},
		{schema.GroupVersionKind{Version: "v2alpha2"}, schema.GroupVersionKind{Version: "v2alpha1"}, 1},
		{schema.GroupVersionKind{Version: "v2beta1"}, schema.GroupVersionKind{Version: "v2alpha1"}, 1},
		{schema.GroupVersionKind{Version: "v1beta2"}, schema.GroupVersionKind{Version: "v1beta1"}, 1},
		{schema.GroupVersionKind{Version: "v1alpha1"}, schema.GroupVersionKind{Version: "v1alpha1"}, 0},
		{schema.GroupVersionKind{Version: "v2beta2"}, schema.GroupVersionKind{Version: "v2beta1"}, 1},
		{schema.GroupVersionKind{Version: "v3alpha1"}, schema.GroupVersionKind{Version: "v2"}, 1},
		{schema.GroupVersionKind{Version: "v3"}, schema.GroupVersionKind{Version: "v3alpha1"}, 1},
		{schema.GroupVersionKind{Version: "v3beta1"}, schema.GroupVersionKind{Version: "v3alpha2"}, 1},
	}

	for _, testCase := range testCases {
		result, err := compareGVK(testCase.gvk1, testCase.gvk2)
		g.Expect(err).To(BeNil())
		g.Expect(result).To(Equal(testCase.expected))
	}
}
