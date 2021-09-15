package main

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
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
			cfg, err := server.DefaultConfig()
			if err != nil {
				return err
			}

			s, err := server.NewApplicationsHandler(context.Background(), cfg)
			if err != nil {
				return err
			}

			cfg.Logger.Info("server staring", "address", addr)
			return http.ListenAndServe(addr, s)
		},
	}

	return cmd
}
