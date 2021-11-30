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

const addr = "0.0.0.0:8000"

func main() {
	var helmRepoNamespace, helmRepoName string

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

			s, err := server.NewProfilesHandler(context.Background(), logr, helmRepoNamespace, helmRepoName)
			if err != nil {
				return err
			}

			logr.Info("server starting", "address", addr)
			return http.ListenAndServe(addr, s)
		},
	}

	cmd.Flags().StringVar(&helmRepoNamespace, "helm-repo-namespace", "default", "the namespace of the Helm Repository resource to scan for profiles")
	cmd.Flags().StringVar(&helmRepoName, "helm-repo-name", "weaveworks-charts", "the name of the Helm Repository resource to scan for profiles")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
