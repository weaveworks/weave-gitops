package profile

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
)

type ProfileService interface {
	// Add adds a new profile to the cluster
	Add(configGit git.Git, gitProvider gitproviders.GitProvider, params AddParams) error
}

type ProfileSvc struct {
	Context context.Context
	Osys    osys.Osys
	Logger  logger.Logger
	Clock   clock.Clock
}

func NewService(ctx context.Context, logger logger.Logger, osys osys.Osys) *ProfileSvc {
	return &ProfileSvc{
		Context: ctx,
		Logger:  logger,
		Osys:    osys,
		Clock:   clock.New(),
	}
}
