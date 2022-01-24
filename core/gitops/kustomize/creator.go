package kustomize

import (
	"context"

	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"k8s.io/client-go/rest"
)

const (
	kustomizations = "kustomizations"
)

type KubeCreator interface {
	Create(ctx context.Context, client *rest.RESTClient, kustomization *v1beta2.Kustomization) (result *v1beta2.Kustomization, err error)
}

func NewK8sCreator() KubeCreator {
	return &kustomizationCreator{}
}

type kustomizationCreator struct {
}

func (a kustomizationCreator) Create(ctx context.Context, client *rest.RESTClient, kustomization *v1beta2.Kustomization) (result *v1beta2.Kustomization, err error) {
	result = &v1beta2.Kustomization{}
	err = client.Post().
		Namespace(kustomization.ObjectMeta.Namespace).
		Resource(kustomizations).
		Name(kustomization.ObjectMeta.Name).
		Body(kustomization).
		Do(ctx).
		Into(result)

	return
}
