package server_test

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/cache"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/core/nsaccess/nsaccessfakes"
	"github.com/weaveworks/weave-gitops/core/server"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

var k8sEnv *testutils.K8sTestEnv
var nsChecker nsaccessfakes.FakeChecker

func TestMain(m *testing.M) {
	os.Setenv("KUBEBUILDER_ASSETS", "../../tools/bin/envtest")

	var err error
	k8sEnv, err = testutils.StartK8sTestEnvironment([]string{
		"../../manifests/crds",
		"../../tools/testcrds",
	})

	if err != nil {
		panic(err)
	}

	code := m.Run()

	k8sEnv.Stop()

	os.Exit(code)
}

func makeGRPCServer(cfg *rest.Config, t *testing.T) (pb.CoreClient, server.CoreServerConfig) {
	log := logr.Discard()
	cacheContainer := cache.NewContainer(
		log,
		cache.WithSimpleCaches(
			cache.WithNamespaceCache(cfg),
		),
	)

	nsChecker = nsaccessfakes.FakeChecker{}
	nsChecker.FilterAccessibleNamespacesStub = func(ctx context.Context, c *rest.Config, n []v1.Namespace) ([]v1.Namespace, error) {
		// Pretend the user has access to everything
		return n, nil
	}

	fetcher := &clustersmngrfakes.FakeClusterFetcher{}
	fetcher.FetchReturns([]clustersmngr.Cluster{restConfigToCluster(k8sEnv.Rest)}, nil)

	clientsFactory := clustersmngr.NewClientFactory(fetcher, &nsChecker, log)

	coreCfg := server.NewCoreConfig(log, cfg, cacheContainer, "foobar", clientsFactory)
	coreCfg.NSAccess = &nsChecker

	core, err := server.NewCoreServer(coreCfg)
	if err != nil {
		t.Fatal(err)
	}

	lis := bufconn.Listen(1024 * 1024)

	principal := &auth.UserPrincipal{}
	s := grpc.NewServer(
		withClientsPoolInterceptor(clientsFactory, cfg, principal),
	)

	pb.RegisterCoreServer(s, core)

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	go func(tt *testing.T) {
		if err := s.Serve(lis); err != nil {
			tt.Error(err)
		}
	}(t)

	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		s.GracefulStop()
		conn.Close()
	})

	return pb.NewCoreClient(conn), coreCfg
}

func withClientsPoolInterceptor(clientsFactory clustersmngr.ClientsFactory, config *rest.Config, user *auth.UserPrincipal) grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := clientsFactory.UpdateClusters(ctx); err != nil {
			return nil, err
		}
		if err := clientsFactory.UpdateNamespaces(ctx); err != nil {
			return nil, err
		}

		clusterClient, err := clientsFactory.GetImpersonatedClient(ctx, user)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, clustersmngr.ClustersClientCtxKey, clusterClient)

		return handler(ctx, req)
	})
}

func updateNamespaceCache(cfg server.CoreServerConfig) {
	_ = cfg.CacheContainer.ForceRefresh(cache.NamespaceStorage)
}

func restConfigToCluster(cfg *rest.Config) clustersmngr.Cluster {
	return clustersmngr.Cluster{
		Name:        "Default",
		Server:      cfg.Host,
		BearerToken: cfg.BearerToken,
		TLSConfig:   cfg.TLSClientConfig,
	}
}
