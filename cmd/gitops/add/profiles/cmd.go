package profiles

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"github.com/weaveworks/weave-gitops/pkg/services/profiles"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var opts profiles.AddOptions

// AddCommand provides support for adding a profile to a cluster.
func AddCommand(client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "profile",
		Aliases:       []string{"profiles"},
		Short:         "Add a profile to a cluster",
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: `
		# Add a profile to a cluster
		gitops add profile --name=podinfo --cluster=prod --version=1.0.0 --config-repo=ssh://git@github.com/owner/config-repo.git
		`,
		RunE: addProfileCmdRunE(client),
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Name of the profile")
	cmd.Flags().StringVar(&opts.Version, "version", "latest", "Version of the profile")
	cmd.Flags().StringVar(&opts.ConfigRepo, "config-repo", "", "URL of external repository (if any) which will hold automation manifests")
	cmd.Flags().StringVar(&opts.Cluster, "cluster", "", "Name of the cluster to add the profile to")
	cmd.Flags().StringVar(&opts.Port, "port", server.DefaultPort, "Port the profiles API is running on")
	cmd.Flags().BoolVar(&opts.AutoMerge, "auto-merge", false, "If set, 'gitops add profile' will merge automatically into the set --branch")

	return cmd
}

func addProfileCmdRunE(client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		log := internal.NewCLILogger(os.Stdout)
		fluxClient := flux.New(osys.New(), &runner.CLIRunner{})
		factory := services.NewFactory(fluxClient, log)
		providerClient := internal.NewGitProviderClient(os.Stdout, os.LookupEnv, auth.NewAuthCLIHandler, log)

		validatedOpts, err := validateAddOptions(opts)
		if err != nil {
			return err
		}

		ns, err := cmd.Parent().Parent().Flags().GetString("namespace")
		if err != nil {
			return err
		}
		opts.Namespace = ns

		ctx := context.Background()
		_, gitProvider, err := factory.GetGitClients(ctx, providerClient, services.GitConfigParams{
			ConfigRepo:       opts.ConfigRepo,
			Namespace:        opts.Namespace,
			IsHelmRepository: true,
			DryRun:           false,
		})
		if err != nil {
			return fmt.Errorf("failed to get git clients: %w", err)
		}

		config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
		if err != nil {
			return fmt.Errorf("error initializing kubernetes config: %w", err)
		}

		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			return fmt.Errorf("error initializing kubernetes client: %w", err)
		}

		profilesService := profiles.NewService(clientSet)
		return profilesService.Add(ctx, gitProvider, validatedOpts)
	}
}

func validateAddOptions(opts profiles.AddOptions) (profiles.AddOptions, error) {
	if opts.Name == "" {
		return opts, errors.New("--name should be provided")
	}

	if automation.ApplicationNameTooLong(opts.Name) {
		return opts, fmt.Errorf("--name value is too long: %s; must be <= %d characters",
			opts.Name, automation.MaxKubernetesResourceNameLength)
	}

	if strings.HasPrefix(opts.Name, "wego") {
		return opts, fmt.Errorf("the prefix 'wego' is used by weave gitops and is not allowed for a profile name")
	}

	if opts.ConfigRepo == "" {
		return opts, errors.New("--config-repo should be provided")
	}
	if opts.Cluster == "" {
		return opts, errors.New("--cluster should be provided")
	}
	if _, err := semver.StrictNewVersion(opts.Version); err != nil {
		return opts, fmt.Errorf("error parsing --version=%s: %s", opts.Version, err)
	}
	return opts, nil
}
