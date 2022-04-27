package auth

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/gitops/pkg/names"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/gitops/pkg/runner"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

type actualFluxRunner struct {
	runner.Runner
}

func (r *actualFluxRunner) Run(command string, args ...string) ([]byte, error) {
	cmd := "../../../../tools/bin/flux"

	return r.Runner.Run(cmd, args...)
}

var _ = Describe("auth", func() {
	var namespace *corev1.Namespace
	repoUrlString := "ssh://git@github.com/my-org/my-repo.git"
	configRepoUrl, err := gitproviders.NewRepoURL(repoUrlString)
	Expect(err).NotTo(HaveOccurred())
	repoUrl, err := gitproviders.NewRepoURL(repoUrlString)
	Expect(err).NotTo(HaveOccurred())

	BeforeEach(func() {
		namespace = &corev1.Namespace{}
		namespace.Name = "kube-test-" + rand.String(5)
		Expect(k8sClient.Create(context.Background(), namespace)).To(Succeed())
	})
	Describe("AuthService", func() {
		var (
			ctx        context.Context
			secretName names.GeneratedSecretName
			gp         gitprovidersfakes.FakeGitProvider
			as         AuthService
			fluxClient flux.Flux
		)
		BeforeEach(func() {
			ctx = context.Background()
			secretName = names.CreateRepoSecretName(configRepoUrl)
			Expect(err).NotTo(HaveOccurred())
			gp = gitprovidersfakes.FakeGitProvider{}
			gp.GetRepoVisibilityReturns(gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPrivate), nil)
			fluxClient = flux.New(&actualFluxRunner{Runner: &runner.CLIRunner{}})

			as = &authSvc{
				log:         logr.Discard(), //Stay silent in tests.
				fluxClient:  fluxClient,
				k8sClient:   k8sClient,
				gitProvider: &gp,
			}
		})
		It("create and stores a deploy key if none exists", func() {
			_, err := as.CreateGitClient(ctx, configRepoUrl, namespace.Name, false)
			Expect(err).NotTo(HaveOccurred())
			sn := SecretName{Name: secretName, Namespace: namespace.Name}
			secret := &corev1.Secret{}
			Expect(k8sClient.Get(ctx, sn.NamespacedName(), secret)).To(Succeed())

			Expect(secret.StringData["identity"]).NotTo(BeNil())
			Expect(secret.StringData["identity.pub"]).NotTo(BeNil())
		})
		It("doesn't create a deploy key when dry-run is true", func() {
			_, err := as.CreateGitClient(ctx, configRepoUrl, namespace.Name, true)
			Expect(err).NotTo(HaveOccurred())
			sn := SecretName{Name: secretName, Namespace: namespace.Name}
			secret := &corev1.Secret{}
			Expect(k8sClient.Get(ctx, sn.NamespacedName(), secret)).To(HaveOccurred())
		})
		It("uses an existing deploy key when present", func() {
			gp.DeployKeyExistsReturns(true, nil)
			sn := SecretName{Name: secretName, Namespace: namespace.Name}
			// using `generateDeployKey` as a helper for the test setup.
			_, secret, err := (&authSvc{fluxClient: fluxClient}).generateDeployKey(sn, repoUrl)
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sClient.Create(ctx, secret)).To(Succeed())

			_, err = as.CreateGitClient(ctx, configRepoUrl, namespace.Name, false)
			Expect(err).NotTo(HaveOccurred())
			// We should NOT have uploaded anything since the key already exists
			Expect(gp.UploadDeployKeyCallCount()).To(Equal(0))
		})

		It("handles the case where a deploy key exists on the provider, but not the cluster", func() {
			gp.DeployKeyExistsReturns(true, nil)
			sn := SecretName{Name: secretName, Namespace: namespace.Name}

			_, err = as.CreateGitClient(ctx, configRepoUrl, namespace.Name, false)
			Expect(err).NotTo(HaveOccurred())

			newSecret := &corev1.Secret{}
			Expect(k8sClient.Get(ctx, sn.NamespacedName(), newSecret)).To(Succeed())
			Expect(gp.UploadDeployKeyCallCount()).To(Equal(1))
		})

	})
})
