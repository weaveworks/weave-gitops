package bcrypt

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func HashCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bcrypt-hash",
		Short: "Generates a hashed secret",
		Example: `
# PASSWORD="<your password>"
# echo $PASSWORD | gitops get bcrypt-hash
`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          hashCommandRunE(),
	}

	return cmd
}

func hashCommandRunE() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		file := os.Stdin
		stats, err := file.Stat()

		if err != nil {
			return err
		}

		var p []byte

		if stats.Size() == 0 {
			fmt.Print("error: no password found\nEnter Password: ")

			p, err = term.ReadPassword(int(os.Stdin.Fd()))

			if err != nil {
				return nil
			}
		} else {
			p, err = io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
		}

		secret, err := bcrypt.GenerateFromPassword(p, bcrypt.DefaultCost)

		if err != nil {
			return err
		}

		fmt.Println(string(secret))

		return nil
	}
}
