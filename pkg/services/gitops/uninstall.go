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

	if params.DryRun {
		g.logger.Actionf("Deleting Weave Gitops manifests")
	} else {
		for _, manifest := range manifests.Manifests {
			if err := g.kube.Delete(ctx, manifest); err != nil {
				if !apierrors.IsNotFound(err) {
					g.logger.Printf("received error deleting manifest: %q", err)

					errorOccurred = true
				}
			}
		}
	}

	if err := g.flux.Uninstall(params.Namespace, params.DryRun); err != nil {
		g.logger.Printf("received error uninstalling flux: %q, continuing with uninstall", err)

		errorOccurred = true
	}

	if errorOccurred {
		return UninstallError{}
	}

	return nil
}
