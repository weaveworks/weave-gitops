package check

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"k8s.io/client-go/discovery"
)

const (
	kubernetesConstraints = ">=1.26"
)

// KubernetesVersion checks if the Kubernetes version of the client is recent enough and
// returns a proper string explaining the result of the check. An error is returned
// if the check could not be performed (e.g. the cluster is not reachable).
func KubernetesVersion(c discovery.DiscoveryInterface) (string, error) {
	v, err := c.ServerVersion()
	if err != nil {
		return "", fmt.Errorf("failed getting server version: %w", err)
	}

	sv, err := semver.NewVersion(v.GitVersion)
	if err != nil {
		return "", fmt.Errorf("failed parsing server version %q: %w", v.GitVersion, err)
	}

	cons, _ := semver.NewConstraint(kubernetesConstraints)
	if !cons.Check(sv) {
		return "", fmt.Errorf("✗ kubernetes version %s does not match %s", sv.Original(), kubernetesConstraints)
	}

	return fmt.Sprintf("✔ Kubernetes %s %s", sv, kubernetesConstraints), nil
}
