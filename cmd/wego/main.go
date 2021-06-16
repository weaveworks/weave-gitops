package main

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/cmd/wego/status"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/wego/add"
	"github.com/weaveworks/weave-gitops/cmd/wego/flux"
	"github.com/weaveworks/weave-gitops/cmd/wego/install"
	"github.com/weaveworks/weave-gitops/cmd/wego/version"
	fluxBin "github.com/weaveworks/weave-gitops/pkg/flux"
)

var options struct {
	verbose bool
}

var rootCmd = &cobra.Command{
	Use:   "wego",
	Short: "Weave GitOps",
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

var ApplicationCmd = &cobra.Command{
	Use:   "app [app Add or status application]",
	Short: "Use app commands",
	Args: cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
}

func main() {
	fluxBin.SetupFluxBin()
	rootCmd.PersistentFlags().BoolVarP(&options.verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().String("namespace", "wego-system", "gitops runtime namespace")
	rootCmd.AddCommand(install.Cmd)
	rootCmd.AddCommand(version.Cmd)
	rootCmd.AddCommand(flux.Cmd)

	ApplicationCmd.AddCommand(status.Cmd)
	ApplicationCmd.AddCommand(add.Cmd)

	rootCmd.AddCommand(ApplicationCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	versionInfo, _ := version.CheckVersion()
	if versionInfo != "" {
		log.Info(versionInfo)
	}
}
