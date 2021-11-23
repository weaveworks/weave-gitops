package applicationv2

import (
	"context"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"

	"github.com/weaveworks/weave-gitops/pkg/git"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Fetcher
type Remover interface {
	Remove(ctx context.Context, appName string, namespace string, autoMerge bool) error
}

type RepoPusher func(client git.Git) error

func NewRemover(gitClient git.Git, gitProvider gitproviders.GitProvider) Remover {
	return remover{
		gitClient:   gitClient,
		gitProvider: gitProvider,
	}
}

type remover struct {
	gitClient   git.Git
	gitProvider gitproviders.GitProvider
}

func (r remover) Remove(ctx context.Context, appName string, namespace string, autoMerge bool) error {
	return nil
}
