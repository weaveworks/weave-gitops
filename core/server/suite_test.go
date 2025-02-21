package server_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	typedauth "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster/clusterfakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/fetcher"
	"github.com/weaveworks/weave-gitops/core/nsaccess/nsaccessfakes"
	"github.com/weaveworks/weave-gitops/core/server"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/health"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/crd"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

var (
	k8sEnv    *testutils.K8sTestEnv
	nsChecker nsaccessfakes.FakeChecker
)

func TestMain(m *testing.M) {
	var err error
	k8sEnv, err = testutils.StartK8sTestEnvironment([]string{
		"../../tools/testcrds",
	})
	if err != nil {
		panic(err)
	}

	code := m.Run()

	k8sEnv.Stop()

	os.Exit(code)
}

func makeGRPCServer(ctx context.Context, t *testing.T, cfg *rest.Config) pb.CoreClient {
	t.Helper()
	log := logr.Discard()
	nsChecker = nsaccessfakes.FakeChecker{}
	nsChecker.FilterAccessibleNamespacesStub = func(ctx context.Context, t typedauth.AuthorizationV1Interface, n []v1.Namespace) ([]v1.Namespace, error) {
		// Pretend the user has access to everything
		return n, nil
	}

	scheme, err := kube.CreateScheme()
	if err != nil {
		t.Fatalf("Failed to create scheme: %v", err)
	}

	cluster, err := cluster.NewSingleCluster("Default", k8sEnv.Rest, scheme, kube.UserPrefixes{})
	if err != nil {
		t.Fatalf("Failed to create cluster: %v", err)
	}

	fetch := fetcher.NewSingleClusterFetcher(cluster)

	hc := health.NewHealthChecker()

	clustersManager := clustersmngr.NewClustersManager([]clustersmngr.ClusterFetcher{fetch}, &nsChecker, log)
	coreCfg, err := server.NewCoreConfig(log, cfg, "foobar", clustersManager, hc)
	if err != nil {
		t.Fatalf("Failed to create CoreConfig: %v", err)
	}

	coreCfg.NSAccess = &nsChecker
	coreCfg.CRDService = crd.NewNoCacheFetcher(clustersManager)

	core, err := server.NewCoreServer(ctx, coreCfg)
	if err != nil {
		t.Fatalf("Failed to create CoreServer: %v", err)
	}

	lis := bufconn.Listen(1024 * 1024)

	s := grpc.NewServer(
		withClientsPoolInterceptor(clustersManager),
	)

	pb.RegisterCoreServer(s, core)

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	go func(t *testing.T) {
		t.Helper()
		if err := s.Serve(lis); err != nil {
			t.Errorf("Failed to serve: %v", err)
		}
	}(t)

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	t.Cleanup(func() {
		s.GracefulStop()
		conn.Close()
	})

	return pb.NewCoreClient(conn)
}

type userKey struct{}

var UserKey = userKey{}

const (
	MetadataUserKey   string = "test_principal_user"
	MetadataGroupsKey string = "test_principal_groups"
)

func withClientsPoolInterceptor(clustersManager clustersmngr.ClustersManager) grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := clustersManager.UpdateClusters(ctx); err != nil {
			return nil, fmt.Errorf("failed to update clusters: %w", err)
		}
		if err := clustersManager.UpdateNamespaces(ctx); err != nil {
			return nil, fmt.Errorf("failed to update namespaces: %w", err)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, fmt.Errorf("getting metadata from context failed")
		}

		var user string
		if len(md[MetadataUserKey]) > 0 {
			user = md[MetadataUserKey][0]
		}
		groups := md[MetadataGroupsKey]
		principal := auth.UserPrincipal{ID: user, Groups: groups}
		clustersManager.UpdateUserNamespaces(ctx, &principal)

		ctx = auth.WithPrincipal(ctx, &principal)

		return handler(ctx, req)
	})
}

func makeServerConfig(t *testing.T, fakeClient client.Client, clusterName string) server.CoreServerConfig {
	t.Helper()
	log := logr.Discard()
	nsChecker = nsaccessfakes.FakeChecker{}
	nsChecker.FilterAccessibleNamespacesStub = func(ctx context.Context, t typedauth.AuthorizationV1Interface, n []v1.Namespace) ([]v1.Namespace, error) {
		// Pretend the user has access to everything
		return n, nil
	}
	clientset := fake.NewClientset()

	cluster := clusterfakes.FakeCluster{}

	if clusterName == "" {
		clusterName = "Default"
	}

	cluster.GetNameReturns(clusterName)
	cluster.GetUserClientReturns(fakeClient, nil)
	cluster.GetUserClientsetReturns(clientset, nil)
	cluster.GetServerClientReturns(fakeClient, nil)

	fetcher := fetcher.NewSingleClusterFetcher(&cluster)

	// Don't include the clustersmngr.DefaultKubeConfigOptions here as we're using a fake kubeclient
	// and the default options include the Flowcontrol setup which is not mocked out
	clustersManager := clustersmngr.NewClustersManager([]clustersmngr.ClusterFetcher{fetcher}, &nsChecker, log)

	hc := health.NewHealthChecker()

	coreCfg, err := server.NewCoreConfig(log, &rest.Config{}, "foobar", clustersManager, hc)
	if err != nil {
		t.Fatalf("Failed to create CoreServerConfig: %v", err)
	}

	coreCfg.NSAccess = &nsChecker

	return coreCfg
}

func makeServer(ctx context.Context, t *testing.T, cfg server.CoreServerConfig) pb.CoreClient {
	t.Helper()
	core, err := server.NewCoreServer(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create CoreServer: %v", err)
	}

	lis := bufconn.Listen(1024 * 1024)

	s := grpc.NewServer(
		withClientsPoolInterceptor(cfg.ClustersManager),
	)

	pb.RegisterCoreServer(s, core)

	go func(t *testing.T) {
		t.Helper()
		if err := s.Serve(lis); err != nil {
			t.Errorf("Failed to serve: %v", err)
		}
	}(t)

	conn := dialBufnet(t, lis)

	t.Cleanup(func() {
		s.GracefulStop()
		conn.Close()
	})

	return pb.NewCoreClient(conn)
}

func dialBufnet(t *testing.T, lis *bufconn.Listener) *grpc.ClientConn {
	t.Helper()
	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Insecure for testing
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	return conn
}
