package server

import (
	"github.com/weaveworks/weave-gitops/core/server/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getMatchingLabels(appName string) client.MatchingLabels {
	var opts client.MatchingLabels
	if appName != "" {
		opts = client.MatchingLabels{
			types.PartOfLabel: appName,
		}
	}

	return opts
}
