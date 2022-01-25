package gitopswriter

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"github.com/weaveworks/weave-gitops/pkg/services/gitrepo"

	"sigs.k8s.io/yaml"
)

const (
	AddCommitMessage     = "Add application manifests"
	RemoveCommitMessage  = "Remove application manifests"
	ClusterCommitMessage = "Associate cluster"
)

var _ GitOpsDirectoryWriter = &gitOpsDirectoryWriterSvc{}

type GitOpsDirectoryWriter interface {
	AddApplication(ctx context.Context, app models.Application, clusterName string, autoMerge bool) error
	RemoveApplication(ctx context.Context, app models.Application, clusterName string, autoMerge bool) error
}

type gitOpsDirectoryWriterSvc struct {
	Automation automation.AutomationGenerator
	RepoWriter gitrepo.RepoWriter
	Osys       osys.Osys
	Logger     logger.Logger
}

func NewGitOpsDirectoryWriter(automationSvc automation.AutomationGenerator, repoWriter gitrepo.RepoWriter, osys osys.Osys, logger logger.Logger) GitOpsDirectoryWriter {
	return &gitOpsDirectoryWriterSvc{
		Automation: automationSvc,
		RepoWriter: repoWriter,
		Osys:       osys,
		Logger:     logger,
	}
}

func (dw *gitOpsDirectoryWriterSvc) AddApplication(ctx context.Context, app models.Application, clusterName string, autoMerge bool) error {
	auto, err := dw.Automation.GenerateApplicationAutomation(ctx, app, clusterName)
	if err != nil {
		return fmt.Errorf("could not generate GitOps Automation manifests for application %s: %w", app.Name, err)
	}

	manifests := auto.Manifests()

	defaultBranch, err := dw.RepoWriter.GetDefaultBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve default branch for repository: %w", err)
	}

	remover, repoDir, err := dw.RepoWriter.CloneRepo(ctx, defaultBranch)
	if err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	defer remover()

	resourceEntry, err := appKustomizeReference(getUserKustomizationRepoPath(clusterName), appPath(app.Name))
	if err != nil {
		return err
	}

	kManifest, err := addKustomizeResources(app, repoDir, clusterName, resourceEntry)
	if err != nil {
		return err
	}

	manifests = append(manifests, kManifest)

	dw.Logger.Actionf("Adding application %q to cluster %q and repository", app.Name, clusterName)

	if autoMerge {
		if err := dw.RepoWriter.WriteAndMerge(ctx, repoDir, AddCommitMessage, manifests); err != nil {
			return fmt.Errorf("failed writing automation to disk: %w", err)
		}

		return nil
	}

	files := []gitprovider.CommitFile{}

	for _, manifest := range manifests {
		manifestPath := manifest.Path
		content := string(manifest.Content)

		files = append(files, gitprovider.CommitFile{Path: &manifestPath, Content: &content})
	}

	prInfo := gitproviders.PullRequestInfo{
		Title:         fmt.Sprintf("Gitops add %s", app.Name),
		Description:   fmt.Sprintf("Added yamls for %s", app.Name),
		CommitMessage: AddCommitMessage,
		TargetBranch:  defaultBranch,
		NewBranch:     automation.GetRandomName("wego"),
		Files:         files,
	}

	if err := dw.RepoWriter.CreatePullRequest(ctx, prInfo); err != nil {
		return fmt.Errorf("failed creating pull request: %w", err)
	}

	return nil
}

