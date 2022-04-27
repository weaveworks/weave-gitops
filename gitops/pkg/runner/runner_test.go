package runner_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/gitops/pkg/runner"
)

var (
	cliRunner runner.Runner
)

var _ = BeforeEach(func() {
	cliRunner = &runner.CLIRunner{}
})

var _ = Describe("Run", func() {
	It("runs a command and returns the output", func() {
		out, err := cliRunner.Run("echo", "foo")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(out)).To(Equal("foo\n"))
	})

	It("returns the error output when it fails", func() {
		_, err := cliRunner.Run("fooo", "bla")
		Expect(errors.Unwrap(err)).To(MatchError("executable file not found in $PATH"))
	})
})
