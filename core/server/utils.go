package server

import (
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/weaveworks/weave-gitops/pkg/api/core"
)

func getMatchingLabels(appName string) client.MatchingLabels {
	return matchLabel(withPartOfLabel(appName))
}

func GetTenant(namespace, clusterName string, clusterUserNamespaces map[string][]v1.Namespace) string {
	for _, ns := range clusterUserNamespaces[clusterName] {
		if ns.GetName() == namespace {
			return ns.Labels["toolkit.fluxcd.io/tenant"]
		}
	}

	return ""
}

func GetClusterUserNamespacesNames(clusterUserNamespaces map[string][]v1.Namespace) []*api.ClusterNamespaceList {
	clusterNamespaces := []*api.ClusterNamespaceList{}

	for clusterName, namespaces := range clusterUserNamespaces {
		namespaceNames := []string{}
		for _, namespace := range namespaces {
			namespaceNames = append(namespaceNames, namespace.Name)
		}
		clusterNamespaces = append(clusterNamespaces, &api.ClusterNamespaceList{
			ClusterName: clusterName,
			Namespaces:  namespaceNames,
		})
	}

	return clusterNamespaces
}

// ExtractValueFromMap gets string value from map or empty string if the value is empty
func ExtractStringValueFromMap(mapName map[string]string, key string) string {
	value, ok := mapName[key]
	if !ok {
		return ""
	}
	return value
}
