package run_test

import (
	"context"

	"github.com/fluxcd/flux2/pkg/manifestgen/install"
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
	testVersion = "test-version"
)

var _ = Describe("InstallFlux", func() {
	var fakeLogger *loggerfakes.FakeLogger
	var fakeContext context.Context
	var fakeInstallOptions install.Options

	BeforeEach(func() {
		fakeLogger = &loggerfakes.FakeLogger{}
		fakeContext = context.Background()
		fakeInstallOptions = install.MakeDefaultOptions()
	})

	It("should install flux successfully", func() {
		man := &mockResourceManagerForApply{}

		err := run.InstallFlux(fakeLogger, fakeContext, fakeInstallOptions, man)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return an apply all error if the resource manager returns an apply all error", func() {
		man := &mockResourceManagerForApply{state: stateApplyAllReturnErr}

		err := run.InstallFlux(fakeLogger, fakeContext, fakeInstallOptions, man)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(applyAllErrorMsg))
	})

	It("should return a wait for set error if the resource manager returns a wait for set error", func() {
		man := &mockResourceManagerForApply{state: stateWaitForSetReturnErr}

		err := run.InstallFlux(fakeLogger, fakeContext, fakeInstallOptions, man)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(waitForSetErrorMsg))
	})
})

var _ = Describe("GetFluxVersion", func() {
	var fakeLogger *loggerfakes.FakeLogger

	BeforeEach(func() {
		fakeLogger = &loggerfakes.FakeLogger{}
	})

	It("gets flux version", func() {
		kubeClientOpts := run.GetKubeClientOptions()

		contextName := "test-context"

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
