package server

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/core/services/remotecluster"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"k8s.io/client-go/rest"
)

func Hydrate(ctx context.Context, mux *runtime.ServeMux, config *rest.Config) error {
	cfgs := map[string]*rest.Config{
		"leaf-cluster-1": {
			Host: "https://172.18.0.3:6443",
			TLSClientConfig: rest.TLSClientConfig{
				CertData: []byte(""),
			},
		},
		"management-cluster": {
			Host: "https://172.18.0.2:6443",
			TLSClientConfig: rest.TLSClientConfig{
				CertData: []byte(""),
			},
		},
	}

	rc := remotecluster.NewConfigGetter(cfgs)

	appsServer := NewAppServer(config, rc)
	if err := pb.RegisterAppsHandlerServer(ctx, mux, appsServer); err != nil {
		return fmt.Errorf("could not register new app server: %w", err)
	}

	return nil
}
