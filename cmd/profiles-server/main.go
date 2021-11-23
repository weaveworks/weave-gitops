package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/server"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"github.com/spf13/cobra"
)

func main() {
	if err := NewAPIServerCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

const addr = "0.0.0.0:8000"

func NewAPIServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "profiles-server",
		Long: `The profiles-server handles HTTP requests for the Profiles API`,

		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			zapLog, err := zap.NewDevelopment()
			if err != nil {
				log.Fatalf("could not create zap logger: %v", err)
			}
			logr := zapr.NewLogger(zapLog)

			s, err := server.NewProfilesHandler(context.Background(), logr)
			if err != nil {
				return err
			}

			logr.Info("server starting", "address", addr)
			return http.ListenAndServe(addr, s)
		},
	}

	return cmd
}
