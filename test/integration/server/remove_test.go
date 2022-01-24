//go:build !unittest
// +build !unittest

package server_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fluxcd/go-git-providers/gitlab"
	glAPI "github.com/xanzy/go-gitlab"

	ghAPI "github.com/google/go-github/v32/github"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"github.com/weaveworks/weave-gitops/test/integration/server/helpers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/kustomize/api/types"
)

var _ = Describe("RemoveApplication", func() {
	var (
		namespace      *corev1.Namespace
		ctx            context.Context
		sourceRepoName string
		client         pb.ApplicationsClient
	)

	BeforeEach(func() {
		namespace = &corev1.Namespace{}
		namespace.Name = "kube-test-" + rand.String(5)
		Expect(env.Client.Create(context.Background(), namespace)).To(Succeed())
		sourceRepoName = "test-source-repo-" + rand.String(5)
		client = pb.NewApplicationsClient(conn)
	})

	Context("Github", func() {
		var gh *ghAPI.Client
		var gp gitprovider.Client
		var githubOrg = "weaveworks-gitops-test"
		var githubToken = os.Getenv("GITHUB_TOKEN")
		var sourceRepoURL string
		var sourceRepo gitprovider.OrgRepository
		var sourceRef *gitprovider.OrgRepositoryRef
		var addAppRequest *pb.AddApplicationRequest
		var appName string
		BeforeEach(func() {
			gh = helpers.NewGithubClient(ctx, githubToken)
			ctx = middleware.ContextWithGRPCAuth(context.Background(), githubToken)
			Expect(err).NotTo(HaveOccurred())
			gp, err = github.NewClient(
				gitprovider.WithDestructiveAPICalls(true),
				gitprovider.WithOAuth2Token(os.Getenv("GITHUB_TOKEN")),
			)
			Expect(err).NotTo(HaveOccurred())
			sourceRepoURL = fmt.Sprintf("https://github.com/%s/%s", githubOrg, sourceRepoName)
			Expect(helpers.SetWegoConfig(env.Client, namespace.Name, sourceRepoURL)).To(Succeed())
			sourceRepo, sourceRef, err = helpers.CreatePopulatedSourceRepo(ctx, gp, sourceRepoURL)
			Expect(err).NotTo(HaveOccurred())

			appName = "my-app"

			addAppRequest = &pb.AddApplicationRequest{
				Name:      appName,
				Namespace: namespace.Name,
				Url:       sourceRef.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
				Branch:    "main",
				Path:      "k8s/overlays/development",
			}

		})
		Context("via pull request", func() {
			It("remove app using the same app as config repo", func() {

				defer func() { Expect(sourceRepo.Delete(ctx)).To(Succeed()) }()

				addAppRequest.AutoMerge = true
				addAppRequest.ConfigRepo = sourceRef.GetCloneURL(gitprovider.TransportTypeSSH) + ".git"

				res, err := client.AddApplication(ctx, addAppRequest)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.Success).To(BeTrue(), "request should have been successful")

				expectedApp := wego.ApplicationSpec{
					URL:            addAppRequest.Url,
					Branch:         addAppRequest.Branch,
					Path:           addAppRequest.Path,
					ConfigRepo:     sourceRef.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
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

				root := helpers.InAppRoot

				fs := helpers.MakeWeGOFS(root, addAppRequest.Name, clusterName)

				commits, _, err := gh.Repositories.ListCommits(ctx, githubOrg, sourceRepoName, &ghAPI.CommitsListOptions{SHA: "main"})
				Expect(err).NotTo(HaveOccurred())

				appAddCommit := commits[0]

				c, _, err := gh.Repositories.GetCommit(ctx, githubOrg, sourceRepoName, *appAddCommit.SHA)
				Expect(err).NotTo(HaveOccurred())

				actual, err := helpers.GetGithubFilesContents(ctx, gh, githubOrg, sourceRepoName, fs, c.Files)
				Expect(err).NotTo(HaveOccurred())

				kustomizationPath := filepath.Join(root, "clusters", clusterName, "user", "kustomization.yaml")

				expectedKustomizationBefore := helpers.WeGODirectoryFS{
					kustomizationPath: &types.Kustomization{
						TypeMeta: types.TypeMeta{
							Kind:       types.KustomizationKind,
							APIVersion: types.KustomizationVersion,
						},
						MetaData: &types.ObjectMeta{
							Name:      clusterName,
							Namespace: namespace.Name,
						},
						Resources: []string{filepath.Join("../../../apps/", appName)},
					},
				}

				diff, err := helpers.DiffFS(helpers.WeGODirectoryFS{
					kustomizationPath: actual[kustomizationPath],
				}, expectedKustomizationBefore)
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

				prs, err := sourceRepo.PullRequests().List(ctx)
				Expect(err).NotTo(HaveOccurred())

				Expect(prs).To(HaveLen(1))

				fs = helpers.MakeWeGOFS(root, addAppRequest.Name, clusterName)

				fetcher, err := helpers.NewFileFetcher(gitproviders.GitProviderGitHub, githubToken)
				Expect(err).NotTo(HaveOccurred())

				actual, err = fetcher.GetFilesForPullRequest(ctx, 1, githubOrg, sourceRepoName, fs)
				Expect(err).NotTo(HaveOccurred())

				path := filepath.Join(root, "clusters", clusterName, "user", "kustomization.yaml")

				// Assert that we have no resources
				afterK := &types.Kustomization{
					TypeMeta: types.TypeMeta{
						Kind:       types.KustomizationKind,
						APIVersion: types.KustomizationVersion,
					},
					MetaData: &types.ObjectMeta{
						Name:      clusterName,
						Namespace: namespace.Name,
					},
				}

				Expect(len(afterK.Resources)).To(Equal(0))

				expectedKustomization := map[string]interface{}{
					path: &types.Kustomization{
						TypeMeta: types.TypeMeta{
							Kind:       types.KustomizationKind,
							APIVersion: types.KustomizationVersion,
						},
						MetaData: &types.ObjectMeta{
							Name:      clusterName,
							Namespace: namespace.Name,
						},
					},
				}

				diff, err = helpers.DiffFS(actual, expectedKustomization)
				if err != nil {
					GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
				}

			})
		})

	})
	Context("Gitlab", func() {
		var gitlabGroup string
		var gitlabToken string
		var gitlabProviderClient gitprovider.Client
		var gitlabAPIClient *glAPI.Client
		var sourceRepoURL string
		var sourceRepo gitprovider.OrgRepository
		var sourceRef *gitprovider.OrgRepositoryRef
		var addAppRequest *pb.AddApplicationRequest
		var appName string
		BeforeEach(func() {
			gitlabGroup = os.Getenv("GITLAB_ORG")
			gitlabToken = os.Getenv("GITLAB_TOKEN")
			ctx = middleware.ContextWithGRPCAuth(context.Background(), gitlabToken)
			gitlabProviderClient, err = gitlab.NewClient(
				gitlabToken,
				"oauth2",
				gitprovider.WithDestructiveAPICalls(true),
			)

			gitlabAPIClient, err = glAPI.NewClient(gitlabToken)
			Expect(err).NotTo(HaveOccurred())
			sourceRepoURL = fmt.Sprintf("https://gitlab.com/%s/%s", gitlabGroup, sourceRepoName)
			Expect(helpers.SetWegoConfig(env.Client, namespace.Name, sourceRepoURL)).To(Succeed())
			sourceRepo, sourceRef, err = helpers.CreatePopulatedSourceRepo(ctx, gitlabProviderClient, sourceRepoURL)
			Expect(err).NotTo(HaveOccurred())

			appName = "my-app"

			addAppRequest = &pb.AddApplicationRequest{
				Name:      appName,
				Namespace: namespace.Name,
				Url:       sourceRef.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
				Branch:    "main",
				Path:      "k8s/overlays/development",
			}
		})
		Context("via merge request", func() {
			It("remove app with no config repo", func() {

				defer func() { Expect(sourceRepo.Delete(ctx)).To(Succeed()) }()

				addAppRequest.AutoMerge = true
				addAppRequest.ConfigRepo = sourceRef.GetCloneURL(gitprovider.TransportTypeSSH) + ".git"

				res, err := client.AddApplication(ctx, addAppRequest)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.Success).To(BeTrue(), "request should have been successful")

				expectedApp := wego.ApplicationSpec{
					URL:            addAppRequest.Url,
					Branch:         addAppRequest.Branch,
					Path:           addAppRequest.Path,
					ConfigRepo:     sourceRef.GetCloneURL(gitprovider.TransportTypeSSH) + ".git",
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

				root := helpers.InAppRoot

				fs := helpers.MakeWeGOFS(root, addAppRequest.Name, clusterName)

				fullRepoPath := fmt.Sprintf("%s/%s", sourceRef.OrganizationRef.Organization, sourceRepoName)

				branch := "main"
				commits, _, err := gitlabAPIClient.Commits.ListCommits(fullRepoPath, &glAPI.ListCommitsOptions{
					RefName: &branch,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(commits).To(HaveLen(3))

				appAddCommit := commits[0]

				diffs, _, err := gitlabAPIClient.Commits.GetCommitDiff(fullRepoPath, appAddCommit.ID, &glAPI.GetCommitDiffOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(err).NotTo(HaveOccurred())

				actual, err := helpers.GetGitlabFilesContents(gitlabAPIClient, fullRepoPath, fs, appAddCommit.ID, diffs)
				Expect(err).NotTo(HaveOccurred())

				kustomizationPath := filepath.Join(root, "clusters", clusterName, "user", "kustomization.yaml")

				expectedKustomizationBefore := helpers.WeGODirectoryFS{
					kustomizationPath: &types.Kustomization{
						TypeMeta: types.TypeMeta{
							Kind:       types.KustomizationKind,
							APIVersion: types.KustomizationVersion,
						},
						MetaData: &types.ObjectMeta{
							Name:      clusterName,
							Namespace: namespace.Name,
						},
						Resources: []string{filepath.Join("../../../apps/", appName)},
					},
				}

				diff, err := helpers.DiffFS(helpers.WeGODirectoryFS{
					kustomizationPath: actual[kustomizationPath],
				}, expectedKustomizationBefore)
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

				prs, err := sourceRepo.PullRequests().List(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(prs).To(HaveLen(1))

				fs = helpers.MakeWeGOFS(root, addAppRequest.Name, clusterName)

				fetcher, err := helpers.NewFileFetcher(gitproviders.GitProviderGitLab, gitlabToken)
				Expect(err).NotTo(HaveOccurred())

				actual, err = fetcher.GetFilesForPullRequest(ctx, 1, gitlabGroup, sourceRepoName, fs)
				Expect(err).NotTo(HaveOccurred())

				expected := map[string]interface{}{
					filepath.Join(root, "clusters", clusterName, "user", "kustomization.yaml"): &types.Kustomization{
						TypeMeta: types.TypeMeta{
							Kind:       types.KustomizationKind,
							APIVersion: types.KustomizationVersion,
						},
						MetaData: &types.ObjectMeta{
							Name:      clusterName,
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
