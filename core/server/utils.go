package server

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getMatchingLabels(appName string) client.MatchingLabels {
	return matchLabel(withPartOfLabel(appName))
}
