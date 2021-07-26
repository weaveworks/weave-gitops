package app

import (
	"fmt"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"k8s.io/apimachinery/pkg/types"
)

type DeploymentType string
type SourceType string

const (
	// TODO: use wego.DeployType and wego.SourceType everywhere
	// Convert these for now to avoid having to change large parts of the code base to the wego types
	DeployTypeKustomize DeploymentType = DeploymentType(wego.DeploymentTypeKustomize)
	DeployTypeHelm      DeploymentType = DeploymentType(wego.DeploymentTypeHelm)

	SourceTypeGit  SourceType = SourceType(wego.SourceTypeGit)
	SourceTypeHelm SourceType = SourceType(wego.SourceTypeHelm)
)

// AppService entity that manages applications
type AppService interface {
	// Add adds a new application to the cluster
	Add(params AddParams) error
	// Get returns a given applicaiton
	Get(name types.NamespacedName) (*wego.Application, error)
	// Status returns flux resources status and the last successful reconciliation time
	Status(params StatusParams) (string, string, error)
}

type App struct {
	git                git.Git
	flux               flux.Flux
	kube               kube.Kube
	logger             logger.Logger
	gitProviderFactory func(token string) (gitproviders.GitProvider, error)
}

func New(logger logger.Logger, git git.Git, flux flux.Flux, kube kube.Kube) *App {
	return &App{
		git:                git,
		flux:               flux,
		kube:               kube,
		logger:             logger,
		gitProviderFactory: createGitProvider,
	}
}

// Make sure App implements all the required methods.
var _ AppService = &App{}

func createGitProvider(token string) (gitproviders.GitProvider, error) {
	provider, err := gitproviders.New(gitproviders.Config{
		Provider: gitproviders.GitProviderGitHub,
		Token:    token,
	})
	if err != nil {
		return nil, fmt.Errorf("failed initializing git provider: %w", err)
	}

	return provider, nil
}
