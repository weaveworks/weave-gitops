package breakglass

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/tf-controller/tfctl"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

func Command(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "break-glass",
		Aliases: []string{"break-the-glass", "bg", "btg"},
		Short:   "Break the glass",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := tfctl.New("", "")

			return app.BreakTheGlass(os.Stdout, args[0])
		},
	}
	return cmd
}
