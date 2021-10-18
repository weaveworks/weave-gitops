package gitops

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

type InstallParams struct {
	Namespace string
	DryRun    bool
}

func (g *Gitops) Install(params InstallParams) ([]byte, error) {
	ctx := context.Background()
	status := g.kube.GetClusterStatus(ctx)

	switch status {
	case kube.FluxInstalled:
		return nil, errors.New("Weave GitOps does not yet support installation onto a cluster that is using Flux.\nPlease uninstall flux before proceeding:\n  $ flux uninstall")
	case kube.Unknown:
		return nil, errors.New("Weave GitOps cannot talk to the cluster")
	}

	fluxManifests, err := g.flux.Install(params.Namespace, params.DryRun)
	if err != nil {
		return fluxManifests, fmt.Errorf("error on flux install %s", err)
	}

	if params.DryRun {
		fluxManifests = append(fluxManifests, manifests.AppCRD...)
	} else {

		for _, manifest := range manifests.Manifests {
			if err := g.kube.Apply(ctx, manifest, params.Namespace); err != nil {
				return nil, fmt.Errorf("could not apply manifest: %w", err)
			}
		}

		version := version.Version
		if os.Getenv("IS_TEST_ENV") != "" {
			version = "latest"
		}

		wegoAppDeploymentManifest, err := manifests.GenerateWegoAppDeploymentManifest(version)
		if err != nil {
			return nil, fmt.Errorf("error generating wego-app deployment, %w", err)
		}
		if err := g.kube.Apply(ctx, wegoAppDeploymentManifest, params.Namespace); err != nil {
			return nil, fmt.Errorf("could not apply wego-app deployment manifest: %w", err)
		}
	}

	return fluxManifests, nil
}
