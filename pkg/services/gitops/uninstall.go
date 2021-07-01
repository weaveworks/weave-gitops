package gitops

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

type UinstallParams struct {
	Namespace string
	DryRun    bool
}

func (g *Gitops) Uninstall(params UinstallParams) error {
	if g.kube.GetClusterStatus(context.Background()) != kube.WeGOInstalled {
		return fmt.Errorf("Wego is not installed... exiting")
	}

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
