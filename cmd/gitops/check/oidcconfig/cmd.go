package oidcconfig

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/oidc/check"
	"github.com/weaveworks/weave-gitops/pkg/run"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// OIDCConfigCommand returns the cobra command for running `oidc-config`.
func OIDCConfigCommand(opts *config.Options) *cobra.Command {
	var (
		kubeConfigArgs    *genericclioptions.ConfigFlags
		fromSecretFlag    string
		skipSecretFlag    bool
		clientIDFlag      string
		clientSecretFlag  string
		scopesFlag        []string
		claimUsernameFlag string
		claimGroupsFlag   string
		issuerURLFlag     string
	)

	cmd := &cobra.Command{
		Use:   "oidc-config",
		Short: "Check an OIDC configuration for proper functionality.",
		Long: `This command will send the user through an OIDC authorization code flow using the given OIDC configuration. This is helpful for verifying that a given configuration will work properly with Weave GitOps or for debugging issues. Without any provided flags it will read the configuration from a Secret on the cluster.

NOTE: Make sure to configure your OIDC provider so that it accepts "http://localhost:9876" as redirect URI.`,
		Example: `
# Check the OIDC configuration stored in the flux-system/oidc-auth Secret
gitops check oidc-config

# Check a different set of scopes
gitops check oidc-config --scopes=openid,groups

# Check a different username cliam
gitops check oidc-config --claim-username=sub

# Check configuration without fetching a Secret from the cluster
gitops check oidc-config --skip-secret --client-id=CID --client-secret=SEC --issuer-url=https://example.org
		`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if skipSecretFlag {
				fromSecretFlag = "" // the skip flag overrides this one.
			}

			cfg, err := kubeConfigArgs.ToRESTConfig()
			if err != nil {
				return err
			}

			if fromSecretFlag == "" {
				if clientIDFlag == "" ||
					clientSecretFlag == "" ||
					issuerURLFlag == "" {
					return fmt.Errorf("when not reading OIDC configuration from a Secret, you need to provide " +
						"client ID, client secret and issuer URL using the respective command-line flags")
				}
			}

			log := logger.NewCLILogger(os.Stdout)
			kubeClient, err := run.GetKubeClient(log, *kubeConfigArgs.Context, cfg, nil)
			if err != nil {
				return cmderrors.ErrGetKubeClient
			}

			ns, err := cmd.Flags().GetString("namespace")
			if err != nil {
				return fmt.Errorf("failed getting namespace flag: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()
			claims, err := check.GetPrincipal(ctx, check.Options{
				ClientID:        clientIDFlag,
				ClientSecret:    clientSecretFlag,
				IssuerURL:       issuerURLFlag,
				SecretName:      fromSecretFlag,
				SecretNamespace: ns,
				Scopes:          scopesFlag,
				ClaimUsername:   claimUsernameFlag,
				ClaimGroups:     claimGroupsFlag,
			}, log, kubeClient)
			if err != nil {
				return fmt.Errorf("failed getting claims: %w", err)
			}

			log.Println("user: %s", claims.ID)
			if len(claims.Groups) != 0 {
				log.Println("groups: %s", strings.Join(claims.Groups, ", "))
			} else {
				log.Println("no groups claim")
			}

			return nil
		},
		DisableAutoGenTag: true,
	}

	kubeConfigArgs = run.GetKubeConfigArgs()
	kubeConfigArgs.AddFlags(cmd.Flags())

	cmd.Flags().StringVar(&clientIDFlag, "client-id", "", "OIDC client ID")
	cmd.Flags().StringVar(&clientSecretFlag, "client-secret", "", "OIDC client secret")
	cmd.Flags().StringVar(&issuerURLFlag, "issuer-url", "", "OIDC issuer URL")
	cmd.Flags().StringVar(&fromSecretFlag, "from-secret", "oidc-auth", "Get OIDC configuration from the given Secret resource")
	cmd.Flags().BoolVar(&skipSecretFlag, "skip-secret", false,
		"Do not read OIDC configuration from a Kubernetes Secret but rely solely on the values from the given flags.")
	cmd.Flags().StringVar(&claimUsernameFlag, "username-claim", "", "ID token claim to use for the user name.")
	cmd.Flags().StringVar(&claimGroupsFlag, "groups-claim", "", "ID token claim to use for the groups.")
	cmd.Flags().StringSliceVar(&scopesFlag, "scopes", nil, fmt.Sprintf("OIDC scopes to request (default [%s])", strings.Join(auth.DefaultScopes, ",")))

	return cmd
}
