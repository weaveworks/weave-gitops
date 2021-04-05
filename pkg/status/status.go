package status

import (
	"fmt"
	"os/exec"
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
	if err := lookupHandler("deployment wego-controller -n wego-system"); err == nil {
		return WeGOInstalled
	}

	if err := lookupHandler("customresourcedefinition buckets.source.toolkit.fluxcd.io"); err == nil {
		return FluxInstalled
	}

	if err := lookupHandler("deployment coredns -n kube-system"); err == nil {
		return Unmodified
	}

	return Unknown
}

func kubectlHandler(args string) error {
	cmd := exec.Command(fmt.Sprintf("kubectl get %s", args))
	return cmd.Run()
}
