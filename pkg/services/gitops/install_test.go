package gitops_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	log "github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/services/gitops"
)

var (
	installParams gitops.InstallParams
	dir           string
)
var _ = Describe("Install", func() {
	BeforeEach(func() {
		fluxClient = &fluxfakes.FakeFlux{}
		kubeClient = &kubefakes.FakeKube{
			GetClusterStatusStub: func(c context.Context) kube.ClusterStatus {
				return kube.Unmodified
			},
		}
		gp := &gitprovidersfakes.FakeGitProvider{}
		fakeGit := &gitfakes.FakeGit{}
		// fakeGitClient := git.New(nil, fakeGit)
		fakeGit.WriteStub = func(path string, manifest []byte) error {
			return nil
		}

		dir, err := ioutil.TempDir("", "wego-install-test-")
		Expect(err).ShouldNot(HaveOccurred())

		gitClient := git.New(nil, wrapper.NewGoGit())
		ok, err := gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ok).Should(BeTrue())

		gitopsSrv = gitops.New(log.NewCLILogger(os.Stderr), fluxClient, kubeClient, gp, fakeGit)

		installParams = gitops.InstallParams{
			Namespace: "wego-system",
			DryRun:    false,
		}
	})
	var _ = AfterEach(func() {
		Expect(os.RemoveAll(dir)).To(Succeed())
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
	Context("when app url specified", func() {
		BeforeEach(func() {
			installParams.AppConfigURL = "ssh://127.0.0.1"
			fluxClient.InstallStub = func(s string, b bool) ([]byte, error) {
				return []byte("manifests"), nil
			}
		})
		It("calls flux install", func() {
			manifests, err := gitopsSrv.Install(installParams)
			Expect(err).ShouldNot(HaveOccurred())
			fmt.Printf("manifests returned are :%v\n", manifests)
			Expect(string(manifests)).To(ContainSubstring("manifests"))

			Expect(fluxClient.InstallCallCount()).To(Equal(2))

			namespace, dryRun := fluxClient.InstallArgsForCall(0)
			Expect(namespace).To(Equal("wego-system"))
			Expect(dryRun).To(Equal(true))
		})
	})

})
