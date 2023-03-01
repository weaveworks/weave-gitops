package clustersmngr

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClusterNamespacesLogValue(t *testing.T) {
	g := NewGomegaWithT(t)

	nss := func(names ...string) []v1.Namespace {
		var namespaces []v1.Namespace
		for _, name := range names {
			namespaces = append(namespaces, v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			})
		}
		return namespaces
	}

	tests := []struct {
		name              string
		clusterNamespaces map[string][]v1.Namespace
		expected          map[string]namespacesSlice
	}{
		{
			name:              "empty",
			clusterNamespaces: map[string][]v1.Namespace{},
			expected:          map[string]namespacesSlice{},
		},
		{
			name: "one cluster",
			clusterNamespaces: map[string][]v1.Namespace{
				"cluster1": nss("ns1", "ns2"),
			},
			expected: map[string]namespacesSlice{
				"cluster1": {
					namespaceSample: []string{"ns1", "ns2"},
					namespaceCount:  2,
				},
			},
		},
		{
			name: "two clusters",
			clusterNamespaces: map[string][]v1.Namespace{
				"cluster1": nss("ns1", "ns2"),
				"cluster2": nss("ns3", "ns4", "ns5", "ns6"),
			},
			expected: map[string]namespacesSlice{
				"cluster1": {
					namespaceSample: []string{"ns1", "ns2"},
					namespaceCount:  2,
				},
				"cluster2": {
					namespaceSample: []string{"ns3", "ns4", "ns5"},
					namespaceCount:  4,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clusterNamespacesLogValue(tt.clusterNamespaces)
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}
