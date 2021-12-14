package flux_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner/runnerfakes"
)

var (
	runner     *runnerfakes.FakeRunner
	fluxClient *flux.FluxClient
)

var _ = BeforeEach(func() {
	runner = &runnerfakes.FakeRunner{}
	fluxClient = flux.New(osys.New(), runner)
})

var _ = Describe("Install", func() {
	It("installs flux", func() {
		_, err := fluxClient.Install(wego.DefaultNamespace, false)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(runner.RunWithOutputStreamCallCount()).To(Equal(1))

		cmd, args := runner.RunWithOutputStreamArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))
		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("install --namespace %s --components-extra image-reflector-controller,image-automation-controller", wego.DefaultNamespace)))
	})

	It("exports the install manifests", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}

		out, err := fluxClient.Install(wego.DefaultNamespace, true)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))
		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("install --namespace %s --components-extra image-reflector-controller,image-automation-controller --export", wego.DefaultNamespace)))
	})
})

var _ = Describe("Uninstall", func() {
	It("uninstalls flux", func() {
		err := fluxClient.Uninstall(wego.DefaultNamespace, false)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(runner.RunWithOutputStreamCallCount()).To(Equal(1))

		cmd, args := runner.RunWithOutputStreamArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))
		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("uninstall -s --namespace %s", wego.DefaultNamespace)))
	})

	It("add dry-run to the call", func() {
		err := fluxClient.Uninstall(wego.DefaultNamespace, true)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(runner.RunWithOutputStreamCallCount()).To(Equal(1))

		cmd, args := runner.RunWithOutputStreamArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))
		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("uninstall -s --namespace %s --dry-run", wego.DefaultNamespace)))
	})
})

var _ = Describe("CreateSourceGit", func() {
	It("creates a git source", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}

		repoUrl, err := gitproviders.NewRepoURL("https://github.com/foo/my-name")
		Expect(err).ShouldNot(HaveOccurred())
		out, err := fluxClient.CreateSourceGit("my-name", repoUrl, "main", "my-secret", wego.DefaultNamespace, nil)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))
		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("create source git my-name --branch main --namespace %s --interval 30s --export --secret-ref my-secret --url https://github.com/foo/my-name.git", wego.DefaultNamespace)))
	})

	It("creates a git source for a public repo", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}
		repoUrl, err := gitproviders.NewRepoURL("ssh://git@github.com/foo/my-name")
		Expect(err).ShouldNot(HaveOccurred())
		out, err := fluxClient.CreateSourceGit("my-name", repoUrl, "main", "", wego.DefaultNamespace, nil)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("create source git my-name --branch main --namespace %s --interval 30s --export --url https://github.com/foo/my-name.git", wego.DefaultNamespace)))
	})

	It("creates a git source for a public gitlab repo", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}

		repoUrl, err := gitproviders.NewRepoURL("ssh://git@gitlab.com/foo/my-name")
		Expect(err).ShouldNot(HaveOccurred())
		out, err := fluxClient.CreateSourceGit("my-name", repoUrl, "main", "", wego.DefaultNamespace, nil)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("create source git my-name --branch main --namespace %s --interval 30s --export --url https://gitlab.com/foo/my-name.git", wego.DefaultNamespace)))
	})

	It("passes http credentials to flux", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}

		repoUrl, err := gitproviders.NewRepoURL("https://gitlab.com/foo/my-name")
		Expect(err).ShouldNot(HaveOccurred())
		out, err := fluxClient.CreateSourceGit("my-name", repoUrl, "main", "", wego.DefaultNamespace, &flux.HTTPSCreds{Username: "test", Password: "password"})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("create source git my-name --branch main --namespace %s --interval 30s --export --url https://gitlab.com/foo/my-name.git --username test --password password", wego.DefaultNamespace)))
	})
})

var _ = Describe("CreateSourceHelm", func() {
	It("creates a source helm", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}
		out, err := fluxClient.CreateSourceHelm("my-name", "https://github.com/foo/my-name", wego.DefaultNamespace)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("create source helm my-name --url https://github.com/foo/my-name --namespace %s --interval 30s --export", wego.DefaultNamespace)))
	})
})

var _ = Describe("CreateKustomization", func() {
	It("creates a kustomization", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}
		out, err := fluxClient.CreateKustomization("my-name", "my-source", "./path", wego.DefaultNamespace)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("create kustomization my-name --path ./path --source my-source --namespace %s --prune true --interval 1m --export", wego.DefaultNamespace)))
	})
})

var _ = Describe("CreateHelmReleaseGitRepository", func() {
	It("creates a helm release with a git repository", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}
		out, err := fluxClient.CreateHelmReleaseGitRepository("my-name", "my-source", "./chart-path", wego.DefaultNamespace, "")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("create helmrelease my-name --source GitRepository/my-source --chart ./chart-path --namespace %s --interval 5m --export", wego.DefaultNamespace)))
	})

	It("creates a helm release with a git repository and a target namespace", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}
		out, err := fluxClient.CreateHelmReleaseGitRepository("my-name", "my-source", "./chart-path", wego.DefaultNamespace, "sock-shop")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("create helmrelease my-name --source GitRepository/my-source --chart ./chart-path --namespace %s --interval 5m --export --target-namespace sock-shop", wego.DefaultNamespace)))
	})
})

var _ = Describe("CreateHelmReleaseHelmRepository", func() {
	It("creates a helm release with a helm repository", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}
		out, err := fluxClient.CreateHelmReleaseHelmRepository("my-name", "my-chart", wego.DefaultNamespace, "")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("create helmrelease my-name --source HelmRepository/my-name --chart my-chart --namespace %s --interval 5m --export", wego.DefaultNamespace)))
	})

	It("creates a helm release with a helm repository and a target namespace", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte("out"), nil
		}
		out, err := fluxClient.CreateHelmReleaseHelmRepository("my-name", "my-chart", wego.DefaultNamespace, "sock-shop")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("out")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("create helmrelease my-name --source HelmRepository/my-name --chart my-chart --namespace %s --interval 5m --export --target-namespace sock-shop", wego.DefaultNamespace)))
	})
})

var _ = Describe("CreateSecretGit", func() {
	It("creates a git secret and returns the deploy key", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte(`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCh...`), nil
		}

		repoUrl, err := gitproviders.NewRepoURL("ssh://git@github.com/foo/bar.git")
		Expect(err).ShouldNot(HaveOccurred())
		out, err := fluxClient.CreateSecretGit("my-secret", repoUrl, wego.DefaultNamespace, nil)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCh...")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("create secret git my-secret --url ssh://git@github.com/foo/bar.git --namespace %s --export", wego.DefaultNamespace)))
	})

	It("passes http credentials through to flux", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte(`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCh...`), nil
		}

		repoUrl, err := gitproviders.NewRepoURL("ssh://git@github.com/foo/bar.git")
		Expect(err).ShouldNot(HaveOccurred())
		out, err := fluxClient.CreateSecretGit("my-secret", repoUrl, wego.DefaultNamespace, &flux.HTTPSCreds{Username: "test", Password: "password"})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCh...")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("create secret git my-secret --url ssh://git@github.com/foo/bar.git --namespace %s --export --username test --password password", wego.DefaultNamespace)))
	})

})

func fluxPath() string {
	homeDir, err := os.UserHomeDir()
	Expect(err).ShouldNot(HaveOccurred())

	return filepath.Join(homeDir, ".wego", "bin", "flux-0.12.0")
}
