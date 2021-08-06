package auth

import (
	"bytes"
	"context"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
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
	repoUrl := "ssh://git@github.com/my-org/my-repo.git"
	BeforeEach(func() {
		namespace = &corev1.Namespace{}
		namespace.Name = "kube-test-" + rand.String(5)

		Expect(k8sClient.Create(context.Background(), namespace)).To(Succeed())
	})
	Describe("AuthService", func() {
		var (
			ctx        context.Context
			secretName string
			gp         gitprovidersfakes.FakeGitProvider
			osysClient *osys.OsysClient
			as         *AuthService
			name       types.NamespacedName
			fluxClient flux.Flux
		)
		BeforeEach(func() {
			ctx = context.Background()
			secretName = utils.CreateAppSecretName(testClustername, repoUrl)
			osysClient = osys.New()
			gp = gitprovidersfakes.FakeGitProvider{}
			fluxClient = flux.New(osysClient, &actualFluxRunner{Runner: &runner.CLIRunner{}})

			as = &AuthService{
				logger:      logger.NewCLILogger(bytes.NewBuffer([]byte{})), //Stay silent in tests.
				fluxClient:  fluxClient,
				k8sClient:   k8sClient,
				gitProvider: &gp,
			}
			name = types.NamespacedName{
				Name:      "my-repo",
				Namespace: namespace.Name,
			}

			gp.GetAccountTypeStub = func(s string) (gitproviders.ProviderAccountType, error) {
				return gitproviders.AccountTypeOrg, nil
			}

			gp.GetRepoInfoStub = func(pat gitproviders.ProviderAccountType, s1, s2 string) (*gitprovider.RepositoryInfo, error) {
				return &gitprovider.RepositoryInfo{
					Visibility: gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityInternal),
				}, nil
			}
		})
		It("returns an error if deploy keys are not set up", func() {
			_, err := as.GitClient()
			Expect(err).To(MatchError(ErrNoDeployKeysSetup))
		})
		It("create and stores a deploy key if none exists", func() {
			Expect(as.SetupDeployKey(ctx, name, testClustername, repoUrl)).To(Succeed())
			sn := types.NamespacedName{Name: secretName, Namespace: namespace.Name}
			secret := &corev1.Secret{}
			Expect(k8sClient.Get(ctx, sn, secret)).To(Succeed())

			Expect(secret.Data["identity"]).NotTo(BeNil())
			Expect(secret.Data["identity.pub"]).NotTo(BeNil())
			_, err := as.GitClient()
			Expect(err).NotTo(HaveOccurred())
		})
		It("uses an existing deploy key when present", func() {
			gp.DeployKeyExistsStub = func(s1, s2 string) (bool, error) {
				return true, nil
			}
			sn := types.NamespacedName{Name: secretName, Namespace: namespace.Name}
			// using `generateDeployKey` as a helper for the test setup.
			_, secret, err := (&AuthService{fluxClient: fluxClient}).generateDeployKey(testClustername, sn, repoUrl)
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sClient.Create(ctx, secret)).To(Succeed())

			Expect(as.SetupDeployKey(ctx, name, testClustername, repoUrl)).To(Succeed())
			_, err = as.GitClient()
			Expect(err).NotTo(HaveOccurred())
			// We should NOT have uploaded anything since the key already exists
			Expect(gp.UploadDeployKeyCallCount()).To(Equal(0))
		})
	})
})
