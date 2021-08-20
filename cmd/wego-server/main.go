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
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	appspb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	commitspb "github.com/weaveworks/weave-gitops/pkg/api/commits"
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

const addr = "0.0.0.0:8000"

func NewAPIServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "wego-server",
		Long: `The wego-server handles HTTP requests for Weave GitOps Applications`,

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
	commitSrv := server.NewCommitsServer(kubeClient)

	mux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(log))
	httpHandler := middleware.WithLogging(log, mux)

	if err := appspb.RegisterApplicationsHandlerServer(ctx, mux, appsSrv); err != nil {
		return fmt.Errorf("could not register application: %w", err)
	}
	if err := commitspb.RegisterCommitsHandlerServer(ctx, mux, commitSrv); err != nil {
		return fmt.Errorf("could not register commits: %w", err)
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
