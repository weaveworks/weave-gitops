package gitrepo

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/models"

	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

type RepoWriter interface {
	CreatePullRequest(ctx context.Context, info gitproviders.PullRequestInfo) error
	WriteAndMerge(ctx context.Context, repoDir, commitMsg string, manifests []models.Manifest) error
	CloneRepo(ctx context.Context, branch string) (func(), string, error)
	GetDefaultBranch(ctx context.Context) (string, error)
	CommitAndPush(ctx context.Context, commitMsg string, filters ...func(string) bool) error
	CheckoutBranch(newBranch string) error
	Write(ctx context.Context, path string, content []byte) error
	Remove(ctx context.Context, path string) error
}

type RepoWriterSvc struct {
	URL         gitproviders.RepoURL
	GitProvider gitproviders.GitProvider
	GitClient   git.Git
	Logger      logger.Logger
}

var _ RepoWriter = &RepoWriterSvc{}

func NewRepoWriter(url gitproviders.RepoURL, gitProvider gitproviders.GitProvider, gitClient git.Git, logger logger.Logger) RepoWriter {
	return &RepoWriterSvc{URL: url, GitProvider: gitProvider, GitClient: gitClient, Logger: logger}
}

func (rw *RepoWriterSvc) CreatePullRequest(ctx context.Context, info gitproviders.PullRequestInfo) error {
	pr, err := rw.GitProvider.CreatePullRequest(ctx, rw.URL, info)
	if err != nil {
		return fmt.Errorf("unable to create pull request: %w", err)
	}

	rw.Logger.Println("Pull Request created: %s\n", pr.Get().WebURL)

	return nil
}

func (rw *RepoWriterSvc) WriteAndMerge(ctx context.Context, repoDir, commitMsg string, manifests []models.Manifest) error {
	for _, m := range manifests {
		if err := rw.GitClient.Write(m.Path, m.Content); err != nil {
			return fmt.Errorf("Failed to write manifest: %w", err)
		}
	}

	return rw.CommitAndPush(ctx, commitMsg, func(fname string) bool {
		for _, m := range manifests {
			if fname == m.Path {
				return true
			}
		}

		return false
	})
}

func (rw *RepoWriterSvc) CloneRepo(ctx context.Context, branch string) (func(), string, error) {
	return CloneRepo(ctx, rw.GitClient, rw.URL, branch)
}

func (rw *RepoWriterSvc) CommitAndPush(ctx context.Context, commitMsg string, filters ...func(string) bool) error {
	return CommitAndPush(ctx, rw.GitClient, commitMsg, rw.Logger, filters...)
}

func (rw *RepoWriterSvc) CheckoutBranch(newBranch string) error {
	err := rw.GitClient.Checkout(newBranch)
	if err != nil {
		return fmt.Errorf("error checking out branch %s, %w", newBranch, err)
	}

	return nil
}

func CommitAndPush(ctx context.Context, client git.Git, commitMsg string, logger logger.Logger, filters ...func(string) bool) error {
	logger.Actionf("Committing and pushing gitops updates for application")

	_, err := client.Commit(git.Commit{
		Author:  git.Author{Name: "Weave Gitops", Email: "weave-gitops@weave.works"},
		Message: commitMsg,
	}, filters...)
	if err != nil && err != git.ErrNoStagedFiles {
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
	repoDir, err := ioutil.TempDir("", "user-repo-")
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

func (rw *RepoWriterSvc) GetDefaultBranch(ctx context.Context) (string, error) {
	return rw.GitProvider.GetDefaultBranch(ctx, rw.URL)
}

func (rw *RepoWriterSvc) Remove(ctx context.Context, path string) error {
	return rw.GitClient.Remove(path)
}

func (rw *RepoWriterSvc) Write(ctx context.Context, path string, content []byte) error {
	return rw.GitClient.Write(path, content)
}
