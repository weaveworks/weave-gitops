package gitops

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/weaveworks/weave-gitops/manifests"
)

type InstallParams struct {
	Namespace string
	DryRun    bool
}

func (g *Gitops) Install(params InstallParams) ([]byte, error) {
	present, err := g.kube.FluxPresent()
	if err != nil {
		return []byte{}, errors.Wrap(err, "could not verify flux presence in the cluster")
	}

	if present {
		return []byte{}, fmt.Errorf("Weave GitOps does not yet support installation onto a cluster that is using Flux.\nPlease uninstall flux before proceeding:\n  $ flux uninstall")
	}

	fluxManifests, err := g.flux.Install(params.Namespace, params.DryRun)
	if err != nil {
		return fluxManifests, fmt.Errorf("error on flux install %s", err)
	}

	if params.DryRun {
		fluxManifests = append(fluxManifests, manifests.AppCRD...)
	} else {
		if out, err := g.kube.Apply(manifests.AppCRD, params.Namespace); err != nil {
			return []byte{}, errors.Wrapf(err, "failed to apply App spec CR: %s", string(out))
		}
	}

	return fluxManifests, nil
}
