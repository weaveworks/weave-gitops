package gitops_test

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	log "github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/services/gitops"
)

var installParams gitops.InstallParams

var _ = Describe("Install", func() {
	BeforeEach(func() {
		fluxClient = &fluxfakes.FakeFlux{}
		kubeClient = &kubefakes.FakeKube{
			GetClusterStatusStub: func(c context.Context) kube.ClusterStatus {
				return kube.Unmodified
			},
		}
		gitopsSrv = gitops.New(log.NewCLILogger(os.Stderr), fluxClient, kubeClient)

		installParams = gitops.InstallParams{
			Namespace: "wego-system",
			DryRun:    false,
		}
	})

	It("checks cluster status", func() {
		kubeClient.GetClusterStatusStub = func(c context.Context) kube.ClusterStatus {
			return kube.FluxInstalled
		}
		_, err := gitopsSrv.Install(installParams)
		Expect(err).Should(MatchError("Weave GitOps does not yet support installation onto a cluster that is using Flux.\nPlease uninstall flux before proceeding:\n  $ flux uninstall"))

		kubeClient.GetClusterStatusStub = func(c context.Context) kube.ClusterStatus {
			return kube.Unknown
		}
		_, err = gitopsSrv.Install(installParams)
		Expect(err).Should(MatchError("Weave GitOps cannot talk to the cluster"))
	})

	It("calls flux install", func() {
		_, err := gitopsSrv.Install(installParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(fluxClient.InstallCallCount()).To(Equal(1))

		namespace, dryRun := fluxClient.InstallArgsForCall(0)
		Expect(namespace).To(Equal("wego-system"))
		Expect(dryRun).To(Equal(false))
	})

	It("applies app crd and wego-app manifests", func() {
		_, err := gitopsSrv.Install(installParams)
		Expect(err).ShouldNot(HaveOccurred())

		_, appCRD, namespace := kubeClient.ApplyArgsForCall(0)
		Expect(appCRD).To(ContainSubstring("kind: App"))
		Expect(namespace).To(Equal("wego-system"))

		_, serviceAccount, namespace := kubeClient.ApplyArgsForCall(1)
		Expect(serviceAccount).To(ContainSubstring("kind: ServiceAccount"))
		Expect(namespace).To(Equal("wego-system"))

		_, roleBinding, namespace := kubeClient.ApplyArgsForCall(2)
		Expect(roleBinding).To(ContainSubstring("kind: RoleBinding"))
		Expect(namespace).To(Equal("wego-system"))

		_, role, namespace := kubeClient.ApplyArgsForCall(3)
		Expect(role).To(ContainSubstring("kind: Role"))
		Expect(namespace).To(Equal("wego-system"))

		_, service, namespace := kubeClient.ApplyArgsForCall(4)
		Expect(service).To(ContainSubstring("kind: Service"))
		Expect(namespace).To(Equal("wego-system"))

		_, deployment, namespace := kubeClient.ApplyArgsForCall(5)
		Expect(deployment).To(ContainSubstring("kind: Deployment"))
		Expect(namespace).To(Equal("wego-system"))

	})

	Context("when dry-run", func() {
		BeforeEach(func() {
			installParams.DryRun = true
			fluxClient.InstallStub = func(s string, b bool) ([]byte, error) {
				return []byte("manifests"), nil
			}
		})

		It("calls flux install", func() {
			manifests, err := gitopsSrv.Install(installParams)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(string(manifests)).To(ContainSubstring("manifests"))

			Expect(fluxClient.InstallCallCount()).To(Equal(1))

			namespace, dryRun := fluxClient.InstallArgsForCall(0)
			Expect(namespace).To(Equal("wego-system"))
			Expect(dryRun).To(Equal(true))
		})

		It("appends app crd to flux install output", func() {
			manifests, err := gitopsSrv.Install(installParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(string(manifests)).To(ContainSubstring("kind: App"))
		})

		It("does not call kube apply", func() {
			_, err := gitopsSrv.Install(installParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(kubeClient.ApplyCallCount()).To(Equal(0))
		})
	})
})
