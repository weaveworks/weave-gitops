package server

import (
	"github.com/weaveworks/weave-gitops/core/server/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type matchLabelOptionFn func() (key, value string)

func matchLabel(options ...matchLabelOptionFn) client.MatchingLabels {
	opts := map[string]string{}

	for _, fn := range options {
		key, value := fn()

		if key != "" && value != "" {
			opts[key] = value
		}
	}

	return opts
}

func withPartOfLabel(name string) matchLabelOptionFn {
	return func() (string, string) {
		if name != "" {
			return types.PartOfLabel, name
		}

		return "", ""
	}
}

func withInstanceLabel(name string) matchLabelOptionFn {
	return func() (string, string) {
		if name != "" {
			return types.InstanceLabel, name
		}

		return "", ""
	}
}
