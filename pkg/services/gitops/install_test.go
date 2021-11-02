package gitops_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/services/gitops"
)

var (
	installParams     gitops.InstallParams
	dir               string
	manifestsByPath   map[string][]byte
	fakeFluxManifests = []byte("fluxManifests")
)
var _ = Describe("Install", func() {

	BeforeEach(func() {
		fluxClient = &fluxfakes.FakeFlux{}
		kubeClient = &kubefakes.FakeKube{
			GetClusterStatusStub: func(c context.Context) kube.ClusterStatus {
				return kube.Unmodified
			},
		}
		fakeProvider = &gitprovidersfakes.FakeGitProvider{}
		fakeGit = &gitfakes.FakeGit{}
		fakeGit.WriteReturns(nil)

		dir, err := ioutil.TempDir("", "wego-install-test-")
		Expect(err).ShouldNot(HaveOccurred())

		gitClient := git.New(nil, wrapper.NewGoGit())
		Expect(gitClient.Init(dir, "https://github.com/github/gitignore", "master")).To(BeTrue())

		gitopsSrv = gitops.New(&loggerfakes.FakeLogger{}, fluxClient, kubeClient)

		installParams = gitops.InstallParams{
			Namespace: wego.DefaultNamespace,
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
		_, err := gitopsSrv.Install(fakeGit, fakeProvider, installParams)
		Expect(err).Should(MatchError("Weave GitOps does not yet support installation onto a cluster that is using Flux.\nPlease uninstall flux before proceeding:\n  $ flux uninstall"))

		kubeClient.GetClusterStatusStub = func(c context.Context) kube.ClusterStatus {
			return kube.Unknown
		}
		_, err = gitopsSrv.Install(fakeGit, fakeProvider, installParams)
		Expect(err).Should(MatchError("Weave GitOps cannot talk to the cluster"))
	})

	It("calls flux install", func() {
		_, err := gitopsSrv.Install(fakeGit, fakeProvider, installParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(fluxClient.InstallCallCount()).To(Equal(1))

		namespace, dryRun := fluxClient.InstallArgsForCall(0)
		Expect(namespace).To(Equal(wego.DefaultNamespace))
		Expect(dryRun).To(Equal(false))
	})

	It("applies app crd and wego-app manifests", func() {
		_, err := gitopsSrv.Install(fakeGit, fakeProvider, installParams)
		Expect(err).ShouldNot(HaveOccurred())

		_, appCRD, namespace := kubeClient.ApplyArgsForCall(0)
		Expect(appCRD).To(ContainSubstring("kind: App"))
		Expect(namespace).To(Equal(wego.DefaultNamespace))

		_, serviceAccount, namespace := kubeClient.ApplyArgsForCall(1)
		Expect(serviceAccount).To(ContainSubstring("kind: ServiceAccount"))
		Expect(namespace).To(Equal(wego.DefaultNamespace))

		_, roleBinding, namespace := kubeClient.ApplyArgsForCall(2)
		Expect(roleBinding).To(ContainSubstring("kind: RoleBinding"))
		Expect(namespace).To(Equal(wego.DefaultNamespace))

		_, role, namespace := kubeClient.ApplyArgsForCall(3)
		Expect(role).To(ContainSubstring("kind: Role"))
		Expect(namespace).To(Equal(wego.DefaultNamespace))

		_, service, namespace := kubeClient.ApplyArgsForCall(4)
		Expect(service).To(ContainSubstring("kind: Service"))
		Expect(namespace).To(Equal(wego.DefaultNamespace))

		_, deployment, namespace := kubeClient.ApplyArgsForCall(5)
		Expect(deployment).To(ContainSubstring("kind: Deployment"))
		Expect(namespace).To(Equal(wego.DefaultNamespace))

	})

	Context("when dry-run", func() {
		BeforeEach(func() {
			installParams.DryRun = true
			fluxClient.InstallReturns(fakeFluxManifests, nil)
		})

		It("calls flux install", func() {
			manifests, err := gitopsSrv.Install(fakeGit, fakeProvider, installParams)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(string(manifests)).To(ContainSubstring(string(fakeFluxManifests)))

			Expect(fluxClient.InstallCallCount()).To(Equal(1))

			namespace, dryRun := fluxClient.InstallArgsForCall(0)
			Expect(namespace).To(Equal(wego.DefaultNamespace))
			Expect(dryRun).To(Equal(true))
		})

		It("appends app crd to flux install output", func() {
			manifests, err := gitopsSrv.Install(fakeGit, fakeProvider, installParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(string(manifests)).To(ContainSubstring("kind: App"))
		})

		It("has flux manifests", func() {
			tests := []string{
				"GitRepository",
				"HelmRelease",
				"HelmRepository",
				"Kustomization",
			}

			fluxClient.InstallStub = func(s string, b bool) ([]byte, error) {
				var f string
				for _, k := range tests {
					f += fmt.Sprintf("kind: %s\n", k)
				}

				return []byte(f), nil
			}

			manifests, err := gitopsSrv.Install(fakeGit, fakeProvider, installParams)
			Expect(err).ShouldNot(HaveOccurred())

			s := string(manifests)
			for _, k := range tests {
				Expect(s).To(ContainSubstring("kind: "+k), "Missing CRD for: "+k)
			}
		})

		It("does not call kube apply", func() {
			_, err := gitopsSrv.Install(fakeGit, fakeProvider, installParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(kubeClient.ApplyCallCount()).To(Equal(0))
		})
	})
	Context("when app url specified", func() {
		BeforeEach(func() {
			installParams.AppConfigURL = "ssh://git@github.com/foo/somevalidrepo.git"
			fluxClient.InstallReturns(fakeFluxManifests, nil)
			manifestsByPath = map[string][]byte{}

			fakeGit.WriteStub = func(path string, manifest []byte) error {
				manifestsByPath[path] = manifest
				return nil
			}
		})
		It("calls flux install", func() {
			manifests, err := gitopsSrv.Install(fakeGit, fakeProvider, installParams)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(string(manifests)).To(ContainSubstring(string(fakeFluxManifests)))

			Expect(fluxClient.InstallCallCount()).To(Equal(2))

			namespace, dryRun := fluxClient.InstallArgsForCall(0)
			Expect(namespace).To(Equal("wego-system"))
			Expect(dryRun).To(Equal(true))
		})
		It("flux kustomization file for user and system have hidden directory", func() {
			_, err := gitopsSrv.Install(fakeGit, fakeProvider, installParams)
			Expect(err).ShouldNot(HaveOccurred())
			for i := fluxClient.CreateKustomizationCallCount() - 1; i >= 0; i-- {
				_, _, path, _ := fluxClient.CreateKustomizationArgsForCall(i)
				Expect(path).To(HavePrefix("..weave-gitops"))
			}
		})
		It("flux gitops toolkit file is stored", func() {
			m, err := gitopsSrv.Install(fakeGit, fakeProvider, installParams)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(fakeGit.WriteCallCount()).Should(Equal(7))
			tested := false
			for k, v := range manifestsByPath {
				if strings.Contains(k, "gitops-runtime.yaml") {
					Expect(len(v)).ToNot(BeZero())
					Expect(string(m)).To(ContainSubstring(string(v)))
					tested = true
				}
			}
			Expect(tested).To(BeTrue())
		})

	})
	Context("when app url specified && dry-run", func() {
		BeforeEach(func() {
			installParams.AppConfigURL = "ssh://git@github.com/foo/somevalidrepo.git"
			installParams.DryRun = true
			fluxClient.InstallReturns(fakeFluxManifests, nil)
		})
		It("skips flux install", func() {
			_, err := gitopsSrv.Install(fakeGit, fakeProvider, installParams)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(kubeClient.ApplyCallCount()).Should(Equal(0), "With dry-run and app-config-url nothing should be sent to k8s")
		})
		It("writes no manifests to the repo", func() {
			_, err := gitopsSrv.Install(fakeGit, fakeProvider, installParams)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(fakeGit.WriteCallCount()).Should(Equal(0), "With dry-run and app-config-url nothing should be written to git")
		})
		It("flux manifests are returned", func() {
			Expect(gitopsSrv.Install(fakeGit, fakeProvider, installParams)).To(ContainSubstring(string(fakeFluxManifests)))
			Expect(fakeGit.WriteCallCount()).Should(Equal(0), "With dry-run and app-config-url nothing should be written to git")
		})
	})

})
