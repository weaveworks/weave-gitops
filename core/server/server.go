package server

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/logger"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/telemetry"
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

	logger          logr.Logger
	nsChecker       nsaccess.Checker
	clustersManager clustersmngr.ClustersManager
	primaryKinds    *PrimaryKinds
}

type CoreServerConfig struct {
	log             logr.Logger
	RestCfg         *rest.Config
	clusterName     string
	NSAccess        nsaccess.Checker
	ClustersManager clustersmngr.ClustersManager
	PrimaryKinds    *PrimaryKinds
}

func NewCoreConfig(log logr.Logger, cfg *rest.Config, clusterName string, clustersManager clustersmngr.ClustersManager) CoreServerConfig {
	return CoreServerConfig{
		log:             log.WithName("core-server"),
		RestCfg:         cfg,
		clusterName:     clusterName,
		NSAccess:        nsaccess.NewChecker(nsaccess.DefautltWegoAppRules),
		ClustersManager: clustersManager,
		PrimaryKinds:    DefaultPrimaryKinds(),
	}
}

func NewCoreServer(cfg CoreServerConfig) (pb.CoreServer, error) {
	err := telemetry.InitTelemetry(cfg.ClustersManager)
	if err != nil {
		// If there's an error turning on telemetry, that's not a
		// thing that should interrupt anything else
		cfg.log.V(logger.LogLevelDebug).Info("Couldn't enable telemetry", "error", err)
	}

	return &coreServer{
		logger:          cfg.log,
		nsChecker:       cfg.NSAccess,
		clustersManager: cfg.ClustersManager,
		primaryKinds:    cfg.PrimaryKinds,
	}, nil
}
