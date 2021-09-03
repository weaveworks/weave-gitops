package gitops

import (
	"context"
	"errors"
	"strings"

	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

type UinstallParams struct {
	Namespace string
	DryRun    bool
}

const notFound = "not found"

func (g *Gitops) Uninstall(params UinstallParams) error {
	ctx := context.Background()
	if g.kube.GetClusterStatus(ctx) != kube.WeGOInstalled {
		g.logger.Println("wego is not fully installed... removing any partial installation\n")
	}

	errorOccurred := false

	fluxErr := g.flux.Uninstall(params.Namespace, params.DryRun)
	if fluxErr != nil {
		g.logger.Printf("received error uninstalling flux: %q, continuing with uninstall", fluxErr)
		errorOccurred = true
	}

	if params.DryRun {
		g.logger.Actionf("Deleting App CRD")
	} else {
		if crdErr := g.kube.Delete(ctx, manifests.AppCRD); crdErr != nil {
			if !strings.HasSuffix(crdErr.Error(), notFound) {
				g.logger.Printf("received error uninstalling app CRD: %q", crdErr)
				errorOccurred = true
			}
		}
	}

	if errorOccurred {
		return errors.New("errors occurred during uninstall; the original state of the cluster may not be completely restored")
	}

	return nil
}
