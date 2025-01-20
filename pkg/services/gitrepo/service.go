package gitrepo

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

func CommitAndPush(ctx context.Context, client git.Git, commitMsg string, logger logger.Logger, filters ...func(string) bool) error {
	logger.Actionf("Committing and pushing gitops updates for application")

	_, err := client.Commit(git.Commit{
		Author:  git.Author{Name: "Weave Gitops", Email: "weave-gitops@weaveworks.org"},
		Message: commitMsg,
	}, filters...)
	if err != nil && !errors.Is(err, git.ErrNoStagedFiles) {
		return fmt.Errorf("failed to update the repository: %w", err)
	}

	if err == nil {
		logger.Actionf("Pushing app changes to repository")

		if err = client.Push(ctx); err != nil {
			return fmt.Errorf("failed to push changes: %w", err)
		}
	} else {
		logger.Successf("App is up to date")
	}

	return nil
}

// CloneRepo uses the git client to clone the reop from the URL and branch.  It clones into a temp
// directory and returns a function to use by the caller for cleanup.  The temp directory is
// also returned.
func CloneRepo(ctx context.Context, client git.Git, url gitproviders.RepoURL, branch string) (func(), string, error) {
	repoDir, err := os.MkdirTemp("", "user-repo-")
	if err != nil {
		return nil, "", fmt.Errorf("failed creating temp. directory to clone repo: %w", err)
	}

	_, err = client.Clone(ctx, repoDir, url.String(), branch)
	if err != nil {
		return nil, "", fmt.Errorf("failed cloning user repo: %s: %w", url, err)
	}

	return func() {
		_ = os.RemoveAll(repoDir)
	}, repoDir, nil
}
