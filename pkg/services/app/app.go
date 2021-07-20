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
	DeployTypeKustomize DeploymentType = "kustomize"
	DeployTypeHelm      DeploymentType = "helm"

	SourceTypeGit  SourceType = "git"
	SourceTypeHelm SourceType = "helm"
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
