package run

import (
	"context"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/cmd/internal/config"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/run"
)

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
		_, clusterName, err := kube.RestConfig()
		if err != nil {
			return cmderrors.ErrNoCluster
		}

		kubeConfigOptions := run.GetKubeConfigOptions()
		kubeClientOptions := run.GetKubeClientOptions()

		kubeClient, err := run.GetKubeClient(clusterName, kubeConfigOptions, kubeClientOptions)
		if err != nil {
			return cmderrors.ErrGetKubeClient
		}

		log := internal.NewCLILogger(os.Stdout)

		ctx := context.Background()

		fluxVersion, err := run.GetFluxVersion(log, ctx, kubeClient)
		if err != nil {
			log.Failuref("Error getting Flux version: %v", err)

			fluxVersion = ""
		}

		if fluxVersion == "" {
			log.Actionf("Installing Flux...")

			err = run.InstallFlux(log, ctx, kubeClient, kubeConfigOptions)
			if err != nil {
				log.Failuref("Flux installation failed: %v", err)
				return err
			}
		}

		return nil
	}
}
