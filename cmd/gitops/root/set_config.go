package root

import (
	"fmt"

	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	gitopsConfig "github.com/weaveworks/weave-gitops/pkg/config"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

const (
	analyticsKey = "analytics"
)

var analyticsValue bool

func setConfigCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Set the CLI configuration for Weave GitOps",
		Example: `
# Enables analytics in the current user's CLI configuration for Weave GitOps
gitops set config analytics true`,
		SilenceUsage:      true,
		SilenceErrors:     true,
		PreRunE:           setConfigCommandPreRunE(&opts.Endpoint),
		RunE:              setConfigCommandRunE(opts),
		DisableAutoGenTag: true,
	}

	return cmd
}

func setConfigCommandPreRunE(endpoint *string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("at least two positional arguments required")
		}

		if args[0] != analyticsKey {
			return cmderrors.ErrInvalidArgs
		}

		var err error

		analyticsValue, err = strconv.ParseBool(strings.TrimSpace(args[1]))
		if err != nil {
			return cmderrors.ErrInvalidArgs
		}

		return nil
	}
}

func setConfigCommandRunE(opts *config.Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		log := logger.NewCLILogger(os.Stdout)

		if !analyticsValue {
			log.Warningf("This will only turn off analytics for the GitOps CLI. Please refer to the documentation to turn off the analytics in the GitOps Dashboard.")
		}

		cfg, err := gitopsConfig.GetConfig(log, true)
		if err != nil {
			return err
		}

		cfg.Analytics = analyticsValue

		if cfg.UserID == "" {
			seed := time.Now().UnixNano()

			cfg.UserID = gitopsConfig.GenerateUserID(10, seed)
		}

		if err = gitopsConfig.SaveConfig(log, cfg); err != nil {
			log.Failuref("Error saving GitOps CLI config")
			return err
		}

		return nil
	}
}
