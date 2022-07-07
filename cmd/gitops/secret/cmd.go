package secret

import (
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/internal/config"
	"golang.org/x/crypto/bcrypt"
)

type secretFlags struct{}

var flags secretFlags

func SecretCommand(opts *config.Options, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "makes a secret",
		Long:  "makes a secret long style",
		Example: `
# gitops secret <password>
`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE:       SecretCommandPreRunE(),
		RunE:          SecretCommandRunE(opts, client),
	}

	return cmd
}

func SecretCommandPreRunE() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}

func SecretCommandRunE(opts *config.Options, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("password required: gitops secret <password>")
		}

		if len(args) > 1 {
			return errors.New("password must be one continuous string")
		}

		fmt.Println(bcrypt.GenerateFromPassword([]byte(args[0]), 10))

		return nil
	}
}
