package server_test

import (
	"context"
	"log"
	"net"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Server")

}

const bufSize = 1024 * 1024

var lis *bufconn.Listener

var s *grpc.Server
var apps pb.ApplicationsServer
var client pb.ApplicationsClient
var conn *grpc.ClientConn
var err error

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

var _ = BeforeEach(func() {
	lis = bufconn.Listen(bufSize)
	s = grpc.NewServer()
	apps = server.NewApplicationsServer()
	pb.RegisterApplicationsServer(s, apps)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf(err.Error())
		}
	}()

	ctx := context.Background()
	conn, err = grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())

	Expect(err).NotTo(HaveOccurred())

	client = pb.NewApplicationsClient(conn)
})

var _ = AfterEach(func() {
	conn.Close()
})
