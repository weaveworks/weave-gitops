package status

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
)

type ClusterStatus int

const (
	Unknown ClusterStatus = iota
	Unmodified
	FluxInstalled
	WeGOInstalled
)

var lookupHandler = kubectlHandler

// GetClusterStatus retrieves the current wego status of the cluster. That is,
// it returns one of: Unknown, Unmodified, FluxInstalled, or WeGOInstalled depending on whether the cluster:
// - refuses to be queried
// - has nothing installed
// - has flux installed
// - has wego installed
func GetClusterStatus() ClusterStatus {
	if lookupHandler("deployment wego-controller -n wego-system") == nil {
		return WeGOInstalled
	}

	if lookupHandler("customresourcedefinition buckets.source.toolkit.fluxcd.io") == nil {
		return FluxInstalled
	}

	if lookupHandler("deployment coredns -n kube-system") == nil {
		return Unmodified
	}

	return Unknown
}

func kubectlHandler(args string) error {
	cmd := fmt.Sprintf("kubectl get %s", args)
	_, err := fluxops.CallCommand(cmd)
	return err
}
