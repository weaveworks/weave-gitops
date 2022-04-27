package profiles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/gitops/pkg/names"
	"github.com/weaveworks/weave-gitops/gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/gitops/pkg/services/profiles"
	"github.com/weaveworks/weave-gitops/common/pkg/kube"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var opts profiles.Options

// AddCommand provides support for adding a profile to a cluster.
func AddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "profile",
		Short:         "Add a profile to a cluster",
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: `
		# Add a profile to a cluster
		gitops add profile --name=podinfo --cluster=prod --version=1.0.0 --config-repo=ssh://git@github.com/owner/config-repo.git
		`,
		RunE: addProfileCmdRunE(),
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Name of the profile")
	cmd.Flags().StringVar(&opts.Version, "version", "latest", "Version of the profile specified as semver (e.g.: 0.1.0) or as 'latest'")
	cmd.Flags().StringVar(&opts.ConfigRepo, "config-repo", "", "URL of the external repository that contains the automation manifests")
	cmd.Flags().StringVar(&opts.Cluster, "cluster", "", "Name of the cluster to add the profile to")
	cmd.Flags().BoolVar(&opts.AutoMerge, "auto-merge", false, "If set, 'gitops add profile' will merge automatically into the repository's branch")
	cmd.Flags().StringVar(&opts.Kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "Absolute path to the kubeconfig file")
	internal.AddPRFlags(cmd, &opts.HeadBranch, &opts.BaseBranch, &opts.Description, &opts.Message, &opts.Title)

	requiredFlags := []string{"name", "config-repo", "cluster"}
	for _, f := range requiredFlags {
		if err := cobra.MarkFlagRequired(cmd.Flags(), f); err != nil {
			panic(fmt.Errorf("unexpected error: %w", err))
		}
	}

	return cmd
}

func addProfileCmdRunE() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		log := internal.NewCLILogger(os.Stdout)
		fluxClient := flux.New(&runner.CLIRunner{})
		factory := services.NewFactory(fluxClient, internal.Logr())
		providerClient := internal.NewGitProviderClient(os.Stdout, os.LookupEnv, log)

		if err := validateOptions(opts); err != nil {
			return err
		}

		var err error
		if opts.Namespace, err = cmd.Flags().GetString("namespace"); err != nil {
			return err
		}

		config, err := clientcmd.BuildConfigFromFlags("", opts.Kubeconfig)
		if err != nil {
			return fmt.Errorf("error initializing kubernetes config: %w", err)
		}

		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			return fmt.Errorf("error initializing kubernetes client: %w", err)
		}

		kubeClient, err := kube.NewKubeHTTPClient()
		if err != nil {
			return fmt.Errorf("failed to create kube client: %w", err)
		}

		_, gitProvider, err := factory.GetGitClients(context.Background(), kubeClient, providerClient, services.GitConfigParams{
			ConfigRepo:       opts.ConfigRepo,
			Namespace:        opts.Namespace,
			IsHelmRepository: true,
			DryRun:           false,
		})
		if err != nil {
			return fmt.Errorf("failed to get git clients: %w", err)
		}

		return profiles.NewService(clientSet, log).Add(context.Background(), gitProvider, opts)
	}
}

func validateOptions(opts profiles.Options) error {
	if names.ApplicationNameTooLong(opts.Name) {
		return fmt.Errorf("--name value is too long: %s; must be <= %d characters",
			opts.Name, names.MaxKubernetesResourceNameLength)
	}

	if opts.Version != "latest" {
		if _, err := semver.StrictNewVersion(opts.Version); err != nil {
			return fmt.Errorf("error parsing --version=%s: %w", opts.Version, err)
		}
	}

	return nil
}
