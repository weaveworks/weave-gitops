package upgrade

import (
	"github.com/spf13/cobra"
)

var example = `  # Upgrade Weave GitOps
  gitops upgrade --version 0.0.17 --config-repo https://github.com/my-org/my-management-cluster.git

  # Upgrade Weave GitOps and set the natsURL
  gitops upgrade --version 0.0.17 --set "agentTemplate.natsURL=my-cluster.acme.org:4222" \
    --config-repo https://github.com/my-org/my-management-cluster.git`

var Cmd = &cobra.Command{
	Use:           "upgrade",
	Short:         "Upgrade to Weave GitOps Enterprise",
	Example:       example,
	RunE:          upgradeCmdRunE(),
	SilenceErrors: true,
	SilenceUsage:  true,
}

func upgradeCmdRunE() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		println(`To install enterprise, run:
    flux create whatever-stuff`)
		return nil
	}
}
