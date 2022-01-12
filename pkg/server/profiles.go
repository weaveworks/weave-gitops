package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"

	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache"

	"sigs.k8s.io/controller-runtime/pkg/client"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/metadata"
)

const (
	OctetStreamType = "application/octet-stream"
	JsonType        = "application/json"
)

type ProfilesConfig struct {
	logr              logr.Logger
	helmRepoNamespace string
	helmRepoName      string
	helmCache         cache.Cache
	kubeClient        client.Client
}

func NewProfilesConfig(kubeClient client.Client, helmCache cache.Cache, helmRepoNamespace, helmRepoName string) ProfilesConfig {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("could not create zap logger: %v", err)
	}

	return ProfilesConfig{
		logr:              zapr.NewLogger(zapLog),
		helmRepoNamespace: helmRepoNamespace,
		helmRepoName:      helmRepoName,
		kubeClient:        kubeClient,
		helmCache:         helmCache,
	}
}

type ProfilesServer struct {
	pb.UnimplementedProfilesServer

	KubeClient        client.Client
	Log               logr.Logger
	HelmRepoName      string
	HelmRepoNamespace string
	HelmCache         cache.Cache
}

func NewProfilesServer(config ProfilesConfig) pb.ProfilesServer {
	return &ProfilesServer{
		Log:               config.logr,
		HelmRepoNamespace: config.helmRepoNamespace,
		HelmRepoName:      config.helmRepoName,
		HelmCache:         config.helmCache,
		KubeClient:        config.kubeClient,
	}
}

func (s *ProfilesServer) GetProfiles(ctx context.Context, msg *pb.GetProfilesRequest) (*pb.GetProfilesResponse, error) {
	helmRepo := &sourcev1beta1.HelmRepository{}
	err := s.KubeClient.Get(ctx, client.ObjectKey{
		Name:      s.HelmRepoName,
		Namespace: s.HelmRepoNamespace,
	}, helmRepo)

	if err != nil {
		if apierrors.IsNotFound(err) {
			errMsg := fmt.Sprintf("HelmRepository %q/%q does not exist", s.HelmRepoNamespace, s.HelmRepoName)
			s.Log.Error(err, errMsg)

			return &pb.GetProfilesResponse{
					Profiles: []*pb.Profile{},
				}, &grpcruntime.HTTPStatusError{
					Err:        errors.New(errMsg),
					HTTPStatus: http.StatusOK,
				}
		}

		return nil, fmt.Errorf("failed to get HelmRepository %q/%q", s.HelmRepoNamespace, s.HelmRepoName)
	}

	ps, err := s.HelmCache.GetProfiles(helmRepo.Namespace, helmRepo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to scan HelmRepository %q/%q for charts: %w", s.HelmRepoNamespace, s.HelmRepoName, err)
	}

	return &pb.GetProfilesResponse{
		Profiles: ps,
	}, nil
}

func (s *ProfilesServer) GetProfileValues(ctx context.Context, msg *pb.GetProfileValuesRequest) (*httpbody.HttpBody, error) {
	helmRepo := &sourcev1beta1.HelmRepository{}
	err := s.KubeClient.Get(ctx, client.ObjectKey{
		Name:      s.HelmRepoName,
		Namespace: s.HelmRepoNamespace,
	}, helmRepo)

	if err != nil {
		if apierrors.IsNotFound(err) {
			errMsg := fmt.Sprintf("HelmRepository %q/%q does not exist", s.HelmRepoNamespace, s.HelmRepoName)
			s.Log.Error(err, errMsg)

			return &httpbody.HttpBody{
					ContentType: "application/json",
					Data:        []byte{},
				}, &grpcruntime.HTTPStatusError{
					Err:        errors.New(errMsg),
					HTTPStatus: http.StatusOK,
				}
		}

		return nil, fmt.Errorf("failed to get HelmRepository %q/%q", s.HelmRepoNamespace, s.HelmRepoName)
	}

	data, err := s.HelmCache.GetProfileValues(helmRepo.Namespace, helmRepo.Name, msg.ProfileName, msg.ProfileVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve values file from Helm chart '%s' (%s): %w", msg.ProfileName, msg.ProfileVersion, err)
	}

	var acceptHeader string

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if accept, ok := md["accept"]; ok {
			acceptHeader = strings.Join(accept, ",")
		}
	}

	if strings.Contains(acceptHeader, OctetStreamType) {
		return &httpbody.HttpBody{
			ContentType: OctetStreamType,
			Data:        data,
		}, nil
	}

	res, err := json.Marshal(&pb.GetProfileValuesResponse{
		Values: base64.StdEncoding.EncodeToString(data),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response to JSON: %w", err)
	}

	return &httpbody.HttpBody{
		ContentType: JsonType,
		Data:        res,
	}, nil
}
