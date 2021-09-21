package gitops

import (
	"context"

	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type UninstallParams struct {
	Namespace string
	DryRun    bool
}

type UninstallError struct{}

func (e UninstallError) Error() string {
	return "errors occurred during uninstall; the original state of the cluster may not be completely restored"
}

func (g *Gitops) Uninstall(params UninstallParams) error {
	ctx := context.Background()
	if g.kube.GetClusterStatus(ctx) != kube.GitOpsInstalled {
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
			if !apierrors.IsNotFound(crdErr) {
				g.logger.Printf("received error uninstalling app CRD: %q", crdErr)

				errorOccurred = true
			}
		}
	}

	if errorOccurred {
		return UninstallError{}
	}

	return nil
}
