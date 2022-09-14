package auth_test

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/pkg/names"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

var _ = Describe("auth", func() {
	var (
		configRepoURL gitproviders.RepoURL
		namespace     *corev1.Namespace
		ctx           context.Context
		secretName    names.GeneratedSecretName
		gp            gitprovidersfakes.FakeGitProvider
		as            auth.AuthService
		fluxClient    flux.Flux
	)

	BeforeEach(func() {
		repoURLString := "ssh://git@github.com/my-org/my-repo.git"

		var err error
		configRepoURL, err = gitproviders.NewRepoURL(repoURLString)
		Expect(err).NotTo(HaveOccurred())

		namespace = &corev1.Namespace{}
		namespace.Name = "kube-test-" + rand.String(5)
		Expect(k8sClient.Create(context.Background(), namespace)).To(Succeed())

		ctx = context.Background()
		secretName = names.CreateRepoSecretName(configRepoURL)
		Expect(err).NotTo(HaveOccurred())
		gp = gitprovidersfakes.FakeGitProvider{}
		gp.GetRepoVisibilityReturns(gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPrivate), nil)
		fluxClient = flux.New(&actualFluxRunner{Runner: &runner.CLIRunner{}})

		as = auth.NewAuthService(fluxClient, k8sClient, &gp, logr.Discard())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(context.Background(), namespace)).To(Succeed())
	})

	It("create and stores a deploy key if none exists", func() {
		_, err := as.CreateGitClient(ctx, configRepoURL, namespace.Name, false)
		Expect(err).NotTo(HaveOccurred())
		sn := auth.SecretName{Name: secretName, Namespace: namespace.Name}
		secret := &corev1.Secret{}
		Expect(k8sClient.Get(ctx, sn.NamespacedName(), secret)).To(Succeed())

		Expect(secret.StringData["identity"]).NotTo(BeNil())
		Expect(secret.StringData["identity.pub"]).NotTo(BeNil())
	})

	It("doesn't create a deploy key when dry-run is true", func() {
		_, err := as.CreateGitClient(ctx, configRepoURL, namespace.Name, true)
		Expect(err).NotTo(HaveOccurred())
		sn := auth.SecretName{Name: secretName, Namespace: namespace.Name}
		secret := &corev1.Secret{}
		Expect(k8sClient.Get(ctx, sn.NamespacedName(), secret)).To(HaveOccurred())
	})

	When("a deploy key already exists", func() {
		var original int

		BeforeEach(func() {
			_, err := as.CreateGitClient(ctx, configRepoURL, namespace.Name, false)
			Expect(err).NotTo(HaveOccurred())

			original = gp.UploadDeployKeyCallCount()
		})

		It("does not create a new one", func() {
			// the CallCounts do not reset within tests, so we sanity check it here
			Expect(gp.UploadDeployKeyCallCount()).To(Equal(original))

			gp.DeployKeyExistsReturns(true, nil)

			_, err := as.CreateGitClient(ctx, configRepoURL, namespace.Name, false)
			Expect(err).NotTo(HaveOccurred())

			// We should NOT have uploaded anything since the key already exists
			Expect(gp.UploadDeployKeyCallCount()).To(Equal(original))
		})
	})

	It("handles the case where a deploy key exists on the provider, but not the cluster", func() {
		gp.DeployKeyExistsReturns(true, nil)
		sn := auth.SecretName{Name: secretName, Namespace: namespace.Name}

		_, err := as.CreateGitClient(ctx, configRepoURL, namespace.Name, false)
		Expect(err).NotTo(HaveOccurred())

		newSecret := &corev1.Secret{}
		Expect(k8sClient.Get(ctx, sn.NamespacedName(), newSecret)).To(Succeed())
		Expect(gp.UploadDeployKeyCallCount()).To(Equal(1))
	})
})

type actualFluxRunner struct {
	runner.Runner
}

func (r *actualFluxRunner) Run(command string, args ...string) ([]byte, error) {
	cmd := "flux"

	return r.Runner.Run(cmd, args...)
}
