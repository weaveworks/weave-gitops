package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"testing"

	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	fakehttp "github.com/weaveworks/weave-gitops/pkg/vendorfakes/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"k8s.io/client-go/rest"
)

var k8sEnv *testutils.K8sTestEnv

var diffIgnoredFields = []string{
	"ObjectMeta.UID",
	"ObjectMeta.SelfLink",
	"ObjectMeta.ResourceVersion",
	"ObjectMeta.ManagedFields",
	"ObjectMeta.CreationTimestamp",
	"ObjectMeta.Generation",
	"TypeMeta",
}

func TestMain(m *testing.M) {
	os.Setenv("KUBEBUILDER_ASSETS", "../../tools/bin/envtest")
	os.Setenv("IS_TEST_ENV", "true")

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

func makeGRPCServer(cfg *rest.Config, h *http.Client, t *testing.T) (pb.AppsClient, func()) {
	s := grpc.NewServer()

	apps := &appServer{
		k8s:  placeholderClientGetter{cfg},
		http: h,
	}

	lis := bufconn.Listen(1024 * 1024)

	pb.RegisterAppsServer(s, apps)

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	go func(tt *testing.T) {
		if err := s.Serve(lis); err != nil {
			tt.Error(err)
		}
	}(t)

	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(dialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		s.GracefulStop()
		conn.Close()
	}

	return pb.NewAppsClient(conn), cleanup
}

func mockHttpClient() (*http.Client, *fakehttp.FakeRoundTripper) {
	rt := &fakehttp.FakeRoundTripper{}
	c := &http.Client{Transport: rt}

	return c, rt
}
