package server_test

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
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

func makeGRPCServer(cfg *rest.Config, t *testing.T) pb.CoreClient {
	principal := &auth.UserPrincipal{}
	s := grpc.NewServer(
		withClientsPoolInterceptor(cfg, principal),
	)

	coreCfg := server.NewCoreConfig(logr.Discard(), cfg, "foobar")
	nsChecker = nsaccessfakes.FakeChecker{}
	nsChecker.FilterAccessibleNamespacesStub = func(ctx context.Context, c *rest.Config, n []v1.Namespace) ([]v1.Namespace, error) {
		// Pretend the user has access to everything
		return n, nil
	}
	coreCfg.NSAccess = &nsChecker

	core := server.NewCoreServer(coreCfg)

	lis := bufconn.Listen(1024 * 1024)

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

	return pb.NewCoreClient(conn)
}

func withClientsPoolInterceptor(config *rest.Config, user *auth.UserPrincipal) grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		cluster := clustersmngr.Cluster{
			Name:        "Default",
			Server:      config.Host,
			BearerToken: config.BearerToken,
			TLSConfig:   config.TLSClientConfig,
		}

		clientsPool := clustersmngr.NewClustersClientsPool()
		if err := clientsPool.Add(user, cluster); err != nil {
			return nil, err
		}

		clusterClient := clustersmngr.NewClient(clientsPool)

		ctx = context.WithValue(ctx, clustersmngr.ClustersClientCtxKey, clusterClient)

		return handler(ctx, req)
	})
}
