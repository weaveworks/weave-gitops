package gitops

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/weaveworks/weave-gitops/manifests"
)

type UinstallParams struct {
	Namespace string
	DryRun    bool
}

func (g *Gitops) Uninstall(params UinstallParams) error {
	err := g.flux.Uninstall(params.Namespace, params.DryRun)
	if err != nil {
		return fmt.Errorf("error on flux install %s", err)
	}

	if params.DryRun {
		fmt.Println("Deleting App CRD")
	} else {
		if out, err := g.kube.Delete(manifests.AppCRD, params.Namespace); err != nil {
			return errors.Wrapf(err, "failed to delete App CRD: %s", string(out))
		}
	}

	return nil
}
