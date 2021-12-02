//go:build !unittest
// +build !unittest

package server_test

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sigs.k8s.io/kustomize/api/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/fluxcd/go-git-providers/gitprovider"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	ghAPI "github.com/google/go-github/v32/github"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"github.com/weaveworks/weave-gitops/test/integration/server/helpers"
	glAPI "github.com/xanzy/go-gitlab"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

var _ = Describe("AddApplication", func() {
	var (
		namespace      *corev1.Namespace
		ctx            context.Context
		sourceRepoName string
		configRepoName string
		client         pb.ApplicationsClient
	)

	BeforeEach(func() {
		namespace = &corev1.Namespace{}
		namespace.Name = "kube-test-" + rand.String(5)
		Expect(env.Client.Create(context.Background(), namespace)).To(Succeed())
		sourceRepoName = "test-source-repo-" + rand.String(5)
		configRepoName = "test-config-repo-" + rand.String(5)
		client = pb.NewApplicationsClient(conn)
	})

	Context("GitHub", func() {
		var gh *ghAPI.Client
		var gp gitprovider.Client
		var githubOrg = "weaveworks-gitops-test"
		var githubToken = os.Getenv("GITHUB_TOKEN")
		BeforeEach(func() {
			gh = helpers.NewGithubClient(ctx, githubToken)
			ctx = middleware.ContextWithGRPCAuth(context.Background(), githubToken)
			Expect(err).NotTo(HaveOccurred())
			gp, err = github.NewClient(
				gitprovider.WithDestructiveAPICalls(true),
				gitprovider.WithOAuth2Token(os.Getenv("GITHUB_TOKEN")),
			)
			Expect(err).NotTo(HaveOccurred())
		})
		Context("via pull request", func() {
			It("adds with no config repo specified", func() {
				sourceRepoURL := fmt.Sprintf("https://github.com/%s/%s", githubOrg, sourceRepoName)

				repo, ref, err := helpers.CreatePopulatedSourceRepo(ctx, gp, sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				defer func() { Expect(repo.Delete(ctx)).To(Succeed()) }()

				req := &pb.AddApplicationRequest{
					Name:      "my-app",
					Namespace: namespace.Name,
					Url:       ref.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
					Branch:    "main",
					Path:      "k8s/overlays/development",
				}

				res, err := client.AddApplication(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				Expect(res.Success).To(BeTrue(), "request should have been successful")

				_, err = repo.DeployKeys().Get(ctx, gitproviders.DeployKeyName)
				Expect(err).NotTo(HaveOccurred(), "deploy key should have been found")

				prs, err := repo.PullRequests().List(ctx)
				Expect(err).NotTo(HaveOccurred())

				Expect(prs).To(HaveLen(1))

				root := helpers.InAppRoot

				fs := helpers.MakeWeGOFS(root, req.Name, clusterName)

				fetcher, err := helpers.NewFileFetcher(gitproviders.GitProviderGitHub, githubToken)
				Expect(err).NotTo(HaveOccurred())

				actual, err := fetcher.GetFilesForPullRequest(ctx, 1, githubOrg, sourceRepoName, fs)
				Expect(err).NotTo(HaveOccurred())

				expectedKustomization := kustomizev1.KustomizationSpec{
					// Flux adds a prepending `./` to path arguments that doesn't already have it.
					// https://github.com/fluxcd/flux2/blob/ca496d393d993ac5119ed84f83e010b8fe918c53/cmd/flux/create_kustomization.go#L115
					Path: "./" + req.Path,
					// Flux kustomization default; I couldn't find an export default from the package.
					Interval: metav1.Duration{Duration: time.Duration(1 * time.Minute)},
					Prune:    true,
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Name: req.Name,
						Kind: sourcev1.GitRepositoryKind,
					},
					Force: false,
				}

				repoURL, err := gitproviders.NewRepoURL(sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				expectedSource := sourcev1.GitRepositorySpec{
					URL: req.Url,
					SecretRef: &meta.LocalObjectReference{
						Name: automation.CreateRepoSecretName(repoURL).String(),
					},
					Interval: metav1.Duration{Duration: time.Duration(30 * time.Second)},
					Reference: &sourcev1.GitRepositoryRef{
						Branch: req.Branch,
					},
					Ignore: helpers.GetIgnoreSpec(),
				}

				expectedApp := wego.ApplicationSpec{
					URL:            req.Url,
					Branch:         req.Branch,
					Path:           req.Path,
					ConfigURL:      req.Url,
					DeploymentType: wego.DeploymentTypeKustomize,
					SourceType:     wego.SourceTypeGit,
				}

				expected := helpers.GenerateExpectedFS(req, root, clusterName, expectedApp, expectedKustomization, expectedSource)

				diff, err := helpers.DiffFS(actual, expected)
				if err != nil {
					GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
				}
			})
			It("adds an app with an external config repo", func() {
				sourceRepoURL := fmt.Sprintf("https://github.com/%s/%s", githubOrg, sourceRepoName)
				configRepoURL := fmt.Sprintf("https://github.com/%s/%s", githubOrg, configRepoName)

				configRepo, configRef, err := helpers.CreateRepo(ctx, gp, configRepoURL)
				Expect(err).NotTo(HaveOccurred())

				defer func() { Expect(configRepo.Delete(ctx)).To(Succeed()) }()

				sourceRepo, sourceRef, err := helpers.CreatePopulatedSourceRepo(ctx, gp, sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				defer func() { Expect(sourceRepo.Delete(ctx)).To(Succeed()) }()

				req := &pb.AddApplicationRequest{
					Name:      "my-app",
					Namespace: namespace.Name,
					Url:       sourceRef.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
					Branch:    "main",
					Path:      "k8s/overlays/development",
					ConfigUrl: configRef.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
				}

				res, err := client.AddApplication(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.Success).To(BeTrue(), "request should have been successful")

				_, err = configRepo.DeployKeys().Get(ctx, gitproviders.DeployKeyName)
				Expect(err).NotTo(HaveOccurred(), "deploy key should have been found")

				prs, err := configRepo.PullRequests().List(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(prs).To(HaveLen(1))

				root := helpers.ExternalConfigRoot
				fs := helpers.MakeWeGOFS(root, req.Name, clusterName)

				fetcher, err := helpers.NewFileFetcher(gitproviders.GitProviderGitHub, githubToken)
				Expect(err).NotTo(HaveOccurred())

				actual, err := fetcher.GetFilesForPullRequest(ctx, 1, githubOrg, configRepoName, fs)
				Expect(err).NotTo(HaveOccurred())

				normalizedUrl, err := gitproviders.NewRepoURL(req.Url)
				Expect(err).NotTo(HaveOccurred())

				expectedApp := wego.ApplicationSpec{
					URL:            normalizedUrl.String(),
					Branch:         req.Branch,
					Path:           req.Path,
					ConfigURL:      req.ConfigUrl,
					DeploymentType: wego.DeploymentTypeKustomize,
					SourceType:     wego.SourceTypeGit,
				}

				expectedKustomization := kustomizev1.KustomizationSpec{
					Path:     "./" + req.Path,
					Interval: metav1.Duration{Duration: time.Duration(1 * time.Minute)},
					Prune:    true,
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Name: req.Name,
						Kind: sourcev1.GitRepositoryKind,
					},
					Force: false,
				}

				repoURL, err := gitproviders.NewRepoURL(sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				expectedSource := sourcev1.GitRepositorySpec{
					URL: req.Url,
					SecretRef: &meta.LocalObjectReference{
						// Might be a bug? Should be configRepoURL?
						Name: automation.CreateRepoSecretName(repoURL).String(),
					},
					Interval: metav1.Duration{Duration: time.Duration(30 * time.Second)},
					Reference: &sourcev1.GitRepositoryRef{
						Branch: req.Branch,
					},
					Ignore: helpers.GetIgnoreSpec(),
				}

				expected := helpers.GenerateExpectedFS(req, root, clusterName, expectedApp, expectedKustomization, expectedSource)

				diff, err := helpers.DiffFS(actual, expected)
				if err != nil {
					GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
				}
			})
			It("remove app with config repo specified", func() {
				sourceRepoURL := fmt.Sprintf("https://github.com/%s/%s", githubOrg, sourceRepoName)

				repo, ref, err := helpers.CreatePopulatedSourceRepo(ctx, gp, sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				defer func() { Expect(repo.Delete(ctx)).To(Succeed()) }()

				appName := "my-app"

				req := &pb.AddApplicationRequest{
					Name:      appName,
					Namespace: namespace.Name,
					Url:       ref.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
					Branch:    "main",
					Path:      "k8s/overlays/development",
					AutoMerge: true,
				}

				res, err := client.AddApplication(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				Expect(res.Success).To(BeTrue(), "request should have been successful")

				_, err = repo.DeployKeys().Get(ctx, gitproviders.DeployKeyName)
				Expect(err).NotTo(HaveOccurred(), "deploy key should have been found")

				prs, err := repo.PullRequests().List(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(prs).To(HaveLen(0))

				root := helpers.InAppRoot

				fs := helpers.MakeWeGOFS(root, req.Name, clusterName)

				commits, _, err := gh.Repositories.ListCommits(ctx, githubOrg, sourceRepoName, &ghAPI.CommitsListOptions{SHA: "main"})
				Expect(err).NotTo(HaveOccurred())
				Expect(commits).To(HaveLen(3))

				appAddCommit := commits[0]

				c, _, err := gh.Repositories.GetCommit(ctx, githubOrg, sourceRepoName, *appAddCommit.SHA)
				Expect(err).NotTo(HaveOccurred())

				actual, err := helpers.GetFileContents(ctx, gh, githubOrg, sourceRepoName, fs, c.Files)
				Expect(err).NotTo(HaveOccurred())

				expectedApp := wego.ApplicationSpec{
					URL:            req.Url,
					Branch:         req.Branch,
					Path:           req.Path,
					ConfigURL:      ref.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
					DeploymentType: wego.DeploymentTypeKustomize,
					SourceType:     wego.SourceTypeGit,
				}
				app := &wego.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      appName,
						Namespace: namespace.Name,
					},
					Spec: expectedApp,
				}

				Expect(env.Client.Create(ctx, app)).Should(Succeed())

				expectedKustomization := kustomizev1.KustomizationSpec{
					// Flux adds a prepending `./` to path arguments that doesn't already have it.
					// https://github.com/fluxcd/flux2/blob/ca496d393d993ac5119ed84f83e010b8fe918c53/cmd/flux/create_kustomization.go#L115
					Path: "./" + req.Path,
					// Flux kustomization default; I couldn't find an export default from the package.
					Interval: metav1.Duration{Duration: time.Duration(1 * time.Minute)},
					Prune:    true,
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Name: req.Name,
						Kind: sourcev1.GitRepositoryKind,
					},
					Force: false,
				}

				repoURL, err := gitproviders.NewRepoURL(sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				expectedSrc := sourcev1.GitRepositorySpec{
					URL: req.Url,
					SecretRef: &meta.LocalObjectReference{
						Name: automation.CreateRepoSecretName(repoURL).String(),
					},
					Interval: metav1.Duration{Duration: 30 * time.Second},
					Reference: &sourcev1.GitRepositoryRef{
						Branch: req.Branch,
					},
					Ignore: helpers.GetIgnoreSpec(),
				}

				expected := helpers.GenerateExpectedFS(req, root, clusterName, expectedApp, expectedKustomization, expectedSrc)

				diff, err := helpers.DiffFS(actual, expected)
				if err != nil {
					GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
				}

				removeRequest := &pb.RemoveApplicationRequest{
					Name:      appName,
					Namespace: namespace.Name,
					AutoMerge: false,
				}

				removeResponse, err := client.RemoveApplication(ctx, removeRequest)
				Expect(err).NotTo(HaveOccurred())

				Expect(removeResponse.Success).To(BeTrue())

				prs, err = repo.PullRequests().List(ctx)
				Expect(err).NotTo(HaveOccurred())

				Expect(prs).To(HaveLen(1))

				fs = helpers.MakeWeGOFS(root, req.Name, clusterName)

				fetcher, err := helpers.NewFileFetcher(gitproviders.GitProviderGitHub, githubToken)
				Expect(err).NotTo(HaveOccurred())

				actual, err = fetcher.GetFilesForPullRequest(ctx, 1, githubOrg, sourceRepoName, fs)
				Expect(err).NotTo(HaveOccurred())

				expected = map[string]interface{}{
					filepath.Join(root, "clusters", clusterName, "user", "kustomization.yaml"): &types.Kustomization{
						TypeMeta: types.TypeMeta{
							Kind:       types.KustomizationKind,
							APIVersion: types.KustomizationVersion,
						},
						MetaData: &types.ObjectMeta{
							Name:      appName,
							Namespace: namespace.Name,
						},
					},
				}

				diff, err = helpers.DiffFS(actual, expected)
				if err != nil {
					GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
				}

			})
		})
		Context("via auto merge", func() {
			It("add/remove with no config repo specified", func() {
				sourceRepoURL := fmt.Sprintf("https://github.com/%s/%s", githubOrg, sourceRepoName)

				repo, ref, err := helpers.CreatePopulatedSourceRepo(ctx, gp, sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				defer func() { Expect(repo.Delete(ctx)).To(Succeed()) }()

				appName := "my-app"

				req := &pb.AddApplicationRequest{
					Name:      appName,
					Namespace: namespace.Name,
					Url:       ref.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
					Branch:    "main",
					Path:      "k8s/overlays/development",
					AutoMerge: true,
				}

				res, err := client.AddApplication(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				Expect(res.Success).To(BeTrue(), "request should have been successful")

				_, err = repo.DeployKeys().Get(ctx, gitproviders.DeployKeyName)
				Expect(err).NotTo(HaveOccurred(), "deploy key should have been found")

				prs, err := repo.PullRequests().List(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(prs).To(HaveLen(0))

				root := helpers.InAppRoot

				fs := helpers.MakeWeGOFS(root, req.Name, clusterName)

				commits, _, err := gh.Repositories.ListCommits(ctx, githubOrg, sourceRepoName, &ghAPI.CommitsListOptions{SHA: "main"})
				Expect(err).NotTo(HaveOccurred())
				Expect(commits).To(HaveLen(3))

				appAddCommit := commits[0]

				c, _, err := gh.Repositories.GetCommit(ctx, githubOrg, sourceRepoName, *appAddCommit.SHA)
				Expect(err).NotTo(HaveOccurred())

				actual, err := helpers.GetFileContents(ctx, gh, githubOrg, sourceRepoName, fs, c.Files)
				Expect(err).NotTo(HaveOccurred())

				expectedApp := wego.ApplicationSpec{
					URL:            req.Url,
					Branch:         req.Branch,
					Path:           req.Path,
					ConfigURL:      ref.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
					DeploymentType: wego.DeploymentTypeKustomize,
					SourceType:     wego.SourceTypeGit,
				}
				app := &wego.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      appName,
						Namespace: namespace.Name,
					},
					Spec: expectedApp,
				}

				Expect(env.Client.Create(ctx, app)).Should(Succeed())

				expectedKustomization := kustomizev1.KustomizationSpec{
					// Flux adds a prepending `./` to path arguments that doesn't already have it.
					// https://github.com/fluxcd/flux2/blob/ca496d393d993ac5119ed84f83e010b8fe918c53/cmd/flux/create_kustomization.go#L115
					Path: "./" + req.Path,
					// Flux kustomization default; I couldn't find an export default from the package.
					Interval: metav1.Duration{Duration: time.Duration(1 * time.Minute)},
					Prune:    true,
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Name: req.Name,
						Kind: sourcev1.GitRepositoryKind,
					},
					Force: false,
				}

				repoURL, err := gitproviders.NewRepoURL(sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				expectedSrc := sourcev1.GitRepositorySpec{
					URL: req.Url,
					SecretRef: &meta.LocalObjectReference{
						Name: automation.CreateRepoSecretName(repoURL).String(),
					},
					Interval: metav1.Duration{Duration: 30 * time.Second},
					Reference: &sourcev1.GitRepositoryRef{
						Branch: req.Branch,
					},
					Ignore: helpers.GetIgnoreSpec(),
				}

				expected := helpers.GenerateExpectedFS(req, root, clusterName, expectedApp, expectedKustomization, expectedSrc)

				diff, err := helpers.DiffFS(actual, expected)
				if err != nil {
					GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
				}

				removeRequest := &pb.RemoveApplicationRequest{
					Name:      appName,
					Namespace: namespace.Name,
					AutoMerge: true,
				}

				removeResponse, err := client.RemoveApplication(ctx, removeRequest)
				Expect(err).NotTo(HaveOccurred())

				Expect(removeResponse.Success).To(BeTrue())

				commits, _, err = gh.Repositories.ListCommits(ctx, githubOrg, sourceRepoName, &ghAPI.CommitsListOptions{SHA: "main"})
				Expect(err).NotTo(HaveOccurred())
				Expect(commits).To(HaveLen(4))

				appAddCommit = commits[0]

				c, _, err = gh.Repositories.GetCommit(ctx, githubOrg, sourceRepoName, *appAddCommit.SHA)
				Expect(err).NotTo(HaveOccurred())

				fs = helpers.MakeWeGOFS(root, req.Name, clusterName)

				actual, err = helpers.GetFileContents(ctx, gh, githubOrg, sourceRepoName, fs, c.Files)
				Expect(err).NotTo(HaveOccurred())

				expected = map[string]interface{}{
					filepath.Join(root, "clusters", clusterName, "user", "kustomization.yaml"): &types.Kustomization{
						TypeMeta: types.TypeMeta{
							Kind:       types.KustomizationKind,
							APIVersion: types.KustomizationVersion,
						},
						MetaData: &types.ObjectMeta{
							Name:      appName,
							Namespace: namespace.Name,
						},
					},
				}

				diff, err = helpers.DiffFS(actual, expected)
				if err != nil {
					GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
				}

			})

			It("with an external config repo", func() {
				sourceRepoURL := fmt.Sprintf("https://github.com/%s/%s", githubOrg, sourceRepoName)
				configRepoURL := fmt.Sprintf("https://github.com/%s/%s", githubOrg, configRepoName)

				configRepo, configRef, err := helpers.CreateRepo(ctx, gp, configRepoURL)
				Expect(err).NotTo(HaveOccurred())

				defer func() { Expect(configRepo.Delete(ctx)).To(Succeed()) }()

				sourceRepo, sourceRef, err := helpers.CreatePopulatedSourceRepo(ctx, gp, sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				defer func() { Expect(sourceRepo.Delete(ctx)).To(Succeed()) }()

				req := &pb.AddApplicationRequest{
					Name:      "my-app",
					Namespace: namespace.Name,
					Url:       sourceRef.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
					Branch:    "main",
					Path:      "k8s/overlays/development",
					ConfigUrl: configRef.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
					AutoMerge: true,
				}

				res, err := client.AddApplication(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				Expect(res.Success).To(BeTrue(), "request should have been successful")

				_, err = sourceRepo.DeployKeys().Get(ctx, gitproviders.DeployKeyName)
				Expect(err).NotTo(HaveOccurred(), "deploy key should have been found")

				prs, err := configRepo.PullRequests().List(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(prs).To(HaveLen(0))

				root := helpers.ExternalConfigRoot

				fs := helpers.MakeWeGOFS(root, req.Name, clusterName)

				commits, _, err := gh.Repositories.ListCommits(ctx, githubOrg, configRepoName, &ghAPI.CommitsListOptions{SHA: "main"})
				Expect(err).NotTo(HaveOccurred())
				Expect(commits).To(HaveLen(2))

				appAddCommit := commits[0]

				c, _, err := gh.Repositories.GetCommit(ctx, githubOrg, configRepoName, *appAddCommit.SHA)
				Expect(err).NotTo(HaveOccurred())

				actual, err := helpers.GetFileContents(ctx, gh, githubOrg, configRepoName, fs, c.Files)
				Expect(err).NotTo(HaveOccurred())

				expectedApp := wego.ApplicationSpec{
					URL:            req.Url,
					Branch:         req.Branch,
					Path:           req.Path,
					ConfigURL:      req.ConfigUrl,
					DeploymentType: wego.DeploymentTypeKustomize,
					SourceType:     wego.SourceTypeGit,
				}

				expectedKustomization := kustomizev1.KustomizationSpec{
					// Flux adds a prepending `./` to path arguments that doesn't already have it.
					// https://github.com/fluxcd/flux2/blob/ca496d393d993ac5119ed84f83e010b8fe918c53/cmd/flux/create_kustomization.go#L115
					Path: "./" + req.Path,
					// Flux kustomization default; I couldn't find an export default from the package.
					Interval: metav1.Duration{Duration: time.Duration(1 * time.Minute)},
					Prune:    true,
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Name: req.Name,
						Kind: sourcev1.GitRepositoryKind,
					},
					Force: false,
				}

				repoURL, err := gitproviders.NewRepoURL(sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				expectedSrc := sourcev1.GitRepositorySpec{
					URL: req.Url,
					SecretRef: &meta.LocalObjectReference{
						Name: automation.CreateRepoSecretName(repoURL).String(),
					},
					Interval: metav1.Duration{Duration: 30 * time.Second},
					Reference: &sourcev1.GitRepositoryRef{
						Branch: req.Branch,
					},
					Ignore: helpers.GetIgnoreSpec(),
				}

				expected := helpers.GenerateExpectedFS(req, root, clusterName, expectedApp, expectedKustomization, expectedSrc)

				diff, err := helpers.DiffFS(actual, expected)
				if err != nil {
					GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
				}
			})
		})
	})

	Context("GitLab", func() {
		var gitlabGroup string
		var gitlabToken string
		var gitlabProviderClient gitprovider.Client
		var gl *glAPI.Client
		BeforeEach(func() {
			gitlabGroup = os.Getenv("GITLAB_ORG")
			gitlabToken = os.Getenv("GITLAB_TOKEN")
			ctx = middleware.ContextWithGRPCAuth(context.Background(), gitlabToken)
			gitlabProviderClient, err = gitlab.NewClient(
				gitlabToken,
				"oauth2",
				gitprovider.WithDestructiveAPICalls(true),
			)

			gl, err = glAPI.NewClient(gitlabToken)
			Expect(err).NotTo(HaveOccurred())
		})
		Context("via merge request", func() {
			It("adds with no config repo specified", func() {
				sourceRepoURL := fmt.Sprintf("https://gitlab.com/%s/%s", gitlabGroup, sourceRepoName)

				repo, ref, err := helpers.CreatePopulatedSourceRepo(ctx, gitlabProviderClient, sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				Expect(err).NotTo(HaveOccurred())

				defer func() { Expect(repo.Delete(ctx)).To(Succeed()) }()

				req := &pb.AddApplicationRequest{
					Name:      "my-app",
					Namespace: namespace.Name,
					Url:       ref.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
					Branch:    "main",
					Path:      "k8s/overlays/development",
				}

				res, err := client.AddApplication(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				Expect(res.Success).To(BeTrue())

				_, err = repo.DeployKeys().Get(ctx, gitproviders.DeployKeyName)
				Expect(err).NotTo(HaveOccurred(), "deploy key should have been found")

				prs, err := repo.PullRequests().List(ctx)
				Expect(err).NotTo(HaveOccurred())

				Expect(prs).To(HaveLen(1))

				root := helpers.InAppRoot

				fs := helpers.MakeWeGOFS(root, req.Name, clusterName)

				gl, err := helpers.NewFileFetcher(gitproviders.GitProviderGitLab, gitlabToken)
				Expect(err).NotTo(HaveOccurred())

				actual, err := gl.GetFilesForPullRequest(ctx, 1, gitlabGroup, sourceRepoName, fs)
				Expect(err).NotTo(HaveOccurred())

				expectedKustomization := kustomizev1.KustomizationSpec{
					// Flux adds a prepending `./` to path arguments that doesn't already have it.
					// https://github.com/fluxcd/flux2/blob/ca496d393d993ac5119ed84f83e010b8fe918c53/cmd/flux/create_kustomization.go#L115
					Path: "./" + req.Path,
					// Flux kustomization default; I couldn't find an export default from the package.
					Interval: metav1.Duration{Duration: time.Duration(1 * time.Minute)},
					Prune:    true,
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Name: req.Name,
						Kind: sourcev1.GitRepositoryKind,
					},
					Force: false,
				}

				repoURL, err := gitproviders.NewRepoURL(sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				expectedSource := sourcev1.GitRepositorySpec{
					URL: req.Url,
					SecretRef: &meta.LocalObjectReference{
						Name: automation.CreateRepoSecretName(repoURL).String(),
					},
					Interval: metav1.Duration{Duration: time.Duration(30 * time.Second)},
					Reference: &sourcev1.GitRepositoryRef{
						Branch: req.Branch,
					},
					Ignore: helpers.GetIgnoreSpec(),
				}

				expectedApp := wego.ApplicationSpec{
					URL:            req.Url,
					Branch:         req.Branch,
					Path:           req.Path,
					ConfigURL:      req.Url,
					DeploymentType: wego.DeploymentTypeKustomize,
					SourceType:     wego.SourceTypeGit,
				}

				expected := helpers.GenerateExpectedFS(req, root, clusterName, expectedApp, expectedKustomization, expectedSource)

				diff, err := helpers.DiffFS(actual, expected)
				if err != nil {
					GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
				}
			})
			It("adds an app with an external config repo", func() {

				gp, err := gitlab.NewClient(
					gitlabToken,
					"oauth2",
					gitprovider.WithDestructiveAPICalls(true),
				)
				Expect(err).NotTo(HaveOccurred())

				sourceRepoURL := fmt.Sprintf("https://gitlab.com/%s/%s", gitlabGroup, sourceRepoName)
				configRepoURL := fmt.Sprintf("https://gitlab.com/%s/%s", gitlabGroup, configRepoName)

				configRepo, configRef, err := helpers.CreateRepo(ctx, gp, configRepoURL)
				Expect(err).NotTo(HaveOccurred())

				defer func() { Expect(configRepo.Delete(ctx)).To(Succeed()) }()

				sourceRepo, sourceRef, err := helpers.CreatePopulatedSourceRepo(ctx, gp, sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				defer func() { Expect(sourceRepo.Delete(ctx)).To(Succeed()) }()

				req := &pb.AddApplicationRequest{
					Name:      "my-app",
					Namespace: namespace.Name,
					Url:       sourceRef.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
					Branch:    "main",
					Path:      "k8s/overlays/development",
					ConfigUrl: configRef.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
				}

				res, err := client.AddApplication(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.Success).To(BeTrue(), "request should have been successful")

				_, err = configRepo.DeployKeys().Get(ctx, gitproviders.DeployKeyName)
				Expect(err).NotTo(HaveOccurred(), "deploy key should have been found")

				prs, err := configRepo.PullRequests().List(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(prs).To(HaveLen(1))

				root := helpers.ExternalConfigRoot
				fs := helpers.MakeWeGOFS(root, req.Name, clusterName)

				fetcher, err := helpers.NewFileFetcher(gitproviders.GitProviderGitLab, gitlabToken)
				Expect(err).NotTo(HaveOccurred())

				actual, err := fetcher.GetFilesForPullRequest(ctx, 1, gitlabGroup, configRepoName, fs)
				Expect(err).NotTo(HaveOccurred())

				normalizedUrl, err := gitproviders.NewRepoURL(req.Url)
				Expect(err).NotTo(HaveOccurred())

				expectedApp := wego.ApplicationSpec{
					URL:            normalizedUrl.String(),
					Branch:         req.Branch,
					Path:           req.Path,
					ConfigURL:      req.ConfigUrl,
					DeploymentType: wego.DeploymentTypeKustomize,
					SourceType:     wego.SourceTypeGit,
				}

				expectedKustomization := kustomizev1.KustomizationSpec{
					Path:     "./" + req.Path,
					Interval: metav1.Duration{Duration: time.Duration(1 * time.Minute)},
					Prune:    true,
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Name: req.Name,
						Kind: sourcev1.GitRepositoryKind,
					},
					Force: false,
				}

				repoURL, err := gitproviders.NewRepoURL(sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				expectedSource := sourcev1.GitRepositorySpec{
					URL: req.Url,
					SecretRef: &meta.LocalObjectReference{
						// Might be a bug? Should be configRepoURL?
						Name: automation.CreateRepoSecretName(repoURL).String(),
					},
					Interval: metav1.Duration{Duration: time.Duration(30 * time.Second)},
					Reference: &sourcev1.GitRepositoryRef{
						Branch: req.Branch,
					},
					Ignore: helpers.GetIgnoreSpec(),
				}

				expected := helpers.GenerateExpectedFS(req, root, clusterName, expectedApp, expectedKustomization, expectedSource)

				diff, err := helpers.DiffFS(actual, expected)
				if err != nil {
					GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
				}
			})
			It("remove app with config repo specified", func() {
				sourceRepoURL := fmt.Sprintf("https://gitlab.com/%s/%s", gitlabGroup, sourceRepoName)

				repo, ref, err := helpers.CreatePopulatedSourceRepo(ctx, gitlabProviderClient, sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				defer func() { Expect(repo.Delete(ctx)).To(Succeed()) }()

				appName := "my-app"

				req := &pb.AddApplicationRequest{
					Name:      appName,
					Namespace: namespace.Name,
					Url:       ref.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
					Branch:    "main",
					Path:      "k8s/overlays/development",
					AutoMerge: true,
				}

				res, err := client.AddApplication(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				Expect(res.Success).To(BeTrue(), "request should have been successful")

				_, err = repo.DeployKeys().Get(ctx, gitproviders.DeployKeyName)
				Expect(err).NotTo(HaveOccurred(), "deploy key should have been found")

				prs, err := repo.PullRequests().List(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(prs).To(HaveLen(0))

				root := helpers.InAppRoot

				fs := helpers.MakeWeGOFS(root, req.Name, clusterName)

				fullRepoPath := fmt.Sprintf("%s/%s", ref.OrganizationRef.Organization, sourceRepoName)

				branch := "main"
				commits, _, err := gl.Commits.ListCommits(fullRepoPath, &glAPI.ListCommitsOptions{
					RefName: &branch,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(commits).To(HaveLen(3))

				appAddCommit := commits[0]

				diffs, _, err := gl.Commits.GetCommitDiff(fullRepoPath, appAddCommit.ID, &glAPI.GetCommitDiffOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(err).NotTo(HaveOccurred())

				actual, err := helpers.GetGitlabFilesContents(gl, fullRepoPath, fs, appAddCommit.ID, diffs)
				Expect(err).NotTo(HaveOccurred())

				expectedApp := wego.ApplicationSpec{
					URL:            req.Url,
					Branch:         req.Branch,
					Path:           req.Path,
					ConfigURL:      ref.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
					DeploymentType: wego.DeploymentTypeKustomize,
					SourceType:     wego.SourceTypeGit,
				}
				app := &wego.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      appName,
						Namespace: namespace.Name,
					},
					Spec: expectedApp,
				}

				Expect(env.Client.Create(ctx, app)).Should(Succeed())
				//
				expectedKustomization := kustomizev1.KustomizationSpec{
					// Flux adds a prepending `./` to path arguments that doesn't already have it.
					// https://github.com/fluxcd/flux2/blob/ca496d393d993ac5119ed84f83e010b8fe918c53/cmd/flux/create_kustomization.go#L115
					Path: "./" + req.Path,
					// Flux kustomization default; I couldn't find an export default from the package.
					Interval: metav1.Duration{Duration: time.Duration(1 * time.Minute)},
					Prune:    true,
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Name: req.Name,
						Kind: sourcev1.GitRepositoryKind,
					},
					Force: false,
				}

				repoURL, err := gitproviders.NewRepoURL(sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				expectedSrc := sourcev1.GitRepositorySpec{
					URL: req.Url,
					SecretRef: &meta.LocalObjectReference{
						Name: automation.CreateRepoSecretName(repoURL).String(),
					},
					Interval: metav1.Duration{Duration: 30 * time.Second},
					Reference: &sourcev1.GitRepositoryRef{
						Branch: req.Branch,
					},
					Ignore: helpers.GetIgnoreSpec(),
				}

				expected := helpers.GenerateExpectedFS(req, root, clusterName, expectedApp, expectedKustomization, expectedSrc)

				diff, err := helpers.DiffFS(actual, expected)
				if err != nil {
					GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
				}

				removeRequest := &pb.RemoveApplicationRequest{
					Name:      appName,
					Namespace: namespace.Name,
					AutoMerge: false,
				}

				removeResponse, err := client.RemoveApplication(ctx, removeRequest)
				Expect(err).NotTo(HaveOccurred())

				Expect(removeResponse.Success).To(BeTrue())

				prs, err = repo.PullRequests().List(ctx)
				Expect(err).NotTo(HaveOccurred())

				Expect(prs).To(HaveLen(1))

				fs = helpers.MakeWeGOFS(root, req.Name, clusterName)

				fetcher, err := helpers.NewFileFetcher(gitproviders.GitProviderGitLab, gitlabToken)
				Expect(err).NotTo(HaveOccurred())

				actual, err = fetcher.GetFilesForPullRequest(ctx, 1, gitlabGroup, sourceRepoName, fs)
				Expect(err).NotTo(HaveOccurred())

				expected = map[string]interface{}{
					filepath.Join(root, "clusters", clusterName, "user", "kustomization.yaml"): &types.Kustomization{
						TypeMeta: types.TypeMeta{
							Kind:       types.KustomizationKind,
							APIVersion: types.KustomizationVersion,
						},
						MetaData: &types.ObjectMeta{
							Name:      appName,
							Namespace: namespace.Name,
						},
					},
				}

				diff, err = helpers.DiffFS(actual, expected)
				if err != nil {
					GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
				}

			})
		})
		Context("via auto merge", func() {
			It("add/remove with config repo specified", func() {
				sourceRepoURL := fmt.Sprintf("https://gitlab.com/%s/%s", gitlabGroup, sourceRepoName)

				repo, ref, err := helpers.CreatePopulatedSourceRepo(ctx, gitlabProviderClient, sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				defer func() { Expect(repo.Delete(ctx)).To(Succeed()) }()

				appName := "my-app"

				req := &pb.AddApplicationRequest{
					Name:      appName,
					Namespace: namespace.Name,
					Url:       ref.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
					Branch:    "main",
					Path:      "k8s/overlays/development",
					AutoMerge: true,
				}

				res, err := client.AddApplication(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				Expect(res.Success).To(BeTrue(), "request should have been successful")

				_, err = repo.DeployKeys().Get(ctx, gitproviders.DeployKeyName)
				Expect(err).NotTo(HaveOccurred(), "deploy key should have been found")

				prs, err := repo.PullRequests().List(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(prs).To(HaveLen(0))

				root := helpers.InAppRoot

				fs := helpers.MakeWeGOFS(root, req.Name, clusterName)

				fullRepoPath := fmt.Sprintf("%s/%s", ref.OrganizationRef.Organization, sourceRepoName)

				branch := "main"
				commits, _, err := gl.Commits.ListCommits(fullRepoPath, &glAPI.ListCommitsOptions{
					RefName: &branch,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(commits).To(HaveLen(3))

				appAddCommit := commits[0]

				diffs, _, err := gl.Commits.GetCommitDiff(fullRepoPath, appAddCommit.ID, &glAPI.GetCommitDiffOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(err).NotTo(HaveOccurred())

				actual, err := helpers.GetGitlabFilesContents(gl, fullRepoPath, fs, appAddCommit.ID, diffs)
				Expect(err).NotTo(HaveOccurred())

				expectedApp := wego.ApplicationSpec{
					URL:            req.Url,
					Branch:         req.Branch,
					Path:           req.Path,
					ConfigURL:      ref.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
					DeploymentType: wego.DeploymentTypeKustomize,
					SourceType:     wego.SourceTypeGit,
				}
				app := &wego.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      appName,
						Namespace: namespace.Name,
					},
					Spec: expectedApp,
				}

				Expect(env.Client.Create(ctx, app)).Should(Succeed())
				//
				expectedKustomization := kustomizev1.KustomizationSpec{
					// Flux adds a prepending `./` to path arguments that doesn't already have it.
					// https://github.com/fluxcd/flux2/blob/ca496d393d993ac5119ed84f83e010b8fe918c53/cmd/flux/create_kustomization.go#L115
					Path: "./" + req.Path,
					// Flux kustomization default; I couldn't find an export default from the package.
					Interval: metav1.Duration{Duration: time.Duration(1 * time.Minute)},
					Prune:    true,
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Name: req.Name,
						Kind: sourcev1.GitRepositoryKind,
					},
					Force: false,
				}

				repoURL, err := gitproviders.NewRepoURL(sourceRepoURL)
				Expect(err).NotTo(HaveOccurred())

				expectedSrc := sourcev1.GitRepositorySpec{
					URL: req.Url,
					SecretRef: &meta.LocalObjectReference{
						Name: automation.CreateRepoSecretName(repoURL).String(),
					},
					Interval: metav1.Duration{Duration: 30 * time.Second},
					Reference: &sourcev1.GitRepositoryRef{
						Branch: req.Branch,
					},
					Ignore: helpers.GetIgnoreSpec(),
				}

				expected := helpers.GenerateExpectedFS(req, root, clusterName, expectedApp, expectedKustomization, expectedSrc)

				diff, err := helpers.DiffFS(actual, expected)
				if err != nil {
					GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
				}

				removeRequest := &pb.RemoveApplicationRequest{
					Name:      appName,
					Namespace: namespace.Name,
					AutoMerge: true,
				}

				removeResponse, err := client.RemoveApplication(ctx, removeRequest)
				Expect(err).NotTo(HaveOccurred())

				Expect(removeResponse.Success).To(BeTrue())

				commits, _, err = gl.Commits.ListCommits(fullRepoPath, &glAPI.ListCommitsOptions{
					RefName: &branch,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(commits).To(HaveLen(4))

				appAddCommit = commits[0]

				diffs, _, err = gl.Commits.GetCommitDiff(fullRepoPath, appAddCommit.ID, &glAPI.GetCommitDiffOptions{})
				Expect(err).NotTo(HaveOccurred())

				fs = helpers.MakeWeGOFS(root, req.Name, clusterName)

				actual, err = helpers.GetGitlabFilesContents(gl, fullRepoPath, fs, appAddCommit.ID, diffs)
				Expect(err).NotTo(HaveOccurred())

				expected = map[string]interface{}{
					filepath.Join(root, "clusters", clusterName, "user", "kustomization.yaml"): &types.Kustomization{
						TypeMeta: types.TypeMeta{
							Kind:       types.KustomizationKind,
							APIVersion: types.KustomizationVersion,
						},
						MetaData: &types.ObjectMeta{
							Name:      appName,
							Namespace: namespace.Name,
						},
					},
				}

				diff, err = helpers.DiffFS(actual, expected)
				if err != nil {
					GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
				}

			})
		})
	})

})
