package app

import (
	"context"

	"github.com/weaveworks/weave-gitops/api/v1alpha1"
	"k8s.io/client-go/rest"
)

const (
	apps = "apps"
)

type KubeCreator interface {
	Create(ctx context.Context, client *rest.RESTClient, app *v1alpha1.Application) (*v1alpha1.Application, error)
}

func NewKubeCreator() KubeCreator {
	return &appKubeCreator{}
}

type appKubeCreator struct {
}

func (a appKubeCreator) Create(ctx context.Context, client *rest.RESTClient, app *v1alpha1.Application) (result *v1alpha1.Application, err error) {
	result = &v1alpha1.Application{}
	err = client.Post().
		Namespace(app.ObjectMeta.Namespace).
		Resource(apps).
		Name(app.ObjectMeta.Name).
		Body(app).
		Do(ctx).
		Into(result)

	return
}
