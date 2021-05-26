package status

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/override"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

type ClusterStatus int

const (
	Unknown ClusterStatus = iota
	Unmodified
	FluxInstalled
	WeGOInstalled
)

var lookupHandler = kubectlHandler

// Function to translate ClusterStatus to a string
func (cs ClusterStatus) String() string {
	return toStatusString[cs]
}

var toStatusString = map[ClusterStatus]string{
	Unknown:       "Unknown",
	Unmodified:    "Unmodified",
	FluxInstalled: "FluxInstalled",
	WeGOInstalled: "WeGOInstalled",
}

// GetClusterStatus retrieves the current wego status of the cluster. That is,
// it returns one of: Unknown, Unmodified, FluxInstalled, or WeGOInstalled depending on whether the cluster:
// - refuses to be queried
// - has nothing installed
// - has flux installed
// - has wego installed
func GetClusterStatus() ClusterStatus {
	return statusHandler.(StatusHandler).GetClusterStatus()
}

func getStatus() ClusterStatus {
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

// GetClusterName returns the cluster name associated with the current context in ~/.kube/config
func GetClusterName() (string, error) {
	return statusHandler.(StatusHandler).GetClusterName()
}

func kubectlHandler(args string) error {
	cmd := fmt.Sprintf("kubectl get %s", args)
	err := utils.CallCommandForEffect(cmd)
	return err
}

// Status shim
type StatusHandler interface {
	GetClusterName() (string, error)
	GetClusterStatus() ClusterStatus
}

type defaultStatusHandler struct{}

var statusHandler interface{} = defaultStatusHandler{}

func (h defaultStatusHandler) GetClusterName() (string, error) {
	return utils.GetContextName()
}

func (h defaultStatusHandler) GetClusterStatus() ClusterStatus {
	return getStatus()
}

func Override(handler StatusHandler) override.Override {
	return override.Override{Handler: &statusHandler, Mock: handler, Original: statusHandler}
}

func WithStatusHandler(handler StatusHandler, fun func() error) error {
	originalHandler := statusHandler
	statusHandler = handler
	defer func() {
		statusHandler = originalHandler
	}()
	return fun()
}
