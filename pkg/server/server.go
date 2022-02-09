package server

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"go.uber.org/zap"
)

const DefaultPort = "9001"

// An ApplicationsConfig allows for the customization of an ApplicationsServer.
// Use the DefaultConfig() to use the default dependencies.
type ApplicationsConfig struct {
	Logger        logr.Logger
	ClusterConfig kube.ClusterConfig
}

// DefaultApplicationsConfig creates a populated config with the dependencies for a Server
func DefaultApplicationsConfig() (*ApplicationsConfig, error) {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("could not create zap logger: %v", err)
	}

	logr := zapr.NewLogger(zapLog)

	rand.Seed(time.Now().UnixNano())

	rest, clusterName, err := kube.RestConfig()
	if err != nil {
		return nil, fmt.Errorf("could not create client config: %w", err)
	}

	return &ApplicationsConfig{
		Logger: logr,
		ClusterConfig: kube.ClusterConfig{
			DefaultConfig: rest,
			ClusterName:   clusterName,
		},
	}, nil
}
