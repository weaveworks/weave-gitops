package flux_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/runner/runnerfakes"
)

var (
	runner     *runnerfakes.FakeRunner
	fluxClient *flux.FluxClient
)

var _ = BeforeEach(func() {
	runner = &runnerfakes.FakeRunner{}

	fluxClient = flux.New(runner)
})

var _ = Describe("Install", func() {
	It("installs flux", func() {
		_, err := fluxClient.Install("wego-system", false)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(runner.RunWithOutputStreamCallCount()).To(Equal(1))

		cmd, args := runner.RunWithOutputStreamArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))
		Expect(strings.Join(args, " ")).To(Equal("install --namespace wego-system"))
	})

	It("exports the install manifests", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}

		out, err := fluxClient.Install("wego-system", true)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))
		Expect(strings.Join(args, " ")).To(Equal("install --namespace wego-system --export"))
	})
})

var _ = Describe("Uninstall", func() {
	It("uninstalls flux", func() {
		err := fluxClient.Uninstall("wego-system", false)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(runner.RunWithOutputStreamCallCount()).To(Equal(1))

		cmd, args := runner.RunWithOutputStreamArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))
		Expect(strings.Join(args, " ")).To(Equal("uninstall -s --namespace wego-system"))
	})

	It("add dry-run to the call", func() {
		err := fluxClient.Uninstall("wego-system", true)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(runner.RunWithOutputStreamCallCount()).To(Equal(1))

		cmd, args := runner.RunWithOutputStreamArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))
		Expect(strings.Join(args, " ")).To(Equal("uninstall -s --namespace wego-system --dry-run"))
	})
})

var _ = Describe("CreateSourceGit", func() {
	It("creates a source git", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}
		out, err := fluxClient.CreateSourceGit("my-name", "https://github.com/foo/my-name", "main", "my-secret", "wego-system")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal("create source git my-name --url https://github.com/foo/my-name --branch main --secret-ref my-secret --namespace wego-system --interval 30s --export"))
	})
})

var _ = Describe("CreateSourceHelm", func() {
	It("creates a source helm", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}
		out, err := fluxClient.CreateSourceHelm("my-name", "https://github.com/foo/my-name", "wego-system")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal("create source helm my-name --url https://github.com/foo/my-name --namespace wego-system --interval 30s --export"))
	})
})

var _ = Describe("CreateKustomization", func() {
	It("creates a kustomization", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}
		out, err := fluxClient.CreateKustomization("my-name", "my-source", "./path", "wego-system")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal("create kustomization my-name --path ./path --source my-source --namespace wego-system --prune true --validation client --interval 1m --export"))
	})
})

var _ = Describe("CreateHelmReleaseGitRepository", func() {
	It("creates a helm release with a git repository", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}
		out, err := fluxClient.CreateHelmReleaseGitRepository("my-name", "my-source", "./chart-path", "wego-system")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal("create helmrelease my-name --source GitRepository/my-source --chart ./chart-path --namespace wego-system --interval 5m --export"))
	})
})

var _ = Describe("CreateHelmReleaseHelmRepository", func() {
	It("creates a helm release with a helm repository", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}
		out, err := fluxClient.CreateHelmReleaseHelmRepository("my-name", "my-chart", "wego-system")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal("create helmrelease my-name --source HelmRepository/my-name --chart my-chart --namespace wego-system --interval 5m --export"))
	})
})

var _ = Describe("CreateSecretGit", func() {
	It("creates a git secret and returns the deploy key", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte(`✚ deploy key: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCh...`), nil
		}
		out, err := fluxClient.CreateSecretGit("my-secret", "ssh://git@github.com/foo/bar.git", "wego-system")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCh...")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		fmt.Println(strings.Join(args, " "))
		Expect(strings.Join(args, " ")).To(Equal("create secret git my-secret --url ssh://git@github.com/foo/bar.git --namespace wego-system"))
	})
})

func fluxPath() string {
	homeDir, err := os.UserHomeDir()
	Expect(err).ShouldNot(HaveOccurred())
	return filepath.Join(homeDir, ".wego", "bin", "flux-0.12.0")
}
