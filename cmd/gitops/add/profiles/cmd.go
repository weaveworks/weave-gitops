package profiles

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/profiles"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var opts profiles.AddOptions

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
	cmd.Flags().StringVar(&opts.ConfigRepo, "config-repo", "", "URL of external repository (if any) which will hold automation manifests")
	cmd.Flags().StringVar(&opts.Cluster, "cluster", "", "Name of the cluster to add the profile to")
	cmd.Flags().StringVar(&opts.ProfilesPort, "profiles-port", server.DefaultPort, "Port the Profiles API is running on")
	cmd.Flags().BoolVar(&opts.AutoMerge, "auto-merge", false, "If set, 'gitops add profile' will merge automatically into the repository's default branch")
	cmd.Flags().StringVar(&opts.Kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "Absolute path to the kubeconfig file")

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
		rand.Seed(time.Now().UnixNano())
		log := internal.NewCLILogger(os.Stdout)
		fluxClient := flux.New(osys.New(), &runner.CLIRunner{})
		factory := services.NewFactory(fluxClient, log)
		providerClient := internal.NewGitProviderClient(os.Stdout, os.LookupEnv, auth.NewAuthCLIHandler, log)

		if err := validateAddOptions(opts); err != nil {
			return err
		}

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			return err
		}
		opts.Namespace = ns

		config, err := clientcmd.BuildConfigFromFlags("", opts.Kubeconfig)
		if err != nil {
			return fmt.Errorf("error initializing kubernetes config: %w", err)
		}

		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			return fmt.Errorf("error initializing kubernetes client: %w", err)
		}

		kubeClient, _, err := kube.NewKubeHTTPClient()
		if err != nil {
			return fmt.Errorf("failed to create kube client: %w", err)
		}

		ctx := context.Background()
		_, gitProvider, err := factory.GetGitClients(ctx, kubeClient, providerClient, services.GitConfigParams{
			ConfigRepo:       opts.ConfigRepo,
			Namespace:        opts.Namespace,
			IsHelmRepository: true,
			DryRun:           false,
		})
		if err != nil {
			return fmt.Errorf("failed to get git clients: %w", err)
		}

		return profiles.NewService(clientSet, log).Add(ctx, gitProvider, opts)
	}
}

func validateAddOptions(opts profiles.AddOptions) error {
	if models.ApplicationNameTooLong(opts.Name) {
		return fmt.Errorf("--name value is too long: %s; must be <= %d characters",
			opts.Name, models.MaxKubernetesResourceNameLength)
	}

	if opts.Version != "latest" {
		if _, err := semver.StrictNewVersion(opts.Version); err != nil {
			return fmt.Errorf("error parsing --version=%s: %w", opts.Version, err)
		}
	}
	return nil
}
