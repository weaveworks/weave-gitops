package root

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/validation"
	"log"
	"os"
	"strings"

	runclient "github.com/fluxcd/pkg/runtime/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
)

const defaultNamespace = "flux-system"

var (
	rootCmdFlags      = &config.Options{}
	kubeconfigArgs    = genericclioptions.NewConfigFlags(false)
	kubeclientOptions = new(runclient.Options)
)

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

			if nsErr := utils.ValidateNamespace(ns); nsErr != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", nsErr)
				os.Exit(1)
			}
			if rootCmdFlags.OverrideInCluster {
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
		},
	}

	// rootCmd.PersistentFlags().String("namespace", defaultNamespace, "The namespace scope for this operation")
	rootCmd.PersistentFlags().StringVarP(&rootCmdFlags.Endpoint, "endpoint", "e", os.Getenv("WEAVE_GITOPS_ENTERPRISE_API_URL"), "The Weave GitOps Enterprise HTTP API endpoint can be set with `WEAVE_GITOPS_ENTERPRISE_API_URL` environment variable")
	rootCmd.PersistentFlags().StringVarP(&rootCmdFlags.Username, "username", "u", "", "The Weave GitOps Enterprise username for authentication can be set with `WEAVE_GITOPS_USERNAME` environment variable")
	rootCmd.PersistentFlags().StringVarP(&rootCmdFlags.Password, "password", "p", "", "The Weave GitOps Enterprise password for authentication can be set with `WEAVE_GITOPS_PASSWORD` environment variable")
	rootCmd.PersistentFlags().BoolVar(&rootCmdFlags.OverrideInCluster, "override-in-cluster", false, "override running in cluster check")
	rootCmd.PersistentFlags().StringToStringVar(&rootCmdFlags.GitHostTypes, "git-host-types", map[string]string{}, "Specify which custom domains are running what (github or gitlab)")
	// rootCmd.PersistentFlags().BoolVar(&rootCmdFlags.InsecureSkipTLSVerify, "insecure-skip-tls-verify", false, "If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")
	// rootCmd.PersistentFlags().StringVar(&rootCmdFlags.Kubeconfig, "kubeconfig", "", "Paths to a kubeconfig. Only required if out-of-cluster.")

	configureDefaultNamespace()

	kubeconfigArgs.APIServer = nil // prevent AddFlags from configuring --server flag

	kubeconfigArgs.Timeout = nil // prevent AddFlags from configuring --request-timeout flag, we have --timeout instead

	kubeconfigArgs.AddFlags(rootCmd.PersistentFlags())

	// Since some subcommands use the `-s` flag as a short version for `--silent`, we manually configure the server flag
	// without the `-s` short version. While we're no longer on par with kubectl's flags, we maintain backwards compatibility
	// on the CLI interface.
	apiServer := ""
	kubeconfigArgs.APIServer = &apiServer
	rootCmd.PersistentFlags().StringVar(kubeconfigArgs.APIServer, "server", *kubeconfigArgs.APIServer, "The address and port of the Kubernetes API server")

	kubeclientOptions.BindFlags(rootCmd.PersistentFlags())

	cobra.CheckErr(rootCmd.PersistentFlags().MarkHidden("override-in-cluster"))
	cobra.CheckErr(rootCmd.PersistentFlags().MarkHidden("git-host-types"))

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(getCommand(rootCmdFlags))
	rootCmd.AddCommand(setCommand(rootCmdFlags))
	rootCmd.AddCommand(docsCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(betaCommand(rootCmdFlags))
	rootCmd.AddCommand(createCommand(rootCmdFlags))
	rootCmd.AddCommand(removeCommand(rootCmdFlags))

	return rootCmd
}

func configureDefaultNamespace() {
	*kubeconfigArgs.Namespace = defaultNamespace
	fromEnv := os.Getenv("FLUX_SYSTEM_NAMESPACE")

	if fromEnv != "" {
		// namespace must be a valid DNS label. Assess against validation
		// used upstream, and ignore invalid values as environment vars
		// may not be actively provided by end-user.
		if e := validation.IsDNS1123Label(fromEnv); len(e) > 0 {
			log.Printf("Ignoring invalid FLUX_SYSTEM_NAMESPACE value: %q", fromEnv)
			return
		}

		kubeconfigArgs.Namespace = &fromEnv
	}
}
