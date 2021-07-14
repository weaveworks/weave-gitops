package app

import (
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
	git          git.Git
	flux         flux.Flux
	kube         kube.Kube
	gitProviders gitproviders.GitProviderHandler
	logger       logger.Logger
}

func New(logger logger.Logger, git git.Git, flux flux.Flux, kube kube.Kube, gitProviders gitproviders.GitProviderHandler) *App {
	return &App{
		git:          git,
		flux:         flux,
		kube:         kube,
		gitProviders: gitProviders,
		logger:       logger,
	}
}

// Make sure App implements all the required methods.
var _ AppService = &App{}
