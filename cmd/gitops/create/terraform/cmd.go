package terraform

import (
	"os"

	"github.com/flux-iac/tofu-controller/tfctl"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/run"
)

var kubeConfigArgs *genericclioptions.ConfigFlags

type commandFlags struct {
	Path     string
	Source   string
	Interval string
	Export   bool
}

var flags commandFlags

func Command(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "terraform",
		Aliases: []string{"tf"},
		Args:    cobra.ExactArgs(1),
		Short:   "Create a Terraform object",
		Long:    "Create a Terraform object",
		Example: `
# Create a Terraform resource in the default namespace
gitops create terraform -n default my-resource --source GitRepository/my-project --path ./terraform --interval 15m

# Create and export a Terraform resource manifest to the standard output
gitops create terraform -n default my-resource --source GitRepository/my-project --path ./terraform --interval 15m --export
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := tfctl.New("", "")

			namespace, err := cmd.Flags().GetString("namespace")
			if err != nil {
				return err
			}

			context, err := cmd.Flags().GetString("context")
			if err != nil {
				return err
			}

			export, err := cmd.Flags().GetBool("export")
			if err != nil {
				return err
			}

			kubeConfigArgs.Namespace = &namespace
			kubeConfigArgs.Context = &context

			v := viper.New()
			v.Set("namespace", namespace)
			if err := app.Init(kubeConfigArgs, v); err != nil {
				return err
			}

			return app.Create(
				os.Stdout,
				args[0],
				namespace,
				flags.Path,
				flags.Source,
				flags.Interval,
				export,
			)
		},
	}

	cmdFlags := cmd.Flags()
	cmdFlags.StringVar(&flags.Path, "path", "", "Path to the Terraform configuration")
	cmdFlags.StringVar(&flags.Source, "source", "", "Source of the Terraform configuration")
	cmdFlags.StringVar(&flags.Interval, "interval", "", "Interval at which the Terraform configuration should be applied")

	kubeConfigArgs = run.GetKubeConfigArgs()
	kubeConfigArgs.AddFlags(cmd.Flags())
	kubeConfigArgs.KubeConfig = &opts.Kubeconfig

	return cmd
}