func (dw *gitOpsDirectoryWriterSvc) RemoveApplication(ctx context.Context, app models.Application, clusterName string, autoMerge bool) error {
	defaultBranch, err := dw.RepoWriter.GetDefaultBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve default branch for repository: %w", err)
	}

	newBranchName := automation.GetAppHash(app)

	remover, repoDir, err := dw.RepoWriter.CloneRepo(ctx, defaultBranch)
	if err != nil {
		return fmt.Errorf("failed to clone configuration repo: %w", err)
	}

	defer remover()

	dw.Logger.Actionf("Removing application %q from cluster %q and repository", app.Name, clusterName)

	appSubDir := automation.AppYamlDir(app)
	appDir := filepath.Join(repoDir, appSubDir)

	resourcePaths, err := dw.Osys.ReadDir(appDir)
	if err != nil {
		return fmt.Errorf("failed to read resource files: %w", err)
	}

	if !autoMerge {
		err = dw.RepoWriter.CheckoutBranch(newBranchName)
		if err != nil {
			return fmt.Errorf("failed to checkout branch in configuration repo: %w", err)
		}
	}

	for _, resourcePath := range resourcePaths {
		pathStr := filepath.Join(appSubDir, resourcePath.Name())
		if err := dw.RepoWriter.Remove(ctx, pathStr); err != nil {
			return fmt.Errorf("failed to remove app resource from repository: %w", err)
		}
	}

	// Remove reference in kustomization file
	resourceEntry, err := appKustomizeReference(getUserKustomizationRepoPath(clusterName), appPath(app.Name))
	if err != nil {
		return err
	}

	kManifest, err := removeKustomizeResources(app, repoDir, clusterName, resourceEntry)

	if err != nil {
		return fmt.Errorf("failed to remove app reference from user kustomize file: %w", err)
	}

	if err = dw.RepoWriter.Write(ctx, kManifest.Path, kManifest.Content); err != nil {
		return fmt.Errorf("failed to write updated kustomize file: %w", err)
	}

	err = dw.RepoWriter.CommitAndPush(ctx, RemoveCommitMessage)
	if err != nil {
		return fmt.Errorf("failed to commit and push changes %w", err)
	}

	if !autoMerge {
		prInfo := gitproviders.PullRequestInfo{
			Title:                     fmt.Sprintf("Gitops remove %s", app.Name),
			Description:               fmt.Sprintf("Removed yamls for %s", app.Name),
			CommitMessage:             RemoveCommitMessage,
			TargetBranch:              defaultBranch,
			NewBranch:                 newBranchName,
			SkipAddingFilesOnCreation: true,
		}

		if err := dw.RepoWriter.CreatePullRequest(ctx, prInfo); err != nil {
			return fmt.Errorf("failed creating pull request: %w", err)
		}
	}

	return nil
}

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

func addKustomizeResources(app models.Application, repoDir, clusterName string, resources ...string) (models.Manifest, error) {
	userKustomizationRepoPath := getUserKustomizationRepoPath(clusterName)
	userKustomization := filepath.Join(repoDir, userKustomizationRepoPath)

	k, err := automation.GetOrCreateKustomize(userKustomization, clusterName, app.Namespace)
	if err != nil {
		return models.Manifest{}, err
	}

	k.Resources = append(k.Resources, resources...)

	userKustomizationManifest, err := yaml.Marshal(&k)
	if err != nil {
		return models.Manifest{}, fmt.Errorf("failed to marshal kustomize %v : %w", k, err)
	}

	return models.Manifest{
		Path:    userKustomizationRepoPath,
		Content: userKustomizationManifest,
	}, nil
}

func removeKustomizeResources(app models.Application, repoDir, clusterName string, resources ...string) (models.Manifest, error) {
	userKustomizationRepoPath := getUserKustomizationRepoPath(clusterName)
	userKustomization := filepath.Join(repoDir, userKustomizationRepoPath)

	k, err := automation.GetOrCreateKustomize(userKustomization, clusterName, app.Namespace)
	if err != nil {
		return models.Manifest{}, err
	}

	oldResources := k.Resources
	newResources := []string{}

	for _, oldResource := range oldResources {
		keep := true

		for _, resource := range resources {
			if resource == oldResource {
				keep = false
				break
			}
		}

		if keep {
			newResources = append(newResources, oldResource)
		}
	}

	k.Resources = newResources

	userKustomizationManifest, err := yaml.Marshal(&k)
	if err != nil {
		return models.Manifest{}, fmt.Errorf("failed to marshal kustomize %v : %w", k, err)
	}

	return models.Manifest{
		Path:    userKustomizationRepoPath,
		Content: userKustomizationManifest,
	}, nil
}

func getUserDirRepoPath(clusterName string) string {
	return filepath.Join(git.WegoRoot, git.WegoClusterDir, clusterName, "user")
}
func getUserKustomizationRepoPath(clusterName string) string {
	return filepath.Join(getUserDirRepoPath(clusterName), "kustomization.yaml")
}

func appPath(appName string) string {
	return filepath.Join(git.WegoRoot, git.WegoAppDir, appName)
}

func appKustomizeReference(userKustomizationPath, appPath string) (string, error) {
	r, err := filepath.Rel(filepath.Dir(userKustomizationPath), appPath)
	if err != nil {
		return "", fmt.Errorf("failed to calculate the relative path between the cluster %q and app %q: %w",
			userKustomizationPath, appPath, err)
	}

	return r, nil
}
