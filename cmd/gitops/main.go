package main

import (
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/add"
	"github.com/weaveworks/weave-gitops/cmd/gitops/app"
	"github.com/weaveworks/weave-gitops/cmd/gitops/flux"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get"
	"github.com/weaveworks/weave-gitops/cmd/gitops/install"
	"github.com/weaveworks/weave-gitops/cmd/gitops/ui"
	"github.com/weaveworks/weave-gitops/cmd/gitops/uninstall"
	"github.com/weaveworks/weave-gitops/cmd/gitops/upgrade"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	fluxBin "github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

var options struct {
	verbose  bool
	endpoint string
}

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

  # Add application to gitops control from a local git repository
  gitops app add . --name <myapp>
  OR
  gitops app add <myapp-directory>

  # Add application to gitops control from a github repository
  gitops app add \
    --name <myapp> \
    --url git@github.com:myorg/<myapp> \
    --branch prod-<myapp>

  # Get status of application under gitops control
  gitops app status podinfo

  # Get help for gitops app add command
  gitops app add -h
  gitops help app add

  # Show manifests that would be installed by the gitops gitops install command
  gitops gitops install --dry-run

  # Install gitops in the wego-system namespace
  gitops install

  # Get the version of gitops along with commit, branch, and flux version
  gitops version

  To learn more, you can find our documentation at https://docs.gitops.weave.works/
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		configureLogger()

		ns, _ := cmd.Flags().GetString("namespace")

		if ns == "" {
			return
		}

		if nserr := utils.ValidateNamespace(ns); nserr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", nserr)
			os.Exit(1)
		}
	},
}

func configureLogger() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if options.verbose {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	restyClient := resty.New()
	cliRunner := &runner.CLIRunner{}
	osysClient := osys.New()
	fluxClient := fluxBin.New(osysClient, cliRunner)
	fluxClient.SetupBin()
	rootCmd.PersistentFlags().BoolVarP(&options.verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().String("namespace", "wego-system", "gitops runtime namespace")
	rootCmd.PersistentFlags().StringVarP(&options.endpoint, "endpoint", "e", os.Getenv("WEAVE_GITOPS_ENTERPRISE_API_URL"), "The Weave GitOps Enterprise HTTP API endpoint")

	rootCmd.AddCommand(install.Cmd)
	rootCmd.AddCommand(uninstall.Cmd)
	rootCmd.AddCommand(version.Cmd)
	rootCmd.AddCommand(flux.Cmd)
	rootCmd.AddCommand(ui.Cmd)
	rootCmd.AddCommand(app.ApplicationCmd)
	rootCmd.AddCommand(get.GetCommand(&options.endpoint, restyClient))
	rootCmd.AddCommand(add.GetCommand(&options.endpoint, restyClient))
	rootCmd.AddCommand(upgrade.Cmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
