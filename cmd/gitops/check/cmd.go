package check

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"

	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/version"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

var (
	pre bool
)

var Cmd = &cobra.Command{
	Use:   "check",
	Short: "Validates flux compatibility",
	Example: `
# Validate flux compatibility
gitops check --pre
`,
	RunE: runCmd,
}

var ErrFluxNotFound = errors.New("flux is not installed")

func runCmd(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	fluxClient := flux.New(osys.New(), &runner.CLIRunner{})

	rest, clusterName, err := kube.RestConfig()
	if err != nil {
		return fmt.Errorf("failed getting rest config: %w", err)
	}

	_, k8sClient, err := kube.NewKubeHTTPClientWithConfig(rest, clusterName)
	if err != nil {
		return fmt.Errorf("failed creating k8s client: %w", err)
	}

	kubeClient := &kube.KubeHTTP{
		Client: k8sClient,
	}

	currentFluxVersion, err := getCurrentFluxVersion(os.Stdout, ctx, fluxClient, kubeClient)
	if err != nil {
		return fmt.Errorf("failed getting current flux version: %w", err)
	}

	if err := validateFluxVersion(currentFluxVersion); err != nil {
		return fmt.Errorf("failed comparing flux versions %w", err)
	}

	return nil
}

func getCurrentFluxVersion(out io.Writer, ctx context.Context, fluxClient flux.Flux, kubeClient kube.Kube) (string, error) {
	output, err := fluxClient.PreCheck()
	if err != nil {
		return "", fmt.Errorf("failed running flux pre check %w", err)
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Kubernetes") {
			_, err = out.Write([]byte(line + "\n"))
			if err != nil {
				return "", fmt.Errorf("failed printing out Kubernetes line: %w", err)
			}
		}
	}

	namespacesList, err := kubeClient.GetNamespaces(ctx)
	if err != nil {
		return "", err
	}

	for _, namespace := range namespacesList.Items {
		labels := namespace.GetLabels()
		if labels["app.kubernetes.io/part-of"] == "flux" {
			return labels["app.kubernetes.io/version"], nil
		}
	}

	return "", ErrFluxNotFound
}

func validateFluxVersion(currentFluxVersion string) error {
	expectedFluxVersion, err := parseVersion(version.FluxVersion)
	if err != nil {
		return err
	}

	actualFluxVersion, err := parseVersion(currentFluxVersion)
	if err != nil {
		return err
	}

	if actualFluxVersion.LessThan(expectedFluxVersion) {
		fmt.Printf("✗ Flux %s <%s", actualFluxVersion, expectedFluxVersion)
	} else {
		fmt.Printf("✔ Flux %s >=%s", actualFluxVersion, expectedFluxVersion)
	}

	return nil
}

func parseVersion(version string) (*semver.Version, error) {
	versionLessV := strings.TrimPrefix(version, "v")
	if _, err := semver.StrictNewVersion(versionLessV); err != nil {
		return nil, err
	}

	return semver.NewVersion(version)
}

func init() {
	Cmd.Flags().BoolVarP(&pre, "pre", "d", false, "TODO")
}
