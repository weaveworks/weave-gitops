package gitops

import (
	"fmt"

	"github.com/pkg/errors"
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

	manifests, err := g.flux.Install(params.Namespace, params.DryRun)
	if err != nil {
		return manifests, fmt.Errorf("error on flux install %s", err)
	}

	if params.DryRun {
		manifests = append(manifests, appCRD...)
	} else {
		if out, err := g.kube.Apply(appCRD, params.Namespace); err != nil {
			return []byte{}, errors.Wrapf(err, "failed to apply Application spec CR: %s", string(out))
		}
	}

	return manifests, nil
}
