//go:build !unittest
// +build !unittest

package server_test

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/fluxcd/go-git-providers/gitprovider"
	kustomizev2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/google/go-github/v32/github"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/test/integration/server/helpers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

var _ = Describe("AddApplication", func() {
	var (
		namespace      *corev1.Namespace
		ctx            context.Context
		token          = os.Getenv("GITHUB_TOKEN")
		sourceRepoName string
		configRepoName string
		gh             *github.Client = helpers.NewGithubClient(ctx, token)
		client         pb.ApplicationsClient
	)

	BeforeEach(func() {
		namespace = &corev1.Namespace{}
		namespace.Name = "kube-test-" + rand.String(5)
		Expect(env.Client.Create(context.Background(), namespace)).To(Succeed())
		ctx = middleware.ContextWithGRPCAuth(context.Background(), token)
		sourceRepoName = "test-source-repo-" + rand.String(5)
		configRepoName = "test-config-repo-" + rand.String(5)
		client = pb.NewApplicationsClient(conn)

	})
	Context("via pull request", func() {
		It("adds with no config repo specified", func() {
			sourceRepoURL := fmt.Sprintf("https://github.com/%s/%s", org, sourceRepoName)

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

			actual, err := helpers.GetFilesForPullRequest(ctx, gh, org, sourceRepoName, fs)
			Expect(err).NotTo(HaveOccurred())

			expectedKustomization := kustomizev2.KustomizationSpec{
				// Flux adds a prepending `./` to path arguments that doen't already have it.
				// https://github.com/fluxcd/flux2/blob/ca496d393d993ac5119ed84f83e010b8fe918c53/cmd/flux/create_kustomization.go#L115
				Path: "./" + req.Path,
				// Flux kustomization default; I couldn't find an export default from the package.
				Interval: metav1.Duration{Duration: time.Duration(1 * time.Minute)},
				Prune:    true,
				SourceRef: kustomizev2.CrossNamespaceSourceReference{
					Name: req.Name,
					Kind: sourcev1.GitRepositoryKind,
				},
				Force: false,
			}

			expectedSource := sourcev1.GitRepositorySpec{
				URL: req.Url,
				SecretRef: &meta.LocalObjectReference{
					Name: app.CreateRepoSecretName(clusterName, sourceRepoURL).String(),
				},
				Interval: metav1.Duration{Duration: time.Duration(30 * time.Second)},
				Reference: &sourcev1.GitRepositoryRef{
					Branch: req.Branch,
				},
			}

			expectedApp := wego.ApplicationSpec{
				URL:            req.Url,
				Branch:         req.Branch,
				Path:           req.Path,
				ConfigURL:      "",
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
			sourceRepoURL := fmt.Sprintf("https://github.com/%s/%s", org, sourceRepoName)
			configRepoURL := fmt.Sprintf("https://github.com/%s/%s", org, configRepoName)

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

			actual, err := helpers.GetFilesForPullRequest(ctx, gh, org, configRepoName, fs)
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

			expectedKustomization := kustomizev2.KustomizationSpec{
				Path:     "./" + req.Path,
				Interval: metav1.Duration{Duration: time.Duration(1 * time.Minute)},
				Prune:    true,
				SourceRef: kustomizev2.CrossNamespaceSourceReference{
					Name: req.Name,
					Kind: sourcev1.GitRepositoryKind,
				},
				Force: false,
			}

			expectedSource := sourcev1.GitRepositorySpec{
				URL: req.Url,
				SecretRef: &meta.LocalObjectReference{
					// Might be a bug? Should be configRepoURL?
					Name: app.CreateRepoSecretName(clusterName, sourceRepoURL).String(),
				},
				Interval: metav1.Duration{Duration: time.Duration(30 * time.Second)},
				Reference: &sourcev1.GitRepositoryRef{
					Branch: req.Branch,
				},
			}

			expected := helpers.GenerateExpectedFS(req, root, clusterName, expectedApp, expectedKustomization, expectedSource)

			diff, err := helpers.DiffFS(actual, expected)
			if err != nil {
				GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
			}
		})
	})
	Context("via auto merge", func() {
		It("adds with no config repo specified", func() {
			sourceRepoURL := fmt.Sprintf("https://github.com/%s/%s", org, sourceRepoName)

			repo, ref, err := helpers.CreatePopulatedSourceRepo(ctx, gp, sourceRepoURL)
			Expect(err).NotTo(HaveOccurred())

			defer func() { Expect(repo.Delete(ctx)).To(Succeed()) }()

			req := &pb.AddApplicationRequest{
				Name:      "my-app",
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

			commits, _, err := gh.Repositories.ListCommits(ctx, org, sourceRepoName, &github.CommitsListOptions{SHA: "main"})
			Expect(err).NotTo(HaveOccurred())
			Expect(commits).To(HaveLen(3))

			appAddCommit := commits[0]

			c, _, err := gh.Repositories.GetCommit(ctx, org, sourceRepoName, *appAddCommit.SHA)
			Expect(err).NotTo(HaveOccurred())

			actual, err := helpers.GetFileContents(ctx, gh, org, sourceRepoName, fs, c.Files)
			Expect(err).NotTo(HaveOccurred())

			expectedApp := wego.ApplicationSpec{
				URL:            req.Url,
				Branch:         req.Branch,
				Path:           req.Path,
				ConfigURL:      "",
				DeploymentType: wego.DeploymentTypeKustomize,
				SourceType:     wego.SourceTypeGit,
			}

			expectedKustomization := kustomizev2.KustomizationSpec{
				// Flux adds a prepending `./` to path arguments that doen't already have it.
				// https://github.com/fluxcd/flux2/blob/ca496d393d993ac5119ed84f83e010b8fe918c53/cmd/flux/create_kustomization.go#L115
				Path: "./" + req.Path,
				// Flux kustomization default; I couldn't find an export default from the package.
				Interval: metav1.Duration{Duration: time.Duration(1 * time.Minute)},
				Prune:    true,
				SourceRef: kustomizev2.CrossNamespaceSourceReference{
					Name: req.Name,
					Kind: sourcev1.GitRepositoryKind,
				},
				Force: false,
			}

			expectedSrc := sourcev1.GitRepositorySpec{
				URL: req.Url,
				SecretRef: &meta.LocalObjectReference{
					Name: app.CreateRepoSecretName(clusterName, sourceRepoURL).String(),
				},
				Interval: metav1.Duration{Duration: time.Duration(30 * time.Second)},
				Reference: &sourcev1.GitRepositoryRef{
					Branch: req.Branch,
				},
			}

			expected := helpers.GenerateExpectedFS(req, root, clusterName, expectedApp, expectedKustomization, expectedSrc)

			diff, err := helpers.DiffFS(actual, expected)
			if err != nil {
				GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
			}
		})
		It("with an external config repo", func() {
			sourceRepoURL := fmt.Sprintf("https://github.com/%s/%s", org, sourceRepoName)
			configRepoURL := fmt.Sprintf("https://github.com/%s/%s", org, configRepoName)

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

			commits, _, err := gh.Repositories.ListCommits(ctx, org, configRepoName, &github.CommitsListOptions{SHA: "main"})
			Expect(err).NotTo(HaveOccurred())
			Expect(commits).To(HaveLen(2))

			appAddCommit := commits[0]

			c, _, err := gh.Repositories.GetCommit(ctx, org, configRepoName, *appAddCommit.SHA)
			Expect(err).NotTo(HaveOccurred())

			actual, err := helpers.GetFileContents(ctx, gh, org, configRepoName, fs, c.Files)
			Expect(err).NotTo(HaveOccurred())

			expectedApp := wego.ApplicationSpec{
				URL:            req.Url,
				Branch:         req.Branch,
				Path:           req.Path,
				ConfigURL:      req.ConfigUrl,
				DeploymentType: wego.DeploymentTypeKustomize,
				SourceType:     wego.SourceTypeGit,
			}

			expectedKustomization := kustomizev2.KustomizationSpec{
				// Flux adds a prepending `./` to path arguments that doen't already have it.
				// https://github.com/fluxcd/flux2/blob/ca496d393d993ac5119ed84f83e010b8fe918c53/cmd/flux/create_kustomization.go#L115
				Path: "./" + req.Path,
				// Flux kustomization default; I couldn't find an export default from the package.
				Interval: metav1.Duration{Duration: time.Duration(1 * time.Minute)},
				Prune:    true,
				SourceRef: kustomizev2.CrossNamespaceSourceReference{
					Name: req.Name,
					Kind: sourcev1.GitRepositoryKind,
				},
				Force: false,
			}

			expectedSrc := sourcev1.GitRepositorySpec{
				URL: req.Url,
				SecretRef: &meta.LocalObjectReference{
					Name: app.CreateRepoSecretName(clusterName, sourceRepoURL).String(),
				},
				Interval: metav1.Duration{Duration: time.Duration(30 * time.Second)},
				Reference: &sourcev1.GitRepositoryRef{
					Branch: req.Branch,
				},
			}

			expected := helpers.GenerateExpectedFS(req, root, clusterName, expectedApp, expectedKustomization, expectedSrc)

			diff, err := helpers.DiffFS(actual, expected)
			if err != nil {
				GinkgoT().Errorf("%s: (-actual +expected): %s\n", err.Error(), diff)
			}
		})
	})
})
