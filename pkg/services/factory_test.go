package services

import (
	"context"
	"github.com/weaveworks/weave-gitops/pkg/services/app"

	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
)

var _ = Describe("Services factory", func() {
	var ctx context.Context
	var fakeFlux *fluxfakes.FakeFlux
	var fakeLog *loggerfakes.FakeLogger
	var fakeClient *gitprovidersfakes.FakeClient
	var factory Factory
	var fakeKube = &kubefakes.FakeKube{}

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
			gitClient, gitProvider, err := factory.GetGitClients(ctx, fakeClient, fakeKube, nil, GitConfigParams{})

			Expect(gitClient).To(BeNil())
			Expect(gitProvider).To(BeNil())
			Expect(err.Error()).To(MatchRegexp("error normalizing url*."))
		})

		It("config is for a helm repository", func() {
			gitClient, gitProvider, err := factory.GetGitClients(ctx, fakeClient, fakeKube, nil, GitConfigParams{
				IsHelmRepository: true,
			})

			Expect(gitClient).To(BeNil())
			Expect(gitProvider).To(BeNil())
			Expect(err).To(BeNil())
		})

		It("app add params is helm chart", func() {
			params := app.AddParams{Chart: "this-chart"}
			gitClient, gitProvider, err := factory.GetGitClients(ctx, fakeClient, fakeKube, nil, GitConfigParams{
				IsHelmRepository: params.IsHelmRepository(),
			})

			Expect(gitClient).To(BeNil())
			Expect(gitProvider).To(BeNil())
			Expect(err).To(BeNil())
		})

		It("config type none and empty url return error", func() {
			gitClient, gitProvider, err := factory.GetGitClients(ctx, fakeClient, fakeKube, nil, GitConfigParams{
				ConfigURL:        string(app.ConfigTypeNone),
				IsHelmRepository: false,
			})

			Expect(gitClient).To(BeNil())
			Expect(gitProvider).To(BeNil())
			Expect(err.Error()).To(MatchRegexp("error normalizing url*."))
		})

		It("config type user repo and empty url return error", func() {
			gitClient, gitProvider, err := factory.GetGitClients(ctx, fakeClient, fakeKube, nil, GitConfigParams{
				ConfigURL:        string(app.ConfigTypeUserRepo),
				IsHelmRepository: false,
			})

			Expect(gitClient).To(BeNil())
			Expect(gitProvider).To(BeNil())
			Expect(err.Error()).To(MatchRegexp("error normalizing url*."))
		})
	})
})
