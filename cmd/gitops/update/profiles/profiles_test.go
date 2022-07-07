package profiles_test

import (
	"github.com/weaveworks/weave-gitops/cmd/gitops/root"

	"github.com/go-resty/resty/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Update Profile(s)", func() {
	var cmd *cobra.Command

	BeforeEach(func() {
		cmd = root.RootCmd(resty.New())
	})

	When("the flags are valid", func() {
		It("accepts all known flags for updating a profile", func() {
			cmd.SetArgs([]string{
				"update", "profile",
				"--name", "podinfo",
				"--version", "0.0.1",
				"--cluster", "prod",
				"--namespace", "test-namespace",
				"--config-repo", "https://ssh@github:test/test.git",
				"--auto-merge", "true",
				"--endpoint", "localhost:8080",
				"--skip-auth",
			})

			err := cmd.Execute()
			Expect(err.Error()).NotTo(ContainSubstring("unknown flag"))
		})
	})

	When("flags are not valid", func() {
		It("fails if --name, --cluster, --version or --config-repo are not provided", func() {
			cmd.SetArgs([]string{
				"update", "profile",
				"--endpoint", "localhost:8080",
			})

			err := cmd.Execute()
			Expect(err).To(MatchError("required flag(s) \"cluster\", \"config-repo\", \"name\", \"version\" not set"))
		})

		It("fails if given version is not valid semver", func() {
			cmd.SetArgs([]string{
				"update", "profile",
				"--name", "podinfo",
				"--config-repo", "ssh://git@github.com/owner/config-repo.git",
				"--cluster", "prod",
				"--version", "&%*/v",
				"--endpoint", "localhost:8080",
				"--skip-auth",
			})

			err := cmd.Execute()
			Expect(err).To(MatchError(ContainSubstring("error parsing --version=&%*/v")))
		})
	})

	When("a flag is unknown", func() {
		It("fails", func() {
			cmd.SetArgs([]string{
				"update", "profile",
				"--unknown", "param",
			})

			err := cmd.Execute()
			Expect(err).To(MatchError("unknown flag: --unknown"))
		})
	})
})
