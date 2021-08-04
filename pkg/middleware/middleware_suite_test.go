package middleware_test

import (
	"context"
	"fmt"
	"io"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/api/ping"
)

func TestMiddleware(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Middleware Suite")
}

const (
	// DefaultPongValue is the default value used.
	DefaultResponseValue = "default_response_value"
	// ListResponseCount is the expected number of responses to PingList
	ListResponseCount = 100
)

type TestPingService struct {
	ping.UnimplementedTestServiceServer
}

func (s *TestPingService) PingEmpty(ctx context.Context, _ *ping.Empty) (*ping.PingResponse, error) {
	return &ping.PingResponse{Value: DefaultResponseValue, Counter: 42}, nil
}

func (s *TestPingService) Ping(ctx context.Context, req *ping.PingRequest) (*ping.PingResponse, error) {
	// Send user trailers and headers.
	return &ping.PingResponse{Value: req.Value, Counter: 42}, nil
}

func (s *TestPingService) PingError(ctx context.Context, ping *ping.PingRequest) (*ping.Empty, error) {
	return nil, fmt.Errorf("fooo")
}

func (s *TestPingService) PingList(req *ping.PingRequest, stream ping.TestService_PingListServer) error {
	if req.ErrorCodeReturned != 0 {
		return fmt.Errorf("fooo")
	}
	// Send user trailers and headers.
	for i := 0; i < ListResponseCount; i++ {
		stream.Send(&ping.PingResponse{Value: req.Value, Counter: int32(i)})
	}
	return nil
}

func (s *TestPingService) PingStream(stream ping.TestService_PingStreamServer) error {
	count := 0
	for true {
		p, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		stream.Send(&ping.PingResponse{Value: p.Value, Counter: int32(count)})
		count += 1
	}
	return nil
}
