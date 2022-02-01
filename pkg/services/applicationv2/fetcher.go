package applicationv2

import (
	"context"
	"errors"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/models"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

var ErrNotFound = errors.New("not found")

//counterfeiter:generate . Fetcher
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

	a, err := translateApp(*app)
	if err != nil {
		return models.Application{}, err
	}

	return a, nil
}

func (f fetcher) List(ctx context.Context, namespace string) ([]models.Application, error) {
	list := &wego.ApplicationList{}

	if err := f.k8s.List(ctx, list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	result := []models.Application{}

	for _, a := range list.Items {
		app, err := translateApp(a)
		if err != nil {
			return nil, err
		}

		result = append(result, app)
	}

	return result, nil
}

func translateApp(app wego.Application) (models.Application, error) {
	var (
		appRepoUrl    gitproviders.RepoURL
		configRepoUrl gitproviders.RepoURL
		err           error
		helmSourceURL string
	)

	if wego.DeploymentType(app.Spec.SourceType) == wego.DeploymentType(wego.SourceTypeGit) {
		appRepoUrl, err = gitproviders.NewRepoURL(app.Spec.URL)
		if err != nil {
			return models.Application{}, err
		}
	}

	if wego.DeploymentType(app.Spec.SourceType) == wego.DeploymentType(wego.SourceTypeHelm) {
		helmSourceURL = app.Spec.URL
	}

	if models.IsExternalConfigRepo(app.Spec.ConfigRepo) {
		configRepoUrl, err = gitproviders.NewRepoURL(app.Spec.ConfigRepo)
		if err != nil {
			return models.Application{}, err
		}
	}

	return models.Application{
		Name:                app.Name,
		Namespace:           app.Namespace,
		HelmSourceURL:       helmSourceURL,
		GitSourceURL:        appRepoUrl,
		ConfigRepo:          configRepoUrl,
		Branch:              app.Spec.Branch,
		Path:                app.Spec.Path,
		HelmTargetNamespace: app.Spec.HelmTargetNamespace,
		SourceType:          models.SourceType(app.Spec.SourceType),
		AutomationType:      models.AutomationType(app.Spec.DeploymentType),
	}, nil
}

// FetcherFactory implementations should create applicationv2.Fetcher objects
// from a Kubernetes client.
type FetcherFactory interface {
	Create(client client.Client) Fetcher
}
