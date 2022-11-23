package root

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/weaveworks/weave-gitops/cmd/gitops/remove"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/weaveworks/weave-gitops/cmd/gitops/beta"
	"github.com/weaveworks/weave-gitops/cmd/gitops/check"
	cfg "github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/create"
	"github.com/weaveworks/weave-gitops/cmd/gitops/docs"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get"
	"github.com/weaveworks/weave-gitops/cmd/gitops/set"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/pkg/analytics"
	"github.com/weaveworks/weave-gitops/pkg/config"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"k8s.io/client-go/rest"
)

const defaultNamespace = "flux-system"

var options = &cfg.Options{}

// Only want AutomaticEnv to be called once!
func init() {
	// Setup flag to env mapping:
	//   config-repo => GITOPS_CONFIG_REPO
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("WEAVE_GITOPS")

	viper.AutomaticEnv()
}

func RootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:           "gitops",
		SilenceUsage:  true,
		SilenceErrors: true,
		Short:         "Weave GitOps",
		Long:          "Command line utility for managing Kubernetes applications via GitOps.",
		Example: `
  # Get help for gitops create dashboard command
  gitops create dashboard -h
  gitops help create dashboard

  # Get the version of gitops along with commit, branch, and flux version
  gitops version

  To learn more, you can find our documentation at https://docs.gitops.weave.works/
`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Sync flag values and env vars.
			err := viper.BindPFlags(cmd.Flags())
			if err != nil {
				log.Fatalf("Error binding viper to flags: %v", err)
			}

			ns, _ := cmd.Flags().GetString("namespace")

			if ns == "" {
				return
			}

			if nserr := utils.ValidateNamespace(ns); nserr != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", nserr)
				os.Exit(1)
			}
			if options.OverrideInCluster {
				kube.InClusterConfig = func() (*rest.Config, error) { return nil, rest.ErrNotInCluster }
			}

			err = cmd.Flags().Set("username", viper.GetString("username"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			err = cmd.Flags().Set("password", viper.GetString("password"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			var gitopsConfig *config.GitopsCLIConfig

			gitopsConfig, err = config.GetConfig(false)
			if err != nil {
				fmt.Println("To improve our product, we would like to collect analytics data. You can read more about what data we collect here: https://docs.gitops.weave.works/docs/feedback-and-telemetry/")

				prompt := promptui.Prompt{
					Label:     "Would you like to turn on analytics to help us improve our product",
					IsConfirm: true,
					Default:   "Y",
				}

				enableAnalytics := false

				// Answering "n" causes err to not be nil. Hitting enter without typing
				// does not return the default.
				_, err := prompt.Run()
				if err == nil {
					enableAnalytics = true
				}

				seed := time.Now().UnixNano()

				gitopsConfig = &config.GitopsCLIConfig{
					UserID:    config.GenerateUserID(10, seed),
					Analytics: enableAnalytics,
				}

				_ = config.SaveConfig(gitopsConfig)
			}

			if gitopsConfig.Analytics {
				_ = analytics.TrackCommand(cmd, gitopsConfig.UserID)
			}
		},
	}

	rootCmd.PersistentFlags().String("namespace", defaultNamespace, "The namespace scope for this operation")
	rootCmd.PersistentFlags().StringVarP(&options.Endpoint, "endpoint", "e", os.Getenv("WEAVE_GITOPS_ENTERPRISE_API_URL"), "The Weave GitOps Enterprise HTTP API endpoint can be set with `WEAVE_GITOPS_ENTERPRISE_API_URL` environment variable")
	rootCmd.PersistentFlags().StringVarP(&options.Username, "username", "u", "", "The Weave GitOps Enterprise username for authentication can be set with `WEAVE_GITOPS_USERNAME` environment variable")
	rootCmd.PersistentFlags().StringVarP(&options.Password, "password", "p", "", "The Weave GitOps Enterprise password for authentication can be set with `WEAVE_GITOPS_PASSWORD` environment variable")
	rootCmd.PersistentFlags().BoolVar(&options.OverrideInCluster, "override-in-cluster", false, "override running in cluster check")
	rootCmd.PersistentFlags().StringToStringVar(&options.GitHostTypes, "git-host-types", map[string]string{}, "Specify which custom domains are running what (github or gitlab)")
	rootCmd.PersistentFlags().BoolVar(&options.InsecureSkipTLSVerify, "insecure-skip-tls-verify", false, "If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")
	rootCmd.PersistentFlags().StringVar(&options.Kubeconfig, "kubeconfig", "", "Paths to a kubeconfig. Only required if out-of-cluster.")
	cobra.CheckErr(rootCmd.PersistentFlags().MarkHidden("override-in-cluster"))
	cobra.CheckErr(rootCmd.PersistentFlags().MarkHidden("git-host-types"))

	rootCmd.AddCommand(version.Cmd)
	rootCmd.AddCommand(get.GetCommand(options))
	rootCmd.AddCommand(set.SetCommand(options))
	rootCmd.AddCommand(docs.Cmd)
	rootCmd.AddCommand(check.Cmd)
	rootCmd.AddCommand(beta.GetCommand(options))
	rootCmd.AddCommand(create.GetCommand(options))
	rootCmd.AddCommand(remove.GetCommand(options))

	return rootCmd
}
