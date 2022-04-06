package server_test

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
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
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func makeGRPCServer(c client.Client, t *testing.T) pb.CoreClient {
	principal := &auth.UserPrincipal{}
	s := grpc.NewServer(
		withClientsPoolInterceptor(principal, c),
	)

	coreCfg, err := server.NewCoreConfig(logr.Discard(), &rest.Config{}, c, "foobar")
	if err != nil {
		t.Fatal(err)
	}

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

func makeGRPCServerWithAPIServer(cfg *rest.Config, t *testing.T) pb.CoreClient {
	g := NewGomegaWithT(t)

	principal := &auth.UserPrincipal{}
	c, err := client.New(cfg, client.Options{})
	g.Expect(err).NotTo(HaveOccurred())

	s := grpc.NewServer(
		withClientsPoolInterceptor(principal, c),
	)

	coreCfg, err := server.NewCoreConfig(logr.Discard(), cfg, c, "foobar")
	if err != nil {
		t.Fatal(err)
	}

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

// Not using a counterfeit: I want the real methods on the provided
// `client.Client` to be invoked.
type clientMock struct {
	client.Client
}

func (c clientMock) RestConfig() *rest.Config {
	return &rest.Config{}
}

func withClientsPoolInterceptor(user *auth.UserPrincipal, client client.Client) grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		clientsPool := clustersmngrfakes.FakeClientsPool{}
		clientsPool.ClientsReturns(map[string]clustersmngr.ClusterClient{"default": clientMock{client}})
		clientsPool.ClientReturns(clientMock{client}, nil)

		clusterClient := clustersmngr.NewClient(&clientsPool)

		ctx = context.WithValue(ctx, clustersmngr.ClustersClientCtxKey, clusterClient)

		return handler(ctx, req)
	})
}
