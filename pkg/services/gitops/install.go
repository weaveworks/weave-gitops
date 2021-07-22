package gitops

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

type InstallParams struct {
	Namespace string
	DryRun    bool
	GitToken  string
}

func (g *Gitops) Install(params InstallParams) ([]byte, error) {
	status := g.kube.GetClusterStatus(context.Background())

	switch status {
	case kube.FluxInstalled:
		return []byte{}, errors.New("Weave GitOps does not yet support installation onto a cluster that is using Flux.\nPlease uninstall flux before proceeding:\n  $ flux uninstall")
	case kube.Unknown:
		return []byte{}, errors.New("Weave GitOps cannot talk to the cluster")
	}

	fluxManifests, err := g.flux.Install(params.Namespace, params.DryRun)
	if err != nil {
		return fluxManifests, fmt.Errorf("error on flux install %s", err)
	}

	if params.DryRun {
		fluxManifests = append(fluxManifests, manifests.AppCRD...)
	} else {
		ctx := context.Background()
		if err := g.kube.CreateSecret(ctx, kube.GitTokenSecretName, kube.GitTokenKeyName, params.GitToken, params.Namespace); err != nil {
			return nil, errors.Wrap(err, "could not create git token secret")
		}
		if out, err := g.kube.Apply(manifests.AppCRD, params.Namespace); err != nil {
			return []byte{}, errors.Wrapf(err, "failed to apply App spec CR: %s", string(out))
		}
		if out, err := g.kube.Apply(manifests.ServiceAccountApiService, params.Namespace); err != nil {
			return []byte{}, errors.Wrapf(err, "failed to apply service account manifest for api-service: %s", string(out))
		}
		if out, err := g.kube.Apply(manifests.RoleApiService, params.Namespace); err != nil {
			return []byte{}, errors.Wrapf(err, "failed to apply role manifest for api-service: %s", string(out))
		}
		if out, err := g.kube.Apply(manifests.RoleBindingApiService, params.Namespace); err != nil {
			return []byte{}, errors.Wrapf(err, "failed to apply rolebinding for api-service: %s", string(out))
		}
	}

	return fluxManifests, nil
}
