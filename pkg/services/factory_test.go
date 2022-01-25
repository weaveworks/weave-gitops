package services

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
)

var _ = Describe("Services factory", func() {
	var ctx context.Context
	var fakeFlux *fluxfakes.FakeFlux
	var fakeLog *loggerfakes.FakeLogger
	var fakeClient *gitprovidersfakes.FakeClient
	var fakeKube *kubefakes.FakeKube
	var factory Factory

	BeforeEach(func() {
		ctx = context.Background()
		fakeFlux = &fluxfakes.FakeFlux{}
		fakeClient = &gitprovidersfakes.FakeClient{}
		fakeLog = &loggerfakes.FakeLogger{}
		fakeKube = &kubefakes.FakeKube{}

		factory = NewFactory(fakeFlux, fakeLog)
	})

	Describe("get git clients", func() {
		It("all parameter fields are empty throws error", func() {
			gitClient, gitProvider, err := factory.GetGitClients(ctx, fakeKube, fakeClient, GitConfigParams{})

			Expect(gitClient).To(BeNil())
			Expect(gitProvider).To(BeNil())
			Expect(err.Error()).To(MatchRegexp("error normalizing url*."))
		})

		It("config is for a helm repository", func() {
			gitClient, gitProvider, err := factory.GetGitClients(ctx, fakeKube, fakeClient, GitConfigParams{
				IsHelmRepository: true,
			})

			Expect(gitClient).To(BeNil())
			Expect(gitProvider).To(BeNil())
			Expect(err).To(BeNil())
		})

		It("app add params is helm chart", func() {
			params := app.AddParams{Chart: "this-chart"}
			gitClient, gitProvider, err := factory.GetGitClients(ctx, fakeKube, fakeClient, GitConfigParams{
				IsHelmRepository: params.IsHelmRepository(),
			})

			Expect(gitClient).To(BeNil())
			Expect(gitProvider).To(BeNil())
			Expect(err).To(BeNil())
		})

		It("config type user repo and empty url return error", func() {
			gitClient, gitProvider, err := factory.GetGitClients(ctx, fakeKube, fakeClient, GitConfigParams{
				ConfigRepo:       "",
				IsHelmRepository: false,
			})

			Expect(gitClient).To(BeNil())
			Expect(gitProvider).To(BeNil())
			Expect(err.Error()).To(MatchRegexp("error normalizing url*."))
		})
	})
})
