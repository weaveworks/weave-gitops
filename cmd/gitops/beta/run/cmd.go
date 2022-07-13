package run

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

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

// isLocalCluster checks if it's a local cluster. See https://skaffold.dev/docs/environment/local-cluster/
func isLocalCluster(kubeClient *kube.KubeHTTP) bool {
	const (
		// cluster context prefixes
		kindPrefix = "kind-"
		k3dPrefix  = "k3d-"

		// cluster context names
		minikube         = "minikube"
		dockerForDesktop = "docker-for-desktop"
		dockerDesktop    = "docker-desktop"
	)

	if strings.HasPrefix(kubeClient.ClusterName, kindPrefix) ||
		strings.HasPrefix(kubeClient.ClusterName, k3dPrefix) ||
		kubeClient.ClusterName == minikube ||
		kubeClient.ClusterName == dockerForDesktop ||
		kubeClient.ClusterName == dockerDesktop {
		return true
	} else {
		return false
	}
}

func betaRunCommandRunE(opts *config.Options, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		log := internal.NewCLILogger(os.Stdout)

		log.Actionf("Checking for a cluster in the kube config ...")

		_, clusterName, err := kube.RestConfig()
		if err != nil {
			log.Failuref("Error getting a restconfig: %v", err.Error())
			return cmderrors.ErrNoCluster
		}

		kubeConfigOptions := run.GetKubeConfigOptions()
		kubeClientOptions := run.GetKubeClientOptions()

		kubeClient, err := run.GetKubeClient(log, clusterName, kubeConfigOptions, kubeClientOptions)
		if err != nil {
			return cmderrors.ErrGetKubeClient
		}

		if !isLocalCluster(kubeClient) {
			return errors.New("allowed to run against a local cluster only")
		}

		ctx := context.Background()

		log.Actionf("Checking if Flux is already installed ...")

		if fluxVersion, err := run.GetFluxVersion(log, ctx, kubeClient); err != nil {
			log.Warningf("Flux is not found: %v", err.Error())

			if err := run.InstallFlux(log, ctx, kubeClient, kubeConfigOptions); err != nil {
				return fmt.Errorf("flux installation failed: %w", err)
			} else {
				log.Successf("Flux has been installed")
			}
		} else {
			log.Successf("Flux version %s is found", fluxVersion)
		}

		const fluxSystemNS = "flux-system"
		for _, controllerName := range []string{"source-controller", "kustomize-controller", "helm-controller", "notification-controller"} {
			log.Actionf("Waiting for %s/%s to be ready ...", fluxSystemNS, controllerName)

			if err := run.WaitForDeploymentToBeReady(log, kubeClient, controllerName, fluxSystemNS); err != nil {
				return err
			}

			log.Successf("%s/%s is now ready ...", fluxSystemNS, controllerName)
		}

		if err := run.InstallBucketServer(log, kubeClient); err != nil {
			return err
		}

		return nil
	}
}
