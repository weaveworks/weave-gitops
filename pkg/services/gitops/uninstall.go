package gitops

import (
	"fmt"

	"github.com/pkg/errors"
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
		fmt.Println("Deleting Application CRD")
	} else {
		if out, err := g.kube.Delete(appCRD, params.Namespace); err != nil {
			return errors.Wrapf(err, "failed to delete Application CRD: %s", string(out))
		}
	}

	return nil
}
