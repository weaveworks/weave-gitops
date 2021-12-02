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

	return GetFileContents(ctx, gh.client, org, repoName, fs, files)
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

	files := map[string][]byte{}

	for _, c := range mr.Changes {
		path := c.OldPath
		if c.DeletedFile {
			delete(fs, path)
			continue
		}

		file, _, err := gl.client.RepositoryFiles.GetRawFile(pid, path, &glAPI.GetRawFileOptions{Ref: &mr.DiffRefs.HeadSha})
		if err != nil {
			return nil, fmt.Errorf("could not raw file for merge request: %w", err)
		}

		files[path] = file
	}

	return toK8sObjects(files, fs)
}
