package main

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/server"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	command := NewAPIServerCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}

const addr = "0.0.0.0:8000"

func NewAPIServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "gitops-server",
		Long: `The gitops-server handles HTTP requests for Weave GitOps Applications`,

		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flux.New(osys.New(), &runner.CLIRunner{}).SetupBin()

			appConfig, err := server.DefaultApplicationsConfig()
			if err != nil {
				return err
			}

			//TODO how is this file used?
			profilesConfig := server.NewProfilesConfig("default", "weaveworks-charts")

			s, err := server.NewApplicationsAndProfilesHandler(context.Background(), &server.Config{AppConfig: appConfig, ProfilesConfig: profilesConfig})
			if err != nil {
				return err
			}

			appConfig.Logger.Info("server starting", "address", addr)
			return http.ListenAndServe(addr, s)
		},
	}

	return cmd
}
