package analytics

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("getCommandPath", func() {
	It("gets command path from command", func() {
		var rootCmd = &cobra.Command{
			Use: "root",
		}

		var parentCmd = &cobra.Command{
			Use: "parent",
		}

		var cmd = &cobra.Command{
			Use: "cmd",
		}

		parentCmd.AddCommand(cmd)

		rootCmd.AddCommand(parentCmd)

		commandPath := getCommandPath(cmd)

		Expect(commandPath).To(Equal("parent cmd"))

	})
})
