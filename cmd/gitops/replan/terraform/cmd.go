package terraform

import (
	"context"
	"os"

	"github.com/flux-iac/tofu-controller/tfctl"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/run"
)

var kubeConfigArgs *genericclioptions.ConfigFlags

func Command(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "terraform",
		Aliases: []string{"tf"},
		Args:    cobra.ExactArgs(1),
		Short:   "Trigger replan for a Terraform object",
		Example: `
# Replan the Terraform plan of a Terraform object from the "flux-system" namespace
gitops replan terraform --namespace flux-system my-resource
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := tfctl.New("", "")

			namespace, err := cmd.Flags().GetString("namespace")
			if err != nil {
				return err
			}

			kubeCtx, err := cmd.Flags().GetString("context")
			if err != nil {
				return err
			}

			kubeConfigArgs.Namespace = &namespace
			kubeConfigArgs.Context = &kubeCtx

			v := viper.New()
			v.Set("namespace", namespace)
			if err := app.Init(kubeConfigArgs, v); err != nil {
				return err
			}

			return app.Replan(
				context.TODO(),
				os.Stdout,
				args[0],
			)
		},
	}

	kubeConfigArgs = run.GetKubeConfigArgs()
	kubeConfigArgs.AddFlags(cmd.Flags())
	kubeConfigArgs.KubeConfig = &opts.Kubeconfig

	return cmd
}
