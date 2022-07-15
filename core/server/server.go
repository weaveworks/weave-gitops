package server

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"k8s.io/client-go/rest"
)

func Hydrate(ctx context.Context, mux *runtime.ServeMux, cfg CoreServerConfig) error {
	appsServer, err := NewCoreServer(cfg)
	if err != nil {
		return fmt.Errorf("unable to create new kube client: %w", err)
	}

	if err = pb.RegisterCoreHandlerServer(ctx, mux, appsServer); err != nil {
		return fmt.Errorf("could not register new app server: %w", err)
	}

	return nil
}

const temporarilyEmptyAppName = ""

type coreServer struct {
	pb.UnimplementedCoreServer

	logger         logr.Logger
	nsChecker      nsaccess.Checker
	clientsFactory clustersmngr.ClientsFactory
	primaryKinds   *PrimaryKinds
}

type CoreServerConfig struct {
	log            logr.Logger
	RestCfg        *rest.Config
	clusterName    string
	NSAccess       nsaccess.Checker
	ClientsFactory clustersmngr.ClientsFactory
	PrimaryKinds   *PrimaryKinds
}

func NewCoreConfig(log logr.Logger, cfg *rest.Config, clusterName string, clusterClientFactory clustersmngr.ClientsFactory) CoreServerConfig {
	return CoreServerConfig{
		log:            log.WithName("core-server"),
		RestCfg:        cfg,
		clusterName:    clusterName,
		NSAccess:       nsaccess.NewChecker(nsaccess.DefautltWegoAppRules),
		ClientsFactory: clusterClientFactory,
		PrimaryKinds:   DefaultPrimaryKinds(),
	}
}

func NewCoreServer(cfg CoreServerConfig) (pb.CoreServer, error) {
	return &coreServer{
		logger:         cfg.log,
		nsChecker:      cfg.NSAccess,
		clientsFactory: cfg.ClientsFactory,
		primaryKinds:   cfg.PrimaryKinds,
	}, nil
}
