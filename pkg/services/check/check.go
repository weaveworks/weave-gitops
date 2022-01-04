package check

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

var ErrFluxNotFound = errors.New("flux is not installed")
var ErrKubernetesNotFound = errors.New("no kubernetes version found")

const (
	FluxCompatibleMessage    = "Current flux version is compatible"
	FluxNotCompatibleMessage = "Current flux version is not compatible"
)

// Pre runs pre-install checks
func Pre(ctx context.Context, kubeClient kube.Kube, fluxClient flux.Flux, expectedFluxVersion string) (string, error) {
	output := ""

	k8sOutput, err := runKubernetesCheck(fluxClient)
	if err != nil {
		return "", err
	}

	output += k8sOutput + "\n"

	currentFluxVersion, err := getCurrentFluxVersion(ctx, kubeClient)
	if err != nil {
		if errors.Is(err, ErrFluxNotFound) {
			output += "✔ Flux is not installed"
			return output, nil
		}

		return "", fmt.Errorf("failed getting installed flux version: %w", err)
	}

	fluxOutput, err := validateFluxVersion(currentFluxVersion, expectedFluxVersion)
	if err != nil {
		return "", err
	}

	output += fluxOutput

	return output, nil
}

func runKubernetesCheck(fluxClient flux.Flux) (string, error) {
	output, err := fluxClient.PreCheck()
	if err != nil {
		return "", fmt.Errorf("failed running flux pre check %w", err)
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Kubernetes") {
			return line, nil
		}
	}

	return "", ErrKubernetesNotFound
}

func getCurrentFluxVersion(ctx context.Context, kubeClient kube.Kube) (string, error) {
	namespacesList, err := kubeClient.GetNamespaces(ctx)
	if err != nil {
		return "", err
	}

	for _, namespace := range namespacesList.Items {
		labels := namespace.GetLabels()
		if labels[flux.PartOfLabelKey] == flux.PartOfLabelValue {
			return labels[flux.VersionLabelKey], nil
		}
	}

	return "", ErrFluxNotFound
}

func validateFluxVersion(actualFluxVersion string, expectedFluxVersion string) (string, error) {
	actualParsedFluxVersion, err := parseVersion(actualFluxVersion)
	if err != nil {
		return "", err
	}

	expectedParsedFluxVersion, err := parseVersion(expectedFluxVersion)
	if err != nil {
		return "", err
	}

	fluxOutput := ""

	expectedMajor := expectedParsedFluxVersion.Major()
	expectedMinor := expectedParsedFluxVersion.Minor()
	constraintFormat := fmt.Sprintf("~%d.%d.x", expectedMajor, expectedMinor)

	constraint, err := semver.NewConstraint(constraintFormat)
	if err != nil {
		return "", fmt.Errorf("failed creating semver constraint: %w", err)
	}

	check := constraint.Check(actualParsedFluxVersion)
	if check {
		fluxOutput += fmt.Sprintf("✔ Flux %s ~=%s\n", actualParsedFluxVersion, expectedParsedFluxVersion)
		fluxOutput += FluxCompatibleMessage
	} else {
		fluxOutput += fmt.Sprintf("✗ Flux %s !=%s\n", actualParsedFluxVersion, constraintFormat)
		fluxOutput += FluxNotCompatibleMessage
	}

	return fluxOutput, nil
}

func parseVersion(version string) (*semver.Version, error) {
	versionLessV := strings.TrimPrefix(version, "v")
	if _, err := semver.StrictNewVersion(versionLessV); err != nil {
		return nil, err
	}

	return semver.NewVersion(version)
}
