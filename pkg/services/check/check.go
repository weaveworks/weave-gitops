package check

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

const (
	FluxCompatibleMessage    = "Current flux version is compatible"
	FluxNotCompatibleMessage = "Current flux version is not compatible"
	kubernetesConstraints    = ">=1.20.6-0"
)

// Pre runs pre-install checks
func Pre(ctx context.Context, kubeClient kube.Kube) (string, error) {
	k8sOutput, err := runKubernetesCheck()
	if err != nil {
		return "", err
	}

	return k8sOutput, nil
}

func runKubernetesCheck() (string, error) {
	versionOutput, err := exec.Command("kubectl", "version", "--short").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("unable to get kubernetes version: %w", err)
	}

	v, err := parseVersion(string(versionOutput))
	if err != nil {
		return "", fmt.Errorf("kubernetes version can't be determined: %w", err)
	}

	return checkKubernetesVersion(v)
}

func checkKubernetesVersion(version *semver.Version) (string, error) {
	var valid bool
	var vrange string
	c, _ := semver.NewConstraint(kubernetesConstraints)
	if c.Check(version) {
		valid = true
		vrange = kubernetesConstraints
	}

	if !valid {
		return "", fmt.Errorf("✗ kubernetes version %s does not match %s", version.Original(), kubernetesConstraints)
	}

	return fmt.Sprintf("✔ Kubernetes %s %s", version.String(), vrange), nil
}

func parseVersion(text string) (*semver.Version, error) {
	version := ""
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Server") {
			version = strings.Replace(line, "Server Version: v", "", 1)
		}
	}

	if _, err := semver.StrictNewVersion(version); err != nil {
		return nil, err
	}

	return semver.NewVersion(version)
}
