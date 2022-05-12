package services

import (
	"context"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

var _ = Describe("Services factory", func() {
	var ctx context.Context
	var fakeFlux *fluxfakes.FakeFlux
	fakeLog := logr.Discard()
	var fakeClient *gitprovidersfakes.FakeClient
	var fakeKube *kube.KubeHTTP
	var factory Factory

	BeforeEach(func() {
		ctx = context.Background()
		fakeFlux = &fluxfakes.FakeFlux{}
		fakeClient = &gitprovidersfakes.FakeClient{}
		fakeKube = &kube.KubeHTTP{}

		factory = NewFactory(fakeFlux, fakeLog)
	})

	Describe("get git clients", func() {
		It("all parameter fields are empty throws error", func() {
			gitClient, gitProvider, err := factory.GetGitClients(ctx, fakeKube, fakeClient, GitConfigParams{})

			Expect(gitClient).To(BeNil())
			Expect(gitProvider).To(BeNil())
			Expect(err.Error()).To(MatchRegexp("error normalizing config url*."))
		})

		It("config type user repo and empty url return error", func() {
			gitClient, gitProvider, err := factory.GetGitClients(ctx, fakeKube, fakeClient, GitConfigParams{
				ConfigRepo:       "",
				IsHelmRepository: false,
			})

			Expect(gitClient).To(BeNil())
			Expect(gitProvider).To(BeNil())
			Expect(err.Error()).To(MatchRegexp("error normalizing config url*."))
		})
	})
})
