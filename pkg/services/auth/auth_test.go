package auth

import (
	"context"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

type actualFluxRunner struct {
	runner.Runner
}

func (r *actualFluxRunner) Run(command string, args ...string) ([]byte, error) {
	cmd := "../../flux/bin/flux"

	return r.Runner.Run(cmd, args...)
}

var _ = Describe("auth", func() {
	var namespace *corev1.Namespace
	testClustername := "test-cluster"
	repoUrlString := "ssh://git@github.com/my-org/my-repo.git"
	configRepoUrl, err := gitproviders.NewRepoURL(repoUrlString, true)
	Expect(err).NotTo(HaveOccurred())
	repoURL, err := gitproviders.NewRepoURL(repoUrlString, false)
	Expect(err).NotTo(HaveOccurred())

	BeforeEach(func() {
		namespace = &corev1.Namespace{}
		namespace.Name = "kube-test-" + rand.String(5)
		Expect(k8sClient.Create(context.Background(), namespace)).To(Succeed())
	})
	Describe("AuthService", func() {
		var (
			ctx        context.Context
			secretName automation.GeneratedSecretName
			gp         gitprovidersfakes.FakeGitProvider
			osysClient osys.Osys
			as         AuthService
			fluxClient flux.Flux
		)
		BeforeEach(func() {
			ctx = context.Background()
			secretName = automation.CreateRepoSecretName(configRepoUrl)
			Expect(err).NotTo(HaveOccurred())
			osysClient = osys.New()
			gp = gitprovidersfakes.FakeGitProvider{}
			gp.GetRepoVisibilityReturns(gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPrivate), nil)
			fluxClient = flux.New(osysClient, &actualFluxRunner{Runner: &runner.CLIRunner{}})

			as = &authSvc{
				logger:      &loggerfakes.FakeLogger{}, //Stay silent in tests.
				fluxClient:  fluxClient,
				k8sClient:   k8sClient,
				gitProvider: &gp,
			}
		})
		It("create and stores a deploy key if none exists", func() {
			_, err := as.CreateGitClient(ctx, repoURL, testClustername, namespace.Name, false, nil)
			Expect(err).NotTo(HaveOccurred())

			sn := SecretName{Name: secretName, Namespace: namespace.Name}
			secret := &corev1.Secret{}
			Expect(k8sClient.Get(ctx, sn.NamespacedName(), secret)).To(Succeed())

			Expect(secret.StringData["identity"]).NotTo(BeNil())
			Expect(secret.StringData["identity.pub"]).NotTo(BeNil())
		})
		It("doesn't create a deploy key when dry-run is true", func() {
			_, err := as.CreateGitClient(ctx, repoURL, testClustername, namespace.Name, true, nil)
			Expect(err).NotTo(HaveOccurred())

			sn := SecretName{Name: secretName, Namespace: namespace.Name}
			secret := &corev1.Secret{}
			Expect(k8sClient.Get(ctx, sn.NamespacedName(), secret)).To(HaveOccurred())
		})
		It("uses an existing deploy key when present", func() {
			gp.DeployKeyExistsReturns(true, nil)
			sn := SecretName{Name: secretName, Namespace: namespace.Name}
			// using `generateDeployKey` as a helper for the test setup.
			_, secret, err := (&authSvc{fluxClient: fluxClient}).generateRepoAuthSecret(testClustername, sn, repoURL, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sClient.Create(ctx, secret)).To(Succeed())

			_, err = as.CreateGitClient(ctx, repoURL, testClustername, namespace.Name, false, nil)
			Expect(err).NotTo(HaveOccurred())

			// We should NOT have uploaded anything since the key already exists
			Expect(gp.UploadDeployKeyCallCount()).To(Equal(0))
		})
		It("handles the case where a deploy key exists on the provider, but not the cluster", func() {
			gp.DeployKeyExistsReturns(true, nil)
			sn := SecretName{Name: secretName, Namespace: namespace.Name}

			_, err = as.CreateGitClient(ctx, repoURL, testClustername, namespace.Name, false, nil)
			Expect(err).NotTo(HaveOccurred())

			newSecret := &corev1.Secret{}
			Expect(k8sClient.Get(ctx, sn.NamespacedName(), newSecret)).To(Succeed())
			Expect(gp.UploadDeployKeyCallCount()).To(Equal(1))
		})

		It("avoids deploying key for non config repos", func() {
			gp.GetRepoVisibilityReturns(gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPublic), nil)

			_, err = as.CreateGitClient(ctx, repoURL, testClustername, namespace.Name, false, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(gp.UploadDeployKeyCallCount()).To(Equal(0))
		})

		It("does not create a deploy key when using https", func() {
			repoURL, err = gitproviders.NewRepoURL("https://github.com/my-org/my-repo.git", false)
			Expect(err).NotTo(HaveOccurred())
			_, err := as.CreateGitClient(ctx, repoURL, testClustername, namespace.Name, false, &flux.HTTPSCreds{Username: "test", Password: "topsecret"})
			Expect(err).NotTo(HaveOccurred())

			sn := SecretName{Name: secretName, Namespace: namespace.Name}
			secret := &corev1.Secret{}
			Expect(k8sClient.Get(ctx, sn.NamespacedName(), secret)).To(Succeed())
			Expect(secret.StringData["password"]).To(Equal("topsecret"))
			Expect(secret.StringData["username"]).To(Equal("test"))
		})
	})
})
