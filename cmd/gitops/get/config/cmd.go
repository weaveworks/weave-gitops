package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"

	gitopsConfig "github.com/weaveworks/weave-gitops/pkg/config"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

func ConfigCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Prints out the CLI configuration for Weave GitOps",
		Example: `
# Prints out the CLI configuration for Weave GitOps
gitops get config`,
		SilenceUsage:      true,
		SilenceErrors:     true,
		RunE:              getConfigCommandRunE(opts),
		DisableAutoGenTag: true,
	}

	return cmd
}

func getConfigCommandRunE(opts *config.Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		log := logger.NewCLILogger(os.Stdout)

		cfg, err := gitopsConfig.GetConfig(log, false)
		if err != nil {
			return err
		}

		log.Successf("Your CLI configuration for Weave GitOps:")

		cfgStr, err := cfg.String()
		if err != nil {
			log.Failuref("Error printing config")
			return err
		}

		fmt.Println(cfgStr)

		return nil
	}
}
