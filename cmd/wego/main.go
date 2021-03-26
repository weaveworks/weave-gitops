package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/weaveworks/go-checkpoint"
	"github.com/weaveworks/weave-gitops/cmd/wego/version"
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

func main() {
	rootCmd.PersistentFlags().BoolVarP(&options.verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.AddCommand(version.Cmd)

	if checkResponse, err := checkpoint.Check(&checkpoint.CheckParams{
		Product: "weave-gitops",
		Version: version.Version,
	}); err == nil && checkResponse.Outdated {
		log.Infof("wego version %s is available; please update at %s",
			checkResponse.CurrentVersion, checkResponse.CurrentDownloadURL)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
