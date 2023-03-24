package clustersmngr

import (
	"sort"

	v1 "k8s.io/api/core/v1"
)

type namespacesSlice struct {
	namespaceSample []string
	namespaceCount  int
}

// clusterNamespacesLogValues returns a map of cluster names to a slice of namespaces
// that are on that cluster. The slice is limited to 3 namespaces, and the total
// number of namespaces is also included.
func clusterNamespacesLogValue(clusterNamespaces map[string][]v1.Namespace) map[string]namespacesSlice {
	out := map[string]namespacesSlice{}
	for cluster, namespaces := range clusterNamespaces {
		namespaceNames := []string{}
		for _, n := range namespaces {
			namespaceNames = append(namespaceNames, n.Name)
		}
		sort.Strings(namespaceNames)

		if len(namespaceNames) > 3 {
			namespaceNames = namespaceNames[:3]
		}

		out[cluster] = namespacesSlice{
			namespaceSample: namespaceNames,
			namespaceCount:  len(namespaces),
		}
	}

	return out
}
