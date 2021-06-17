package server

import (
	"context"

	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
)

type server struct {
	pb.UnimplementedApplicationsServer
}

func NewApplicationsServer() pb.ApplicationsServer {
	return &server{}
}

func (s *server) ListApplications(ctx context.Context, msg *pb.ListApplicationsRequest) (*pb.ListApplicationsResponse, error) {

	fakeApps := []string{"cool-app", "slick-app", "neat-app"}

	list := []*pb.Application{}

	for _, a := range fakeApps {
		list = append(list, &pb.Application{Name: a})
	}
	return &pb.ListApplicationsResponse{
		Applications: list,
	}, nil
}
