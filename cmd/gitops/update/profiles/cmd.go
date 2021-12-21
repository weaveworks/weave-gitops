package profiles

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/profiles"
	"github.com/weaveworks/weave-gitops/pkg/server"
)

var (
	port        string
	clusterName string
	profileName string
	version     string
)

var Cmd = &cobra.Command{
	Use:           "profile",
	Aliases:       []string{"profiles"},
	Short:         "Update a profile",
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	Example: `
# Update profile
gitops update profile --profile <name> --cluster <cluster-name>
`,
	RunE: runCmd,
}

func init() {
	Cmd.Flags().StringVar(&port, "port", server.DefaultPort, "port the profiles API is running on")
	Cmd.Flags().StringVar(&clusterName, "cluster", "", "cluster that contains the profile to update")
	Cmd.Flags().StringVar(&profileName, "profile", "", "profile name to update")
	Cmd.Flags().StringVar(&version, "version", "", "profile name to update")
}

func runCmd(cmd *cobra.Command, args []string) error {
	return profiles.UpdateProfile(context.Background(), profiles.UpdateOptions{})
}
