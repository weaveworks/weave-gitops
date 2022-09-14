package flux_test

import (
	"fmt"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

		repoURL, err := gitproviders.NewRepoURL("ssh://git@github.com/foo/bar.git")
		Expect(err).ShouldNot(HaveOccurred())
		out, err := fluxClient.CreateSecretGit("my-secret", repoURL, "flux-system")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(Equal([]byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCh...")))

		Expect(runner.RunCallCount()).To(Equal(1))

		cmd, args := runner.RunArgsForCall(0)
		Expect(cmd).To(Equal(fluxPath()))

		Expect(strings.Join(args, " ")).To(Equal(fmt.Sprintf("create secret git my-secret --url ssh://git@github.com/foo/bar.git --namespace %s --export", "flux-system")))
	})
})

func fluxPath() string {
	return filepath.Join("flux")
}
