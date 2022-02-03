package fluxbin

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/version"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func CreateSecretGit(name, namespace string, repoURL gitproviders.RepoURL) (corev1.Secret, error) {
	var secret corev1.Secret

	args := []string{
		"create", "secret", "git", name,
		"--url", repoURL.String(),
		"--namespace", namespace,
		"--export",
	}

	binPath, err := fluxPath()
	if err != nil {
		return secret, err
	}

	cmd := exec.Command(binPath, args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return secret, err
	}

	err = yaml.Unmarshal(out, &secret)
	if err != nil {
		return secret, fmt.Errorf("failed to unmarshal created secret: %w", err)
	}

	return secret, nil
}

func fluxPath() (string, error) {
	if os.Getenv("IS_TEST_ENV") == "true" {
		return "../../pkg/flux/bin/flux", nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting user home directory: %w", err)
	}

	path := fmt.Sprintf("%v/.wego/bin", homeDir)

	return fmt.Sprintf("%v/flux-%v", path, version.FluxVersion), nil
}
