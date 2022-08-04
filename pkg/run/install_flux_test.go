package run_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/server"
	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/run"
	v1 "k8s.io/api/core/v1"
)

const (
	testVersion = "some-version"
)

var _ = Describe("GetFluxVersion", func() {
	var fakeLogger *loggerfakes.FakeLogger

	BeforeEach(func() {
		fakeLogger = &loggerfakes.FakeLogger{}
	})

	It("gets flux version", func() {
		kubeClientOpts := run.GetKubeClientOptions()

		contextName := "some-context"

		kubeClient, err := run.GetKubeClient(fakeLogger, contextName, k8sEnv.Rest, kubeClientOpts)
		Expect(err).NotTo(HaveOccurred())

		ctx := context.Background()

		fluxNs := &v1.Namespace{}
		fluxNs.Name = "flux-ns-test"
		fluxNs.Labels = map[string]string{
			coretypes.PartOfLabel: server.FluxNamespacePartOf,
			flux.VersionLabelKey:  testVersion,
		}

		Expect(kubeClient.Create(ctx, fluxNs)).To(Succeed())

		fluxVersion, err := run.GetFluxVersion(fakeLogger, ctx, kubeClient)

		Expect(err).NotTo(HaveOccurred())
		Expect(fluxVersion).To(Equal(testVersion))
	})
})
