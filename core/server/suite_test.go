package server_test

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/weaveworks/weave-gitops/core/multicluster"
	"github.com/weaveworks/weave-gitops/core/server"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"k8s.io/client-go/rest"
)

var k8sEnv *testutils.K8sTestEnv

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
	s := grpc.NewServer(withClientsPoolInterceptor(cfg, &auth.UserPrincipal{}))

	coreCfg := server.NewCoreConfig(cfg, "foobar")
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
		cluster := multicluster.Cluster{
			Name:        "Default",
			Server:      config.Host,
			BearerToken: config.BearerToken,
			TLSConfig:   config.TLSClientConfig,
		}

		clientsPool := multicluster.NewClustersClientsPool()
		if err := clientsPool.Add(user, cluster); err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, multicluster.ClustersClientsPoolCtxKey, clientsPool)

		return handler(ctx, req)
	})
}
