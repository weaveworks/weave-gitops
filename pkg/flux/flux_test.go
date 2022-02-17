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

var _ = Describe("CreateSecretGit", func() {
	It("creates a git secret and returns the deploy key", func() {
		runner.RunStub = func(s1 string, s2 ...string) ([]byte, error) {
			return []byte(`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCh...`), nil
		}

		repoUrl, err := gitproviders.NewRepoURL("ssh://git@github.com/foo/bar.git")
		Expect(err).ShouldNot(HaveOccurred())
		out, err := fluxClient.CreateSecretGit("my-secret", repoUrl, wego.DefaultNamespace)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCh...")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("create secret git my-secret --url ssh://git@github.com/foo/bar.git --namespace %s --export", wego.DefaultNamespace)))
	})
})

func fluxPath() string {
	homeDir, err := os.UserHomeDir()
	Expect(err).ShouldNot(HaveOccurred())

	return filepath.Join(homeDir, ".wego", "bin", "flux-0.12.0")
}
