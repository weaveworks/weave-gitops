package server

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/rest"

	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

func makeGRPCServer(cfg *rest.Config, t *testing.T) (pb.AppsClient, func()) {
	s := grpc.NewServer()

	apps := NewAppServer(cfg)

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

func TestCreateKustomization_App_Association(t *testing.T) {
	ctx := context.Background()

	f := setUpAppServerTest(t)
	defer f.cleanUpFixture(t)

	c, cleanup := makeGRPCServer(f.env.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(f.env.Rest, "")
	if err != nil {
		t.Fatal(err)
	}

	ns := &corev1.Namespace{}
	ns.Name = "kube-test-" + rand.String(5)

	err = k.Create(ctx, ns)
	if err != nil {
		t.Fatal(err)
	}

	r := &pb.AddKustomizationReq{
		Name:      "mykustomization",
		Namespace: ns.Name,
		AppName:   "someapp",
		SourceRef: &pb.SourceRef{
			Kind: pb.SourceRef_GitRepository,
			Name: "othersource",
		},
	}

	res, err := c.AddKustomization(ctx, r)
	if err != nil {
		t.Fatal(err)
	}

	if !res.Success {
		t.Error("expected success")
	}
}
