package analytics

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("getCommandPath", func() {
	It("gets command path from command", func() {
		rootCmd := &cobra.Command{
			Use: "root",
		}

		parentCmd := &cobra.Command{
			Use: "parent",
		}

		cmd := &cobra.Command{
			Use: "cmd",
		}

		parentCmd.AddCommand(cmd)

		rootCmd.AddCommand(parentCmd)

		commandPath := getCommandPath(cmd)

		Expect(commandPath).To(Equal("parent cmd"))
	})
})

var _ = Describe("sanitizeFlagName", func() {
	It("sanitizes flag names", func() {
		Expect(sanitizeFlagName("timeout")).To(Equal("timeout"))
		Expect(sanitizeFlagName("port-forward")).To(Equal("port_forward"))
		Expect(sanitizeFlagName("allow-k8s-context")).To(Equal("allow_k8s_context"))
		Expect(sanitizeFlagName("very-long-flag-very-long-flag-very-long-flag")).To(Equal("very_long_flag_very_long_flag_ve"))
		Expect(sanitizeFlagName("short-flag")).To(Equal("short_flag"))
	})
})
