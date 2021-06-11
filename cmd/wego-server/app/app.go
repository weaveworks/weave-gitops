package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/server"
)

func NewAPIServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "capi-server",
		Long: `The capi-server servers and handles REST operations for CAPI templates.
		CAPI templates are stored in the cluster as a ConfigMap indexed by their name.`,

		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return StartServer()
		},
	}
	return cmd
}

var addr = "0.0.0.0:8000"

func StartServer() error {
	log.Infof("wego api server started. listen address: %s", addr)
	return RunInProcessGateway(context.Background(), addr)
}

// RunInProcessGateway starts the invoke in process http gateway.
func RunInProcessGateway(ctx context.Context, addr string, opts ...runtime.ServeMuxOption) error {
	mux := runtime.NewServeMux(opts...)

	if err := pb.RegisterApplicationsHandlerServer(ctx, mux, server.NewApplicationsServer()); err != nil {
		return fmt.Errorf("could not register application: %w", err)
	}
	s := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		log.Infof("Shutting down the http gateway server")
		if err := s.Shutdown(context.Background()); err != nil {
			log.Errorf("Failed to shutdown http gateway server: %v", err)
		}
	}()

	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Errorf("Failed to listen and serve: %v", err)
		return err
	}
	return nil
}
