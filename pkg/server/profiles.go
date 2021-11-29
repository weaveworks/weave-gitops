package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/go-logr/logr"
	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/metadata"
)

const (
	OctetStreamType = "application/octet-stream"
	JsonType        = "application/json"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . HelmRepoManager
type HelmRepoManager interface {
	GetCharts(ctx context.Context, hr *sourcev1beta1.HelmRepository, pred helm.ChartPredicate) ([]*pb.Profile, error)
	GetValuesFile(ctx context.Context, helmRepo *sourcev1beta1.HelmRepository, c *helm.ChartReference, filename string) ([]byte, error)
}

func NewProfilesHandler(ctx context.Context, logr logr.Logger) (http.Handler, error) {
	rest, clusterName, err := kube.RestConfig()
	if err != nil {
		return nil, fmt.Errorf("could not create client config: %w", err)
	}

	_, rawClient, err := kube.NewKubeHTTPClientWithConfig(rest, clusterName)
	if err != nil {
		return nil, fmt.Errorf("could not create kube http client: %w", err)
	}

	tempDir, err := ioutil.TempDir("", "helmrepocache")
	if err != nil {
		return nil, err
	}

	helmRepoNs := os.Getenv("RUNTIME_NAMESPACE")
	profilesSrv := &ProfilesServer{
		KubeClient:        rawClient,
		Log:               logr,
		HelmChartManager:  helm.NewRepoManager(rawClient, helmRepoNs, tempDir),
		HelmRepoNamespace: helmRepoNs,
		//TODO make this configurable
		HelmRepoName: "weaveworks-charts",
		cacheDir:     tempDir,
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

	KubeClient        client.Client
	Log               logr.Logger
	HelmChartManager  HelmRepoManager
	HelmRepoName      string
	HelmRepoNamespace string
	cacheDir          string
}

func (s *ProfilesServer) GetProfiles(ctx context.Context, msg *pb.GetProfilesRequest) (*pb.GetProfilesResponse, error) {
	// Look for helm repository object in the current namespace
	helmRepo := &sourcev1beta1.HelmRepository{}
	err := s.KubeClient.Get(ctx, client.ObjectKey{
		Name:      s.HelmRepoName,
		Namespace: s.HelmRepoNamespace,
	}, helmRepo)

	if err != nil {
		errMsg := fmt.Sprintf("cannot find HelmRepository %q/%q", s.HelmRepoNamespace, s.HelmRepoName)
		s.Log.Error(err, errMsg)

		return &pb.GetProfilesResponse{
				Profiles: []*pb.Profile{},
			}, &grpcruntime.HTTPStatusError{
				Err: errors.New(errMsg),
				//TODO: why do we return 200?
				HTTPStatus: http.StatusOK,
			}
	}

	ps, err := s.HelmChartManager.GetCharts(ctx, helmRepo, helm.Profiles)
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
		errMsg := fmt.Sprintf("cannot find HelmRepository %q/%q", s.HelmRepoNamespace, s.HelmRepoName)
		s.Log.Error(err, errMsg)

		return &httpbody.HttpBody{
				ContentType: "application/json",
				Data:        []byte{},
			}, &grpcruntime.HTTPStatusError{
				Err: errors.New(errMsg),
				//TODO: why do we return 200?
				HTTPStatus: http.StatusOK,
			}
	}

	sourceRef := helmv2beta1.CrossNamespaceObjectReference{
		APIVersion: helmRepo.TypeMeta.APIVersion,
		Kind:       helmRepo.TypeMeta.Kind,
		Name:       helmRepo.ObjectMeta.Name,
		Namespace:  helmRepo.ObjectMeta.Namespace,
	}

	ref := &helm.ChartReference{Chart: msg.ProfileName, Version: msg.ProfileVersion, SourceRef: sourceRef}
	valuesBytes, err := s.HelmChartManager.GetValuesFile(ctx, helmRepo, ref, chartutil.ValuesfileName)

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
			Data:        valuesBytes,
		}, nil
	}

	res, err := json.Marshal(&pb.GetProfileValuesResponse{
		Values: base64.StdEncoding.EncodeToString(valuesBytes),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response to JSON: %w", err)
	}

	return &httpbody.HttpBody{
		ContentType: JsonType,
		Data:        res,
	}, nil
}
