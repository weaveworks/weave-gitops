package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"google.golang.org/genproto/googleapis/api/httpbody"
)

func NewProfilesHandler(ctx context.Context, logr logr.Logger) (http.Handler, error) {
	rest, clusterName, err := kube.RestConfig()
	if err != nil {
		return nil, fmt.Errorf("could not create client config: %w", err)
	}

	_, rawClient, err := kube.NewKubeHTTPClientWithConfig(rest, clusterName)
	if err != nil {
		return nil, fmt.Errorf("could not create kube http client: %w", err)
	}

	profilesSrv := &ProfilesServer{
		kube: rawClient,
		log:  logr,
	}

	mux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(logr))
	httpHandler := middleware.WithLogging(logr, mux)

	if err := pb.RegisterProfilesHandlerServer(ctx, mux, profilesSrv); err != nil {
		return nil, fmt.Errorf("could not register Profiles server: %w", err)
	}

	return httpHandler, nil
}

type ProfilesServer struct {
	pb.UnimplementedProfilesServer

	kube client.Client
	log  logr.Logger
}

func (s *ProfilesServer) GetProfiles(ctx context.Context, msg *pb.GetProfilesRequest) (*pb.GetProfilesResponse, error) {
	return nil, nil
}

func (s *ProfilesServer) GetProfileValues(ctx context.Context, msg *pb.GetProfileValuesRequest) (*httpbody.HttpBody, error) {
	return nil, nil
}
