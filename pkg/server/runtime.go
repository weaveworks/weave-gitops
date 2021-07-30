package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

var addr = "0.0.0.0:8000"

func StartServer() error {
	ctx := context.Background()

	s, err := NewAppsHTTPHandler(ctx)
	if err != nil {
		return err
	}

	log.Infof("wego api server starting on %s", addr)
	if err := http.ListenAndServe(addr, s); err != http.ErrServerClosed {
		log.Errorf("Failed to listen and serve: %v", err)
		return err
	}
	return nil
}

func NewAppsHTTPHandler(ctx context.Context, opts ...runtime.ServeMuxOption) (http.Handler, error) {
	mux := runtime.NewServeMux(opts...)

	kubeClient, err := kube.NewKubeHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("could not create kube http client: %w", err)
	}

	if err := pb.RegisterApplicationsHandlerServer(ctx, mux, NewApplicationsServer(kubeClient)); err != nil {
		return nil, fmt.Errorf("could not register application: %w", err)
	}

	return mux, nil
}
