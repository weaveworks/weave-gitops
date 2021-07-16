package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/wego/app"
	"github.com/weaveworks/weave-gitops/cmd/wego/flux"
	"github.com/weaveworks/weave-gitops/cmd/wego/gitops"
	"github.com/weaveworks/weave-gitops/cmd/wego/version"
	fluxBin "github.com/weaveworks/weave-gitops/pkg/flux"
)

var options struct {
	verbose bool
}

var rootCmd = &cobra.Command{
	Use: "wego",
	SilenceUsage:  true,
	SilenceErrors: true,
	Short:         "Weave GitOps",
	Long:          "Command line utility for managing Kubernetes applications via GitOps.",
	Example:`
  # Get verbose output for any wego command
  wego [command] -v, --verbose

  # Get wego app help
  wego help app

  # Add application to wego control from a local git repository
  wego app add . --name <myapp>
  OR
  wego app add <myapp-directory>

  # Add application to wego control from a github repository
  wego app add \
	--name <myapp> \
	--url git@github.com:myorg/<myapp> \
	--private-key ${HOME}/.ssh/<SSH key for myapp> \
	--branch prod-<myapp>

  # Get status of application under wego control
  wego app status podinfo

  # Get help for wego app add command
  wego app add -h
  wego help app add

  # Show manifests that would be installed by the wego gitops install command
  wego gitops install --dry-run

  # Install wego in the wego-system namespace
  wego gitops install

  # Get the version of wego along with commit, branch, and flux version
  wego version
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		configureLogger()
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

var ApplicationCmd = &cobra.Command {
	Use:    "app [app Add or status application]",
	Short:  "Add or status application",
	Args:   cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
}

func main() {
	fluxBin.SetupFluxBin()
	rootCmd.PersistentFlags().BoolVarP(&options.verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().String("namespace", "wego-system", "gitops runtime namespace")

	rootCmd.AddCommand(gitops.Cmd)
	rootCmd.AddCommand(version.Cmd)
	rootCmd.AddCommand(flux.Cmd)

	rootCmd.AddCommand(app.ApplicationCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
