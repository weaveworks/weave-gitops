package root

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/cmd/gitops/add"
	"github.com/weaveworks/weave-gitops/cmd/gitops/check"
	"github.com/weaveworks/weave-gitops/cmd/gitops/delete"
	"github.com/weaveworks/weave-gitops/cmd/gitops/docs"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get"
	"github.com/weaveworks/weave-gitops/cmd/gitops/install"
	"github.com/weaveworks/weave-gitops/cmd/gitops/ui"
	"github.com/weaveworks/weave-gitops/cmd/gitops/update"
	"github.com/weaveworks/weave-gitops/cmd/gitops/upgrade"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"k8s.io/client-go/rest"
)

var options struct {
	endpoint              string
	overrideInCluster     bool
	verbose               bool
	gitHostTypes          map[string]string
	insecureSkipTlsVerify bool
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
		Example: fmt.Sprintf(`
  # Get verbose output for any gitops command
  gitops [command] -v, --verbose

  # Get gitops app help
  gitops help app

  # Add application to gitops control from a local git repository
  gitops add app . --name <myapp>
  OR
  gitops add app <myapp-directory>

  # Add application to gitops control from a github repository
  gitops add app \
    --name <myapp> \
    --url git@github.com:myorg/<myapp> \
    --branch prod-<myapp>

  # Get status of application under gitops control
  gitops get app podinfo

  # Get help for gitops add app command
  gitops add app -h
  gitops help add app

  # Show manifests that would be installed by the gitops install command
  gitops install --dry-run

  # Install gitops in the %s namespace
  gitops install

  # Get the version of gitops along with commit, branch, and flux version
  gitops version

  To learn more, you can find our documentation at https://docs.gitops.weave.works/
`, wego.DefaultNamespace),
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
			if options.insecureSkipTlsVerify {
				client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
			}
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&options.verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().String("namespace", wego.DefaultNamespace, "The namespace scope for this operation")
	rootCmd.PersistentFlags().StringVarP(&options.endpoint, "endpoint", "e", os.Getenv("WEAVE_GITOPS_ENTERPRISE_API_URL"), "The Weave GitOps Enterprise HTTP API endpoint")
	rootCmd.PersistentFlags().BoolVar(&options.overrideInCluster, "override-in-cluster", false, "override running in cluster check")
	rootCmd.PersistentFlags().StringToStringVar(&options.gitHostTypes, "git-host-types", map[string]string{}, "Specify which custom domains are running what (github or gitlab)")
	rootCmd.PersistentFlags().BoolVar(&options.insecureSkipTlsVerify, "insecure-skip-tls-verify", false, "If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")
	cobra.CheckErr(rootCmd.PersistentFlags().MarkHidden("override-in-cluster"))
	cobra.CheckErr(rootCmd.PersistentFlags().MarkHidden("git-host-types"))

	rootCmd.AddCommand(install.Cmd)
	rootCmd.AddCommand(version.Cmd)
	rootCmd.AddCommand(get.GetCommand(&options.endpoint, client))
	rootCmd.AddCommand(add.GetCommand(&options.endpoint, client))
	rootCmd.AddCommand(update.UpdateCommand(&options.endpoint, client))
	rootCmd.AddCommand(delete.DeleteCommand(&options.endpoint, client))
	rootCmd.AddCommand(upgrade.Cmd)
	rootCmd.AddCommand(ui.Cmd)
	rootCmd.AddCommand(docs.Cmd)
	rootCmd.AddCommand(check.Cmd)

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
