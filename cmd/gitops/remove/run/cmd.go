package run

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run"
	"github.com/weaveworks/weave-gitops/pkg/run/install"
	"github.com/weaveworks/weave-gitops/pkg/run/session"
	"github.com/weaveworks/weave-gitops/pkg/run/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
)

type RunCommandFlags struct {
	AllSessions bool
	NoSession   bool

	// Global flags.
	Namespace  string
	KubeConfig string

	// Flags, created by genericclioptions.
	Context string
}

var flags RunCommandFlags

var kubeConfigArgs *genericclioptions.ConfigFlags

func RunCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Remove GitOps Run sessions",
		Long:  "Remove GitOps Run sessions",
		Example: `
# Remove the GitOps Run session "dev-1234" from the "flux-system" namespace
gitops remove run --namespace flux-system dev-1234

# Remove all GitOps Run sessions from the default namespace
gitops remove run --all-sessions

# Remove all GitOps Run sessions from the dev namespace
gitops remove run -n dev --all-sessions

# Clean up resources from a failed GitOps Run in no session mode
gitops remove run --no-session
`,
		PreRunE: removeRunPreRunE(opts),
		RunE:    removeRunRunE(opts),

		SilenceUsage:      true,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}

	cmdFlags := cmd.Flags()

	cmdFlags.BoolVar(&flags.AllSessions, "all-sessions", false, "Remove all GitOps Run sessions")
	cmdFlags.BoolVar(&flags.NoSession, "no-session", false, "Remove all GitOps Run components in the non-session mode")

	kubeConfigArgs = run.GetKubeConfigArgs()

	kubeConfigArgs.AddFlags(cmd.Flags())

	return cmd
}

func getKubeClient(cmd *cobra.Command) (*kube.KubeHTTP, *rest.Config, error) {
	var err error

	log := logger.NewCLILogger(os.Stdout)

	if flags.Namespace, err = cmd.Flags().GetString("namespace"); err != nil {
		return nil, nil, err
	}

	kubeConfigArgs.Namespace = &flags.Namespace

	if flags.KubeConfig, err = cmd.Flags().GetString("kubeconfig"); err != nil {
		return nil, nil, err
	}

	if flags.Context, err = cmd.Flags().GetString("context"); err != nil {
		return nil, nil, err
	}

	if flags.KubeConfig != "" {
		kubeConfigArgs.KubeConfig = &flags.KubeConfig

		if flags.Context == "" {
			log.Failuref("A context should be provided if a kubeconfig is provided")
			return nil, nil, cmderrors.ErrNoContextForKubeConfig
		}
	}

	var contextName string

	if flags.Context != "" {
		contextName = flags.Context
	} else {
		_, contextName, err = kube.RestConfig()
		if err != nil {
			log.Failuref("Error getting a restconfig: %v", err.Error())
			return nil, nil, cmderrors.ErrNoCluster
		}
	}

	cfg, err := kubeConfigArgs.ToRESTConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting a restconfig from kube config args: %w", err)
	}

	kubeClientOpts := run.GetKubeClientOptions()
	kubeClientOpts.BindFlags(cmd.Flags())

	kubeClient, err := run.GetKubeClient(log, contextName, cfg, kubeClientOpts)
	if err != nil {
		return nil, nil, cmderrors.ErrGetKubeClient
	}

	return kubeClient, cfg, nil
}

func removeRunPreRunE(opts *config.Options) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// if flags.NoSession is set, we don't need to check for session name
		if flags.NoSession {
			return nil
		}

		numArgs := len(args)
		if numArgs == 0 && !flags.AllSessions {
			return cmderrors.ErrSessionNameIsRequired
		}

		return nil
	}
}

func removeRunRunE(opts *config.Options) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		kubeClient, _, err := getKubeClient(cmd)
		if err != nil {
			return err
		}

		log := logger.NewCLILogger(os.Stdout)
		ctx := context.Background()

		if flags.NoSession {
			if err := watch.CleanupBucketSourceAndHelm(ctx, log, kubeClient, flags.Namespace); err != nil {
				return err
			}

			if err := watch.CleanupBucketSourceAndKS(ctx, log, kubeClient, flags.Namespace); err != nil {
				return err
			}

			if err := watch.UninstallDevBucketServer(ctx, log, kubeClient); err != nil {
				return err
			}

			if err := install.UninstallFluentBit(ctx, log, kubeClient, flags.Namespace, install.FluentBitHRName); err != nil {
				return err
			}
		} else if flags.AllSessions {
			internalSessions, listErr := session.List(kubeClient, flags.Namespace)
			if listErr != nil {
				return listErr
			}

			for _, internalSession := range internalSessions {
				log.Actionf("Removing session %s/%s ...", internalSession.SessionNamespace, internalSession.SessionName)

				if err := session.Remove(kubeClient, internalSession); err != nil {
					return err
				}

				log.Successf("Session %s/%s was successfully removed.", internalSession.SessionNamespace, internalSession.SessionName)
			}
		} else {
			internalSession, err := session.Get(kubeClient, args[0], flags.Namespace)
			if err != nil {
				return err
			}
			log.Actionf("Removing session %s/%s ...", internalSession.SessionNamespace, internalSession.SessionName)
			if err := session.Remove(kubeClient, internalSession); err != nil {
				return err
			}
			log.Successf("Session %s/%s was successfully removed.", internalSession.SessionNamespace, internalSession.SessionName)
		}

		return nil
	}
}
