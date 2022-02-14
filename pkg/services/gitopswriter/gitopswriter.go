package gitopswriter

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/services/gitrepo"
)

const (
	AddCommitMessage     = "Add application manifests"
	RemoveCommitMessage  = "Remove application manifests"
	ClusterCommitMessage = "Associate cluster"
)

type RepoWriter interface {
	Write(ctx context.Context, repoURL gitproviders.RepoURL, branch string, manifests []gitprovider.CommitFile) error
}

type repoWriter struct {
	log         logger.Logger
	gitClient   git.Git
	gitProvider gitproviders.GitProvider
}

func NewRepoWriter(log logger.Logger, gitClient git.Git, gitProvider gitproviders.GitProvider) RepoWriter {
	return &repoWriter{
		log:         log,
		gitClient:   gitClient,
		gitProvider: gitProvider,
	}
}

func (rw *repoWriter) Write(ctx context.Context, repoURL gitproviders.RepoURL, branch string, manifests []gitprovider.CommitFile) error {
	// TODO: auto-merge will not work for most users
	remover, _, err := gitrepo.CloneRepo(ctx, rw.gitClient, repoURL, branch)
	if err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	defer remover()

	for _, m := range manifests {
		if err := rw.gitClient.Write(*m.Path, []byte(*m.Content)); err != nil {
			return fmt.Errorf("failed to write manifest: %w", err)
		}
	}

	err = gitrepo.CommitAndPush(ctx, rw.gitClient, ClusterCommitMessage, rw.log, func(fname string) bool {
		return true
	})
	if err != nil {
		return fmt.Errorf("failed pushing changes to git provider: %w", err)
	}

	return nil
}
