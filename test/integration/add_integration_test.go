package app_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fluxcd/go-git-providers/gitprovider"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
)

var _ = Describe("Add() integration tests", func() {
	var (
		appSvc      app.AppService
		l           logger.Logger
		g           git.Git
		gp          *gitprovidersfakes.FakeGitProvider
		f           flux.Flux
		o           *testOsys
		kubeClient  kube.Kube
		pwd         string
		userHomeDir string
		namespace   *corev1.Namespace
	)

	BeforeEach(func() {
		namespace = &corev1.Namespace{}
		namespace.Name = "kube-test-" + rand.String(5)

		Expect(k.Create(context.Background(), namespace)).To(Succeed())

		// Make some temp dirs so we don't touch the user's data
		userHomeDir, err := ioutil.TempDir("", "user-home-*")
		Expect(err).NotTo(HaveOccurred())
		pwd, err = ioutil.TempDir("", "app-dir-*")
		Expect(err).NotTo(HaveOccurred())

		g = git.New(nil)
		_, err = g.Init(pwd, "ssh://git@github.com/myorg/myrepo.git", "main")
		Expect(err).NotTo(HaveOccurred())

		err = g.Write(fmt.Sprintf("%s/readme.txt", pwd), []byte("Hello World."))
		Expect(err).NotTo(HaveOccurred())

		g.Commit(git.Commit{
			Author:  git.Author{Email: "user@example.com"},
			Message: "Initial Commit",
		})

		o = &testOsys{Osys: osys.New(), userHome: userHomeDir}
		l = logger.NewCLILogger(bytes.NewBuffer([]byte{}))

		gp = &gitprovidersfakes.FakeGitProvider{}
		f = flux.New(o, &actualFluxRunner{Runner: &runner.CLIRunner{}})
		kubeClient = &kube.KubeHTTP{Client: k, ClusterName: "test-cluster"}

		appSvc = &app.App{
			Osys:   o,
			Logger: l,
			Git:    g,
			Flux:   f,
			Kube:   kubeClient,
			GitProviderFactory: func(token string) (gitproviders.GitProvider, error) {
				return gp, nil
			},
		}
	})

	AfterEach(func() {
		os.RemoveAll(pwd)
		os.RemoveAll(userHomeDir)
	})

	It(".Add() a kustomize deployment in directory", func() {
		ctx := context.Background()
		params := app.AddParams{
			Dir:              pwd,
			Name:             "myapp",
			GitProviderToken: "sometoken",
			Namespace:        namespace.Name,
			DeploymentType:   string(wego.DeploymentTypeKustomize),
		}
		gp.GetDefaultBranchStub = func(url string) (string, error) {
			return "main", nil
		}

		gp.CreatePullRequestToOrgRepoStub = func(orr gitprovider.OrgRepositoryRef, s1, s2 string, cf []gitprovider.CommitFile, s3, s4, s5 string) (gitprovider.PullRequest, error) {
			return &gitprovidersfakes.FakePullRequest{}, nil
		}

		gp.CreatePullRequestToUserRepoStub = func(urr gitprovider.UserRepositoryRef, s1, s2 string, cf []gitprovider.CommitFile, s3, s4, s5 string) (gitprovider.PullRequest, error) {
			return &gitprovidersfakes.FakePullRequest{}, nil
		}

		err := appSvc.Add(params)
		Expect(err).NotTo(HaveOccurred())

		By(".git directory should exist", func() {
			_, err = ioutil.ReadDir(fmt.Sprintf("%s/.git", pwd))
			Expect(err).NotTo(HaveOccurred())
		})
		By("A pull request should have been created.", func() {
			// A pull request should have been created.
			Expect(gp.CreatePullRequestToUserRepoCallCount()).To(BeNumerically(">", 0))
		})
		By("a secret should exist on the cluster", func() {
			secret := corev1.Secret{}
			secretName := types.NamespacedName{Name: "wego-test-cluster-myrepo", Namespace: params.Namespace}
			err = k.Get(ctx, secretName, &secret)
			Expect(err).NotTo(HaveOccurred())
			Expect(secret.Name).To(Equal(secretName.Name))
		})
		By("a source should exist on the cluster", func() {
			source := sourcev1.GitRepository{}
			sourceName := types.NamespacedName{Name: params.Name, Namespace: params.Namespace}
			err = k.Get(ctx, sourceName, &source)
			Expect(err).NotTo(HaveOccurred())
			Expect(source.Name).To(Equal(sourceName.Name))
		})
		By("a kustomization should exist on the cluster", func() {
			kustomization := kustomizev1.Kustomization{}
			kustomizationName := types.NamespacedName{Name: "myapp-apps-dir", Namespace: params.Namespace}
			err = k.Get(ctx, kustomizationName, &kustomization)
			Expect(err).NotTo(HaveOccurred())
			Expect(kustomization.Name).To(Equal(kustomizationName.Name))
		})
	})
})
