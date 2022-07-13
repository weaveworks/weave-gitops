package run

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fluxcd/flux2/pkg/manifestgen/install"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/run"
)

type runCommandFlags struct {
	FluxVersion     string
	AllowK8sContext string
	Components      []string
	ComponentsExtra []string
	Timeout         time.Duration
	KubeConfig      string
	Context         string
	Cluster         string
	// global flags
	Namespace string
}

var flags runCommandFlags

func RunCommand(opts *config.Options) *cobra.Command {
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
		RunE:              betaRunCommandRunE(opts),
		DisableAutoGenTag: true,
	}

	cmd.Flags().StringVar(&flags.FluxVersion, "flux-version", "", "")
	cmd.Flags().StringVar(&flags.AllowK8sContext, "allow-k8s-context", "", "")
	cmd.Flags().StringSliceVar(&flags.Components, "components", flags.Components, "")
	cmd.Flags().StringSliceVar(&flags.ComponentsExtra, "components-extra", flags.ComponentsExtra, "")
	cmd.Flags().DurationVar(&flags.Timeout, "timeout", flags.Timeout, "")
	cmd.Flags().StringVar(&flags.KubeConfig, "kube-config", "", "")
	cmd.Flags().StringVar(&flags.Context, "context", "", "")
	cmd.Flags().StringVar(&flags.Cluster, "cluster", "", "")

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

func betaRunCommandRunE(opts *config.Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		if flags.Namespace, err = cmd.Flags().GetString("namespace"); err != nil {
			return err
		}

		log := internal.NewCLILogger(os.Stdout)

		log.Actionf("Checking for a cluster in the kube config ...")

		_, clusterName, err := kube.RestConfig()
		if err != nil {
			log.Failuref("Error getting a restconfig: %v", err.Error())
			return cmderrors.ErrNoCluster
		}

		kubeConfigArgs := run.GetKubeConfigArgs()
		kubeClientOpts := run.GetKubeClientOptions()

		kubeClient, err := run.GetKubeClient(log, clusterName, kubeConfigArgs, kubeClientOpts)
		if err != nil {
			return cmderrors.ErrGetKubeClient
		}

		contextName := kubeClient.ClusterName
		if flags.AllowK8sContext == contextName {
			log.Infow("Explicitly allow GitOps Run on %s context", contextName)
		} else if !isLocalCluster(kubeClient) {
			return errors.New("allowed to run against a local cluster only")
		}

		ctx := context.Background()

		log.Actionf("Checking if Flux is already installed ...")

		if fluxVersion, err := run.GetFluxVersion(log, ctx, kubeClient); err != nil {
			log.Warningf("Flux is not found: %v", err.Error())

			installOpts := install.Options{
				BaseURL:         install.MakeDefaultOptions().BaseURL,
				Version:         flags.FluxVersion,
				Namespace:       flags.Namespace,
				Components:      flags.Components,
				ComponentsExtra: flags.ComponentsExtra,
				ManifestFile:    "flux-system.yaml",
				Timeout:         flags.Timeout,
			}

			if err := run.InstallFlux(log, ctx, kubeClient, installOpts, kubeConfigArgs); err != nil {
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
