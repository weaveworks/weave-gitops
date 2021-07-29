package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/go-logr/zapr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/middleware"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"go.uber.org/zap"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	command := NewAPIServerCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}

func NewAPIServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "wego-server",
		Long: `The wego-server handles HTTP requests for Weave GitOps Applications`,

		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return StartServer()
		},
	}
	return cmd
}

var addr = "0.0.0.0:8000"

func StartServer() error {
	ctx := context.Background()

	return RunInProcessGateway(ctx, addr)
}

// RunInProcessGateway starts the invoke in process http gateway.
func RunInProcessGateway(ctx context.Context, addr string) error {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("could not create zap logger: %v", err)
	}
	log := zapr.NewLogger(zapLog)

	kubeClient, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("could not create kube http client: %w", err)
	}

	appsSrv := server.NewApplicationsServer(kubeClient)

	mux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(log))
	httpHandler := middleware.WithLogging(log, mux)

	if err := pb.RegisterApplicationsHandlerServer(ctx, mux, appsSrv); err != nil {
		return fmt.Errorf("could not register application: %w", err)
	}

	s := &http.Server{
		Addr:    addr,
		Handler: httpHandler,
	}

	go func() {
		<-ctx.Done()
		log.Info("Shutting down the http gateway server")
		if err := s.Shutdown(context.Background()); err != nil {
			log.Error(err, "failed to shutdown http gateway server")
		}
	}()

	log.Info("wego api server started", "address", addr)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Error(err, "failed to listen and serve")
		return err
	}
	return nil
}
