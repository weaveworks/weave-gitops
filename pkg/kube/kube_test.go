package kube_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/runner/runnerfakes"
)

var (
	runner     *runnerfakes.FakeRunner
	kubeClient *kube.KubeClient
)

var _ = BeforeEach(func() {
	runner = &runnerfakes.FakeRunner{}

	kubeClient = kube.New(runner)
})

var _ = Describe("Apply", func() {
	It("applies manifests", func() {
		runner.RunWithStdinStub = func(s1 string, s2 []string, b []byte) ([]byte, error) {
			return []byte("out"), nil
		}

		out, err := kubeClient.Apply([]byte("manifests"), "wego-system")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		cmd, args, manifests := runner.RunWithStdinArgsForCall(0)
		Expect(cmd).To(Equal("kubectl"))

		Expect(strings.Join(args, " ")).To(Equal("apply --namespace wego-system -f -"))
		Expect(manifests).To(Equal([]byte("manifests")))
	})
})

var _ = Describe("GetClusterStatus", func() {
	It("returns wego is installed", func() {
		runner.RunStub = func(cmd string, args ...string) ([]byte, error) {
			if strings.Join(args, " ") == "get crd apps.wego.weave.works" {
				return []byte("out"), nil
			}

			return []byte("error"), fmt.Errorf("error")
		}

		status := kubeClient.GetClusterStatus()
		Expect(status).To(Equal(kube.WeGOInstalled))
	})

	It("returns flux is installed", func() {
		runner.RunStub = func(cmd string, args ...string) ([]byte, error) {
			if strings.Join(args, " ") == "get namespace flux-system" {
				return []byte("out"), nil
			}

			return []byte("error"), fmt.Errorf("error")
		}

		status := kubeClient.GetClusterStatus()
		Expect(status).To(Equal(kube.FluxInstalled))
	})

	It("returns unmodified cluster", func() {
		runner.RunStub = func(cmd string, args ...string) ([]byte, error) {
			if strings.Join(args, " ") == "deployment coredns -n kube-system" {
				return []byte("out"), nil
			}

			return []byte("error"), fmt.Errorf("error")
		}

		status := kubeClient.GetClusterStatus()
		Expect(status).To(Equal(kube.Unmodified))
	})

	It("returns unknown when it cant talk to the cluster", func() {
		runner.RunStub = func(cmd string, args ...string) ([]byte, error) {
			return []byte("error"), fmt.Errorf("error")
		}

		status := kubeClient.GetClusterStatus()
		Expect(status).To(Equal(kube.Unknown))
	})
})

var _ = Describe("GetClusterName", func() {
	It("returns cluster name", func() {
		runner.RunStub = func(cmd string, args ...string) ([]byte, error) {
			return []byte("cluster-name\n"), nil
		}

		out, err := kubeClient.GetClusterName()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(string(out)).To(Equal("cluster-name"))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal("kubectl"))

		Expect(strings.Join(args, " ")).To(Equal("config current-context"))
	})
})
