/*
Copyright 2021 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

type GoGit struct {
	path       string
	auth       transport.AuthMethod
	repository *gogit.Repository
	git        wrapper.Git
}

func New(auth transport.AuthMethod, wrapper wrapper.Git) Git {
	return &GoGit{
		auth: auth,
		git:  wrapper,
	}
}

// Open opens a git repository in the provided path, and returns a repository.
func (g *GoGit) Open(path string) (*gogit.Repository, error) {
	g.path = path
	repo, err := g.git.PlainOpen(path)

	if err != nil {
		return nil, err
	}

	g.repository = repo

	return repo, nil
}

// Init initialises the directory at path with the remote and branch provided.
//
// If the directory is successfully initialised it returns true, otherwise if
// the directory is already initialised, it returns false.
func (g *GoGit) Init(path, url, branch string) (bool, error) {
	if g.repository != nil {
		return false, nil
	}

	g.path = path

	r, err := g.git.PlainInit(path, false)
	if err != nil {
		return false, err
	}

	if _, err = r.CreateRemote(&config.RemoteConfig{
		Name: gogit.DefaultRemoteName,
		URLs: []string{url},
	}); err != nil {
		return false, err
	}

	branchRef := plumbing.NewBranchReferenceName(branch)

	if err = r.CreateBranch(&config.Branch{
		Name:   branch,
		Remote: gogit.DefaultRemoteName,
		Merge:  branchRef,
	}); err != nil {
		return false, err
	}
	// PlainInit assumes the initial branch to always be master, we can
	// overwrite this by setting the reference of the Storer to a new
	// symbolic reference (as there are no commits yet) that points
	// the HEAD to our new branch.
	if err = r.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, branchRef)); err != nil {
		return false, err
	}

	g.repository = r

	return true, nil
}

// Clone clones a starting repository URL to a path, and checks out the provided
// branch name.
//
// If the directory is successfully initialised, it returns true, otherwise it
// returns false.
func (g *GoGit) Clone(ctx context.Context, path, url, branch string) (bool, error) {
	g.path = path

	r, err := g.clone(ctx, path, url, branch, 0)
	if err != nil {
		if errors.Is(err, transport.ErrEmptyRemoteRepository) ||
			errors.Is(err, gogit.NoMatchingRefSpecError{}) {
			return g.Init(path, url, branch)
		}

		return false, err
	}

	g.repository = r

	return true, nil
}

func (g *GoGit) clone(ctx context.Context, path, url, branch string, depth int) (*gogit.Repository, error) {
	branchRef := plumbing.NewBranchReferenceName(branch)
	r, err := g.git.PlainCloneContext(ctx, path, false, &gogit.CloneOptions{
		URL:           url,
		Auth:          g.auth,
		RemoteName:    gogit.DefaultRemoteName,
		ReferenceName: branchRef,
		SingleBranch:  true,
		NoCheckout:    false,
		Progress:      nil,
		Depth:         depth,
		Tags:          gogit.NoTags,
	})

	if err != nil {
		return nil, err
	}

	return r, nil
}

// Read reads the content from the path
func (g *GoGit) Read(path string) ([]byte, error) {
	if g.repository == nil {
		return nil, ErrNoGitRepository
	}

	wt, err := g.repository.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to open the worktree: %w", err)
	}

	f, err := wt.Filesystem.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer f.Close()

	buf := bytes.NewBuffer(nil)

	_, err = buf.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("error reading bytes from buffer %s: %w", path, err)
	}

	return buf.Bytes(), nil
}

// Write writes the provided content to the path, if the file exists, it will be
// truncated.
func (g *GoGit) Write(path string, content []byte) error {
	if g.repository == nil {
		return ErrNoGitRepository
	}

	wt, err := g.repository.Worktree()
	if err != nil {
		return fmt.Errorf("failed to open the worktree: %w", err)
	}

	f, err := wt.Filesystem.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file in %s: %w", path, err)
	}
	defer f.Close()

	_, err = io.Copy(f, bytes.NewReader(content))

	return err
}

// Remove removes the file at path
func (g *GoGit) Remove(path string) error {
	if g.repository == nil {
		return ErrNoGitRepository
	}

	wt, err := g.repository.Worktree()
	if err != nil {
		return fmt.Errorf("failed to open the worktree: %w", err)
	}

	return wt.Filesystem.Remove(path)
}

func (g *GoGit) Commit(message Commit, filters ...func(string) bool) (string, error) {
	if g.repository == nil {
		return "", ErrNoGitRepository
	}

	wt, err := g.repository.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to open the worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get the worktree status: %w", err)
	}

	// go-git has [a bug](https://github.com/go-git/go-git/issues/253)
	// whereby it thinks broken symlinks to absolute paths are
	// modified. There's no circumstance in which we want to commit a
	// change to a broken symlink: so, detect and skip those.
	var changed bool

	for file, stat := range status {
		if stat.Worktree == gogit.Deleted {
			_, _ = wt.Add(file)
			changed = true

			continue
		}

		abspath := filepath.Join(g.path, file)

		isLink, err := isSymLink(abspath)
		if err != nil {
			return "", err
		}

		if isLink {
			// symlinks are OK; broken symlinks are probably a result
			// of the bug mentioned above, but not of interest in any
			// case.
			if _, err := os.Stat(abspath); os.IsNotExist(err) {
				continue
			}
		}

		skip := false

		for _, filter := range filters {
			if !filter(file) {
				skip = true
				break
			}
		}

		if !skip {
			_, _ = wt.Add(file)
			changed = true
		}
	}

	if !changed {
		head, err := g.repository.Head()
		if err != nil {
			return "", fmt.Errorf("failed to get the worktree HEAD reference: %w", err)
		}

		return head.Hash().String(), ErrNoStagedFiles
	}

	commit, err := wt.Commit(message.Message, &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  message.Name,
			Email: message.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to commit changes: %w", err)
	}

	return commit.String(), nil
}

func (g *GoGit) Push(ctx context.Context) error {
	if g.repository == nil {
		return ErrNoGitRepository
	}

	return g.repository.PushContext(ctx, &gogit.PushOptions{
		RemoteName: gogit.DefaultRemoteName,
		Auth:       g.auth,
		Progress:   nil,
	})
}

// Status returns true if no files in the repository have been modified.
func (g *GoGit) Status() (bool, error) {
	if g.repository == nil {
		return false, ErrNoGitRepository
	}

	wt, err := g.repository.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to open the worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get the worktree status: %w", err)
	}

	return status.IsClean(), nil
}

func (g *GoGit) Head() (string, error) {
	if g.repository == nil {
		return "", ErrNoGitRepository
	}

	head, err := g.repository.Head()
	if err != nil {
		return "", fmt.Errorf("failed getting repository HEAD %w", err)
	}

	return head.Hash().String(), nil
}

// GetRemoteURL returns the url of the first listed remote server
func (g *GoGit) GetRemoteURL(dir string, remoteName string) (string, error) {
	repo, err := g.Open(dir)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %s: %w", dir, err)
	}

	remote, err := repo.Remote(remoteName)
	if err != nil {
		return "", fmt.Errorf("failed to find the origin remote in the repository: %w", err)
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		return "", fmt.Errorf("remote config in %s does not have an url", dir)
	}

	return urls[0], nil
}

func (g *GoGit) ValidateAccess(ctx context.Context, url string, branch string) error {
	path, err := os.MkdirTemp("", "temp-src")
	if err != nil {
		return fmt.Errorf("error creating temporary folder %w", err)
	}

	defer os.RemoveAll(path)

	_, err = g.clone(ctx, path, url, branch, 1)
	if err != nil && !errors.Is(err, transport.ErrEmptyRemoteRepository) {
		return fmt.Errorf("error validating git repo access %w", err)
	}

	return nil
}

func (g *GoGit) Checkout(newBranch string) error {
	wt, err := g.repository.Worktree()
	if err != nil {
		return fmt.Errorf("failed getting repository work-tree %w", err)
	}

	err = wt.Checkout(&gogit.CheckoutOptions{
		Create: true,
		Branch: plumbing.NewBranchReferenceName(newBranch),
	})
	if err != nil {
		err = wt.Checkout(&gogit.CheckoutOptions{
			Force:  true,
			Branch: plumbing.NewBranchReferenceName(newBranch),
		})
		if err != nil {
			return fmt.Errorf("failed checking out branch %w", err)
		}
	}

	return nil
}

func isSymLink(fname string) (bool, error) {
	info, err := os.Lstat(fname)
	if err != nil {
		return false, fmt.Errorf("failed to check if %s is a symlink: %w", fname, err)
	}

	if info.Mode()&os.ModeSymlink > 0 {
		return true, nil
	}

	return false, nil
}
