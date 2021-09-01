package gitops

import (
	"context"
	"errors"
	"fmt"

	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

type UinstallParams struct {
	Namespace string
	DryRun    bool
}

func (g *Gitops) Uninstall(params UinstallParams) error {
	ctx := context.Background()
	if g.kube.GetClusterStatus(ctx) != kube.WeGOInstalled {
		return errors.New("wego is not installed... exiting")
	}

	err := g.flux.Uninstall(params.Namespace, params.DryRun)
	if err != nil {
		return fmt.Errorf("error on flux install %s", err)
	}

	if params.DryRun {
		g.logger.Actionf("Deleting App CRD")
	} else {
		if err := g.kube.Delete(ctx, manifests.AppCRD); err != nil {
			return fmt.Errorf("failed to delete App CRD: %w", err)
		}
	}

	return nil
}
