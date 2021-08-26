package gitops_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/gitops"
)

var uninstallParams gitops.UinstallParams

var _ = Describe("Uninstall", func() {
	BeforeEach(func() {
		fluxClient = &fluxfakes.FakeFlux{}
		kubeClient = &kubefakes.FakeKube{
			GetClusterStatusStub: func(ctx context.Context) kube.ClusterStatus {
				return kube.WeGOInstalled
			},
		}
		gitopsSrv = gitops.New(&loggerfakes.FakeLogger{}, fluxClient, kubeClient)

		uninstallParams = gitops.UinstallParams{
			Namespace: "wego-system",
			DryRun:    false,
		}
	})

	It("checks if wego is installed before proceeding", func() {
		err := gitopsSrv.Uninstall(uninstallParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(kubeClient.GetClusterStatusCallCount()).To(Equal(1))
		Expect(fluxClient.UninstallCallCount()).To(Equal(1))

		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.Unmodified
		}

		err = gitopsSrv.Uninstall(uninstallParams)
		Expect(err).Should(MatchError("Wego is not installed... exiting"))
	})

	It("calls flux uninstall", func() {
		err := gitopsSrv.Uninstall(uninstallParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(fluxClient.UninstallCallCount()).To(Equal(1))

		namespace, dryRun := fluxClient.UninstallArgsForCall(0)
		Expect(namespace).To(Equal("wego-system"))
		Expect(dryRun).To(Equal(false))
	})

	It("deletes app crd", func() {
		err := gitopsSrv.Uninstall(uninstallParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(kubeClient.DeleteCallCount()).To(Equal(1))

		appCRD, namespace := kubeClient.DeleteArgsForCall(0)
		Expect(appCRD).To(ContainSubstring("kind: App"))
		Expect(namespace).To(Equal("wego-system"))
	})

	Context("when dry-run", func() {
		BeforeEach(func() {
			uninstallParams.DryRun = true
		})

		It("calls flux uninstall", func() {
			err := gitopsSrv.Uninstall(uninstallParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(fluxClient.UninstallCallCount()).To(Equal(1))

			namespace, dryRun := fluxClient.UninstallArgsForCall(0)
			Expect(namespace).To(Equal("wego-system"))
			Expect(dryRun).To(Equal(true))
		})

		It("does not call kube apply", func() {
			err := gitopsSrv.Uninstall(uninstallParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(kubeClient.DeleteCallCount()).To(Equal(0))
		})
	})
})
