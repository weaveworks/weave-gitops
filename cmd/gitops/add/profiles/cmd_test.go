package profiles_test

import (
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/root"
)

var _ = Describe("Add Profiles", func() {
	var (
		cmd *cobra.Command
	)

	BeforeEach(func() {
		client := resty.New()
		httpmock.ActivateNonDefault(client.GetClient())
		defer httpmock.DeactivateAndReset()
		cmd = root.RootCmd(client)
	})

	When("the flags are valid", func() {
		It("accepts all known flags for adding a profile", func() {
			cmd.SetArgs([]string{
				"add", "profile",
				"--name", "podinfo",
				"--version", "0.0.1",
				"--cluster", "prod",
				"--config-repo", "https://ssh@github:test/test.git",
				"--auto-merge", "true",
			})

			err := cmd.Execute()
			Expect(err.Error()).NotTo(ContainSubstring("unknown flag"))
		})
	})

	When("a flag is unknown", func() {
		It("fails", func() {
			cmd.SetArgs([]string{
				"add", "profile",
				"--unknown", "param",
			})

			err := cmd.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("unknown flag: --unknown"))
		})
	})
})
