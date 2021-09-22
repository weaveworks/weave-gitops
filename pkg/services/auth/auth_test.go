package auth

import (
	"bytes"
	"context"
	"io"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/osys/osysfakes"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/app"

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
	repoUrl, err := gitproviders.NewNormalizedRepoURL(repoUrlString)
	Expect(err).NotTo(HaveOccurred())
	BeforeEach(func() {
		namespace = &corev1.Namespace{}
		namespace.Name = "kube-test-" + rand.String(5)

		Expect(k8sClient.Create(context.Background(), namespace)).To(Succeed())
	})
	Describe("AuthService", func() {
		var (
			ctx        context.Context
			secretName app.GeneratedSecretName
			gp         gitprovidersfakes.FakeGitProvider
			osysClient *osys.OsysClient
			as         AuthService
			fluxClient flux.Flux
		)
		BeforeEach(func() {
			ctx = context.Background()
			secretName = app.CreateRepoSecretName(testClustername, repoUrl.String())
			Expect(err).NotTo(HaveOccurred())
			osysClient = osys.New()
			gp = gitprovidersfakes.FakeGitProvider{}
			fluxClient = flux.New(osysClient, &actualFluxRunner{Runner: &runner.CLIRunner{}})

			as = &authSvc{
				logger:      logger.NewCLILogger(bytes.NewBuffer([]byte{})), //Stay silent in tests.
				fluxClient:  fluxClient,
				k8sClient:   k8sClient,
				gitProvider: &gp,
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
		It("create and stores a deploy key if none exists", func() {
			_, err := as.CreateGitClient(ctx, testClustername, namespace.Name, repoUrl.String())
			Expect(err).NotTo(HaveOccurred())
			sn := SecretName{Name: secretName, Namespace: namespace.Name}
			secret := &corev1.Secret{}
			Expect(k8sClient.Get(ctx, sn.NamespacedName(), secret)).To(Succeed())

			Expect(secret.Data["identity"]).NotTo(BeNil())
			Expect(secret.Data["identity.pub"]).NotTo(BeNil())
		})
		It("uses an existing deploy key when present", func() {
			gp.DeployKeyExistsStub = func(s1, s2 string) (bool, error) {
				return true, nil
			}
			sn := SecretName{Name: secretName, Namespace: namespace.Name}
			// using `generateDeployKey` as a helper for the test setup.
			_, secret, err := (&authSvc{fluxClient: fluxClient}).generateDeployKey(testClustername, sn, repoUrl)
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sClient.Create(ctx, secret)).To(Succeed())

			_, err = as.CreateGitClient(ctx, testClustername, namespace.Name, repoUrl.String())
			Expect(err).NotTo(HaveOccurred())
			// We should NOT have uploaded anything since the key already exists
			Expect(gp.UploadDeployKeyCallCount()).To(Equal(0))
		})
		It("handles the case where a deploy key exists on the provider, but not the cluster", func() {
			gp.DeployKeyExistsStub = func(s1, s2 string) (bool, error) {
				return true, nil
			}
			sn := SecretName{Name: secretName, Namespace: namespace.Name}

			_, err = as.CreateGitClient(ctx, testClustername, namespace.Name, repoUrl.String())
			Expect(err).NotTo(HaveOccurred())

			newSecret := &corev1.Secret{}
			Expect(k8sClient.Get(ctx, sn.NamespacedName(), newSecret)).To(Succeed())
			Expect(gp.UploadDeployKeyCallCount()).To(Equal(1))
		})
		Context("auth token informational message", func() {
			var osysClient *osysfakes.FakeOsys
			var logger *loggerfakes.FakeLogger
			var authHandler BlockingCLIAuthHandler

			BeforeEach(func() {
				authHandler = func(_ context.Context, _ io.Writer) (string, error) {
					return "a-token", nil
				}
				logger = &loggerfakes.FakeLogger{
					WarningfStub: func(fmtArg string, restArgs ...interface{}) {},
				}
			})

			It("generates an error if an invalid provider name is passed", func() {
				osysClient = &osysfakes.FakeOsys{}
				_, err := getGitProviderWithClients(context.Background(), gitproviders.GitProviderName("badname"), osysClient, authHandler, logger)
				Expect(err.Error()).To(ContainSubstring(`unknown git provider: "badname"`))
			})

			Context("informs the user that she can use a token for auth", func() {
				BeforeEach(func() {
					osysClient = &osysfakes.FakeOsys{
						GetGitProviderTokenStub: func(tokenVarName string) (string, error) {
							return "", osys.ErrNoGitProviderTokenSet
						},
					}
				})

				DescribeTable("generates correct token info messages", func(providerName gitproviders.GitProviderName, msgArg string) {
					_, err := getGitProviderWithClients(context.Background(), providerName, osysClient, authHandler, logger)
					Expect(err).ShouldNot(HaveOccurred())
					fmtArg, restArgs := logger.WarningfArgsForCall(0)
					Expect(fmtArg).Should(Equal("Setting the %q environment variable to a valid token will allow ongoing use of the CLI without requiring a browser-based auth flow...\n"))
					Expect(restArgs[0]).Should(Equal(msgArg))
				},
					Entry("token for GitHub", gitproviders.GitProviderGitHub, "GITHUB_TOKEN"),
					Entry("token for GitLab", gitproviders.GitProviderGitLab, "GITLAB_TOKEN"))
			})

			Context("displays no message if token is set", func() {
				BeforeEach(func() {
					osysClient = &osysfakes.FakeOsys{
						GetGitProviderTokenStub: func(tokenVarName string) (string, error) {
							return "token", nil
						},
					}
				})

				DescribeTable("generates no message if token set", func(providerName gitproviders.GitProviderName) {
					_, err := getGitProviderWithClients(context.Background(), providerName, osysClient, authHandler, logger)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(logger.WarningfCallCount()).To(Equal(0))
				},
					Entry("GitHub", gitproviders.GitProviderGitHub),
					Entry("GitLab", gitproviders.GitProviderGitHub))
			})
		})
	})
})
