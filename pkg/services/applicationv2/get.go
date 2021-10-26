package applicationv2

import (
	"context"
	"errors"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/models"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var ErrNotFound = errors.New("not found")

type Fetcher interface {
	Get(ctx context.Context, name string, namespace string) (models.Application, error)
	List(ctx context.Context, namespace string) ([]models.Application, error)
}

func NewFetcher(k8s client.Client) Fetcher {
	return fetcher{
		k8s: k8s,
	}
}

type fetcher struct {
	k8s client.Client
}

func (f fetcher) Get(ctx context.Context, name string, namespace string) (models.Application, error) {
	app := &wego.Application{}

	err := f.k8s.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, app)

	if apierrors.IsNotFound(err) {
		return models.Application{}, ErrNotFound
	} else if err != nil {
		return models.Application{}, err
	}

	return translateApp(*app), nil
}

func (f fetcher) List(ctx context.Context, namespace string) ([]models.Application, error) {
	list := &wego.ApplicationList{}

	if err := f.k8s.List(ctx, list); err != nil {
		return nil, err
	}

	result := []models.Application{}
	for _, a := range list.Items {
		result = append(result, translateApp(a))
	}

	return result, nil
}

func translateApp(app wego.Application) models.Application {
	return models.Application{
		Name:                app.Name,
		Namespace:           app.Namespace,
		SourceURL:           app.Spec.URL,
		ConfigURL:           app.Spec.ConfigURL,
		Branch:              app.Spec.Branch,
		Path:                app.Spec.Path,
		HelmTargetNamespace: app.Spec.HelmTargetNamespace,
		SourceType:          models.SourceType(app.Spec.SourceType),
		AutomationType:      models.AutomationType(app.Spec.DeploymentType),
	}
}
