package check

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"

	"github.com/weaveworks/weave-gitops/cmd/gitops/check/oidcconfig"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/run"
	"github.com/weaveworks/weave-gitops/pkg/services/check"
)

func GetCommand(opts *config.Options) *cobra.Command {
	var kubeConfigArgs *genericclioptions.ConfigFlags

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Validates flux compatibility",
		Example: `
# Validate flux and kubernetes compatibility
gitops check
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeConfigArgs = run.GetKubeConfigArgs()
			kubeConfigArgs.AddFlags(cmd.Flags())

			cfg, err := kubeConfigArgs.ToRESTConfig()
			if err != nil {
				return err
			}

			c, err := discovery.NewDiscoveryClientForConfig(cfg)
			if err != nil {
				return cmderrors.ErrGetKubeClient
			}
			output, err := check.KubernetesVersion(c)
			if err != nil {
				return err
			}

			fmt.Println(output)

			return nil
		},
	}

	cmd.AddCommand(oidcconfig.OIDCConfigCommand(opts))

	return cmd
}
