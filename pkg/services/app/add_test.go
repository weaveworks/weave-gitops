package app_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
)

var (
	appSrv        app.AppService
	gitClient     *gitfakes.FakeGit
	fluxClient    *fluxfakes.FakeFlux
	kubeClient    *kubefakes.FakeKube
	defaultParams app.AddParams
)

var _ = BeforeEach(func() {
	gitClient = &gitfakes.FakeGit{}
	fluxClient = &fluxfakes.FakeFlux{}
	kubeClient = &kubefakes.FakeKube{
		GetClusterNameStub: func() (string, error) {
			return "test-cluster", nil
		},
		GetClusterStatusStub: func() kube.ClusterStatus {
			return kube.WeGOInstalled
		},
	}

	deps := &app.Dependencies{
		Git:  gitClient,
		Flux: fluxClient,
		Kube: kubeClient,
	}

	appSrv = app.New(deps)

	defaultParams = app.AddParams{
		Url:    "https://github.com/foo/bar",
		Branch: "main",
		Dir:    ".",
	}
})

var _ = Describe("Add", func() {
	It("checks for cluster status", func() {
		err := appSrv.Add(defaultParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(kubeClient.GetClusterStatusCallCount()).To(Equal(1))
	})
})
