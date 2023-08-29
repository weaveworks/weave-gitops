package server

import (
	. "github.com/onsi/gomega"
	"reflect"
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

type mockKnownTypes struct {
	types map[schema.GroupVersionKind]reflect.Type
}

func (m *mockKnownTypes) AllKnownTypes() map[schema.GroupVersionKind]reflect.Type {
	return m.types
}

func TestGetPrimaryKinds(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("should return highest version for each kind", func(t *testing.T) {
		type PodV1 struct{}
		type PodV2 struct{}

		type NodeV1a1 struct{}
		type NodeV1b1 struct{}
		type NodeV1b2 struct{}

		scheme := &mockKnownTypes{
			types: map[schema.GroupVersionKind]reflect.Type{
				{Group: "core", Version: "v1", Kind: "Pod"}:        reflect.TypeOf(PodV1{}),
				{Group: "core", Version: "v2", Kind: "Pod"}:        reflect.TypeOf(PodV2{}),
				{Group: "core", Version: "v1alpha1", Kind: "Node"}: reflect.TypeOf(NodeV1a1{}),
				{Group: "core", Version: "v1beta2", Kind: "Node"}:  reflect.TypeOf(NodeV1b2{}),
				{Group: "core", Version: "v1beta1", Kind: "Node"}:  reflect.TypeOf(NodeV1b1{})},
		}

		primaryKinds, err := getPrimaryKinds(scheme)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(len(primaryKinds.kinds)).To(Equal(2))

		// Expect the highest version of each kind to be returned
		g.Expect(primaryKinds.kinds["Pod"]).To(Equal(schema.GroupVersionKind{Group: "core", Version: "v2", Kind: "Pod"}))
		g.Expect(primaryKinds.kinds["Node"]).To(Equal(schema.GroupVersionKind{Group: "core", Version: "v1beta2", Kind: "Node"}))
	})
}
