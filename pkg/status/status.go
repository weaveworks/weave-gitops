package status

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"sigs.k8s.io/yaml"
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

// GetClusterName returns the cluster name associated with the current context in ~/.kube/config
func GetClusterName() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	config, err := ioutil.ReadFile(filepath.Join(homeDir, ".kube", "config"))
	if err != nil {
		return "", err
	}
	data := map[string]interface{}{}
	err = yaml.Unmarshal(config, &data)
	if err != nil {
		return "", err
	}
	return data["current-context"].(string), nil
}

func kubectlHandler(args string) error {
	cmd := fmt.Sprintf("kubectl get %s", args)
	_, err := fluxops.CallCommand(cmd)
	return err
}
