package server

import (
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getMatchingLabels(appName string) client.MatchingLabels {
	return matchLabel(withPartOfLabel(appName))
}

func getTenant(namespace, clusterName string, clusterUserNamespaces map[string][]v1.Namespace) string {
	for _, ns := range clusterUserNamespaces[clusterName] {
		if ns.GetName() == namespace {
			return ns.Labels["toolkit.fluxcd.io/tenant"]
		}
	}

	return ""
}
