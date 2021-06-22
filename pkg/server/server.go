package server

import (
	"context"
	"fmt"

	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

type server struct {
	pb.UnimplementedApplicationsServer

	kube kube.Kube
}

func NewApplicationsServer(kubeSvc kube.Kube) pb.ApplicationsServer {
	return &server{
		kube: kubeSvc,
	}
}

func (s *server) ListApplications(ctx context.Context, msg *pb.ListApplicationsRequest) (*pb.ListApplicationsResponse, error) {

	apps, err := s.kube.GetApplications(msg.GetNamespace())
	if err != nil {
		return nil, err
	}

	if apps == nil {
		return &pb.ListApplicationsResponse{
			Applications: []*pb.Application{},
		}, nil
	}

	list := []*pb.Application{}
	for _, a := range *apps {
		list = append(list, &pb.Application{Name: a.Name})
	}
	return &pb.ListApplicationsResponse{
		Applications: list,
	}, nil
}

func (s *server) GetApplication(ctx context.Context, msg *pb.GetApplicationRequest) (*pb.GetApplicationResponse, error) {
	app, err := s.kube.GetApplication(ctx, msg.ApplicationName)

	if err != nil {
		return nil, fmt.Errorf("could not get application: %s", err)
	}

	return &pb.GetApplicationResponse{Application: &pb.Application{
		Name: app.Name,
		Url:  app.Spec.URL,
		Path: app.Spec.Path,
	}}, nil
}
