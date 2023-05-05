package server

import (
	api "github.com/weaveworks/weave-gitops/pkg/api/core"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func GetClusterUserNamespacesNames(clusterUserNamespaces map[string][]v1.Namespace) map[string]*api.NamespaceList {
	namespaces := make(map[string]*api.NamespaceList)

	for clusterName := range clusterUserNamespaces {
		var clusterNamespaces []string

		for _, ns := range clusterUserNamespaces[clusterName] {
			clusterNamespaces = append(clusterNamespaces, ns.GetName())
		}

		namespaces[clusterName] = &api.NamespaceList{
			Namespaces: clusterNamespaces,
		}
	}

	return namespaces
}
