// +build !unittest

package helpers

import (
	"context"
	"fmt"

	ghAPI "github.com/google/go-github/v32/github"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	glAPI "github.com/xanzy/go-gitlab"
)

type FileFetcher interface {
	GetFilesForPullRequest(ctx context.Context, id int, org, repoName string, fs WeGODirectoryFS) (WeGODirectoryFS, error)
}

func NewFileFetcher(name gitproviders.GitProviderName, token string) (FileFetcher, error) {
	switch name {
	case gitproviders.GitProviderGitHub:
		return githubPrClient{
			client: NewGithubClient(context.Background(), token),
		}, nil
	case gitproviders.GitProviderGitLab:
		gl, err := glAPI.NewClient(token)
		if err != nil {
			return nil, err
		}

		return gitlabPrClient{
			client: gl,
		}, nil
	}

	return nil, fmt.Errorf("unkown git provider: %s", name)
}

type githubPrClient struct {
	client *ghAPI.Client
}

func (gh githubPrClient) GetFilesForPullRequest(ctx context.Context, prID int, org, repoName string, fs WeGODirectoryFS) (WeGODirectoryFS, error) {
	files, _, err := gh.client.PullRequests.ListFiles(ctx, org, repoName, prID, &ghAPI.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing files for %q: %w", repoName, err)
	}

	return GetGithubFilesContents(ctx, gh.client, org, repoName, fs, files)
}

type gitlabPrClient struct {
	client *glAPI.Client
}

func (gl gitlabPrClient) GetFilesForPullRequest(ctx context.Context, prID int, org, repoName string, fs WeGODirectoryFS) (WeGODirectoryFS, error) {
	pid := fmt.Sprintf("%s/%s", org, repoName)
	mr, _, err := gl.client.MergeRequests.GetMergeRequestChanges(pid, prID, nil)

	if err != nil {
		return nil, fmt.Errorf("could not get merge request: %w", err)
	}

	fullRepoPath := fmt.Sprintf("%s/%s", org, repoName)

	return GetGitlabFilesContents(gl.client, fullRepoPath, fs, mr.DiffRefs.HeadSha, standardizeChangesFromMergeRequest(mr))
}

func standardizeChangesFromMergeRequest(mr *glAPI.MergeRequest) []*glAPI.Diff {
	res := make([]*glAPI.Diff, 0)

	for _, change := range mr.Changes {
		res = append(res, &glAPI.Diff{
			Diff:        change.Diff,
			NewPath:     change.NewPath,
			OldPath:     change.OldPath,
			AMode:       change.AMode,
			BMode:       change.BMode,
			NewFile:     change.NewFile,
			RenamedFile: change.RenamedFile,
			DeletedFile: change.DeletedFile,
		})
	}

	return res
}
