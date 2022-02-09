package root

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/weaveworks/weave-gitops/cmd/gitops/add"
	"github.com/weaveworks/weave-gitops/cmd/gitops/delete"
	"github.com/weaveworks/weave-gitops/cmd/gitops/docs"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get"
	"github.com/weaveworks/weave-gitops/cmd/gitops/install"
	"github.com/weaveworks/weave-gitops/cmd/gitops/ui"
	"github.com/weaveworks/weave-gitops/cmd/gitops/uninstall"
	"github.com/weaveworks/weave-gitops/cmd/gitops/upgrade"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"k8s.io/client-go/rest"
)

var options struct {
	endpoint          string
	overrideInCluster bool
	verbose           bool
	gitHostTypes      map[string]string
}

// Only want AutomaticEnv to be called once!
func init() {
	// Setup flag to env mapping:
	//   config-repo => GITOPS_CONFIG_REPO
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("GITOPS")

	viper.AutomaticEnv()
}

func RootCmd(client *resty.Client) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:           "gitops",
		SilenceUsage:  true,
		SilenceErrors: true,
		Short:         "Weave GitOps",
		Long:          "Command line utility for managing Kubernetes applications via GitOps.",
		Example: `
  # Get verbose output for any gitops command
  gitops [command] -v, --verbose

  # Get gitops app help
  gitops help app

  # Get the version of gitops along with commit, branch, and flux version
  gitops version

  To learn more, you can find our documentation at https://docs.gitops.weave.works/
`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			configureLogger()

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
			if options.overrideInCluster {
				kube.InClusterConfig = func() (*rest.Config, error) { return nil, rest.ErrNotInCluster }
			}
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&options.verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&options.endpoint, "endpoint", "e", os.Getenv("WEAVE_GITOPS_ENTERPRISE_API_URL"), "The Weave GitOps Enterprise HTTP API endpoint")
	rootCmd.PersistentFlags().BoolVar(&options.overrideInCluster, "override-in-cluster", false, "override running in cluster check")
	rootCmd.PersistentFlags().StringToStringVar(&options.gitHostTypes, "git-host-types", map[string]string{}, "Specify which custom domains are running what (github or gitlab)")
	cobra.CheckErr(rootCmd.PersistentFlags().MarkHidden("override-in-cluster"))
	cobra.CheckErr(rootCmd.PersistentFlags().MarkHidden("git-host-types"))

	rootCmd.AddCommand(install.Cmd)
	rootCmd.AddCommand(uninstall.Cmd)
	rootCmd.AddCommand(version.Cmd)
	rootCmd.AddCommand(ui.NewCommand())
	rootCmd.AddCommand(get.GetCommand(&options.endpoint, client))
	rootCmd.AddCommand(add.GetCommand(&options.endpoint, client))
	rootCmd.AddCommand(delete.DeleteCommand(&options.endpoint, client))
	rootCmd.AddCommand(upgrade.Cmd)
	rootCmd.AddCommand(docs.Cmd)

	return rootCmd
}

func configureLogger() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if options.verbose {
		log.SetLevel(log.DebugLevel)
	}
}
