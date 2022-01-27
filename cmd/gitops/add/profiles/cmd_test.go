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
				"--namespace", "test-namespace",
				"--config-repo", "https://ssh@github:test/test.git",
				"--auto-merge", "true",
			})

			err := cmd.Execute()
			Expect(err.Error()).NotTo(ContainSubstring("unknown flag"))
		})
	})

	When("flags are not valid", func() {
		It("fails if --name is not provided", func() {
			cmd.SetArgs([]string{
				"add", "profile",
			})

			err := cmd.Execute()
			Expect(err).To(MatchError("--name should be provided"))
		})

		When("--name is specified", func() {
			It("fails if --name value is <= 63 characters in length", func() {
				cmd.SetArgs([]string{
					"add", "profile",
					"--name", "a234567890123456789012345678901234567890123456789012345678901234",
				})
				err := cmd.Execute()
				Expect(err).To(MatchError("--name value is too long: a234567890123456789012345678901234567890123456789012345678901234; must be <= 63 characters"))
			})

			It("fails if --name is prefixed by 'wego'", func() {
				cmd.SetArgs([]string{
					"add", "profile",
					"--name", "wego-app",
				})
				err := cmd.Execute()
				Expect(err).To(MatchError("the prefix 'wego' is used by weave gitops and is not allowed for a profile name"))
			})
		})

		It("fails if --config-repo is not provided", func() {
			cmd.SetArgs([]string{
				"add", "profile",
				"--name", "podinfo",
			})

			err := cmd.Execute()
			Expect(err).To(MatchError("--config-repo should be provided"))
		})

		It("fails if --config-repo is not provided", func() {
			cmd.SetArgs([]string{
				"add", "profile",
				"--name", "podinfo",
				"--config-repo", "ssh://git@github.com/owner/config-repo.git",
			})

			err := cmd.Execute()
			Expect(err).To(MatchError("--cluster should be provided"))
		})

		It("fails if given version is not valid semver", func() {
			cmd.SetArgs([]string{
				"add", "profile",
				"--name", "podinfo",
				"--config-repo", "ssh://git@github.com/owner/config-repo.git",
				"--cluster", "prod",
				"--version", "&%*/v",
			})

			err := cmd.Execute()
			Expect(err).To(MatchError("error parsing --version=&%*/v: Invalid Semantic Version"))
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
