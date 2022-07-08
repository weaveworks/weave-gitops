package run

import (
	"context"
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/cmd/internal/config"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/run"
)

type runCommandFlags struct{}

// TODO: Add flags when adding the actual run command.
var flags runCommandFlags //nolint

func RunCommand(opts *config.Options, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Set up an interactive sync between your cluster and your local file system",
		Long:  "This will set up a sync between the cluster in your kubeconfig and the path that you specify on your local filesystem.  If you do not have Flux installed on the cluster then this will add it to the cluster automatically.  This is a requirement so we can sync the files successfully from your local system onto the cluster.  Flux will take care of producing the objects for you.",
		Example: `
# Run the sync on the current working directory
gitops beta run . [flags]

# Run the sync against the dev overlay path
gitops beta run ./deploy/overlays/dev [flags]`,
		SilenceUsage:      true,
		SilenceErrors:     true,
		PreRunE:           betaRunCommandPreRunE(&opts.Endpoint),
		RunE:              betaRunCommandRunE(opts, client),
		DisableAutoGenTag: true,
	}

	return cmd
}

func betaRunCommandPreRunE(endpoint *string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		numArgs := len(args)

		if numArgs == 0 {
			return cmderrors.ErrNoFilePath
		}

		if numArgs > 1 {
			return cmderrors.ErrMultipleFilePaths
		}

		return nil
	}
}

func betaRunCommandRunE(opts *config.Options, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cfg, clusterName, err := kube.RestConfig()
		if err != nil {
			return cmderrors.ErrNoCluster
		}

		kubeClient, err := kube.NewKubeHTTPClientWithConfig(cfg, clusterName)
		if err != nil {
			return cmderrors.ErrGetKubeClient
		}

		log := internal.NewCLILogger(os.Stdout)

		ctx := context.Background()

		fluxVersion, err := run.GetFluxVersion(log, ctx, kubeClient)
		if err != nil {
			fmt.Println("error getting flux version", err)

			fluxVersion = ""
		}

		if fluxVersion == "" {
			err = run.InstallFlux(log, ctx, kubeClient)
			if err != nil {
				return fmt.Errorf("flux installation failed: %w", err)
			}
		}

		return nil
	}
}
