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
	"github.com/weaveworks/weave-gitops/pkg/charts"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/metadata"
)

const (
	octetStreamType = "application/octet-stream"
	jsonType        = "application/octet-stream"
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

	tempDir, err := ioutil.TempDir("", "helmrepocache")
	if err != nil {
		return nil, err
	}

	helmRepoNs := os.Getenv("RUNTIME_NAMESPACE")
	chartClient := charts.NewHelmChartClient(rawClient, helmRepoNs, &sourcev1beta1.HelmRepository{}, charts.WithCacheDir(tempDir))

	profilesSrv := &ProfilesServer{
		kubeClient:        rawClient,
		log:               logr,
		chartScanner:      &charts.Scanner{},
		chartClient:       chartClient,
		helmRepoNamespace: helmRepoNs,
		//TODO make this configurable
		helmRepoName:           "weaveworks-charts",
		helmRepositoryCacheDir: tempDir,
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

	kubeClient             client.Client
	log                    logr.Logger
	chartScanner           charts.ChartScanner
	chartClient            charts.ChartClient
	helmRepoName           string
	helmRepoNamespace      string
	helmRepositoryCacheDir string
}

func (s *ProfilesServer) GetProfiles(ctx context.Context, msg *pb.GetProfilesRequest) (*pb.GetProfilesResponse, error) {
	// Look for helm repository object in the current namespace
	helmRepo := &sourcev1beta1.HelmRepository{}
	err := s.kubeClient.Get(ctx, client.ObjectKey{
		Name:      s.helmRepoName,
		Namespace: s.helmRepoNamespace,
	}, helmRepo)

	if err != nil {
		errMsg := fmt.Sprintf("cannot find HelmRepository %q/%q", s.helmRepoNamespace, s.helmRepoName)
		s.log.Error(err, errMsg)

		return &pb.GetProfilesResponse{
				Profiles: []*pb.Profile{},
			}, &grpcruntime.HTTPStatusError{
				Err: errors.New(errMsg),
				//TODO: why do we return 200?
				HTTPStatus: http.StatusOK,
			}
	}

	ps, err := s.chartScanner.ScanCharts(ctx, helmRepo, charts.Profiles)
	if err != nil {
		return nil, fmt.Errorf("failed to scan HelmRepository %q/%q for charts: %w", s.helmRepoNamespace, s.helmRepoName, err)
	}

	return &pb.GetProfilesResponse{
		Profiles: ps,
	}, nil
}

func (s *ProfilesServer) GetProfileValues(ctx context.Context, msg *pb.GetProfileValuesRequest) (*httpbody.HttpBody, error) {
	helmRepo := &sourcev1beta1.HelmRepository{}
	err := s.kubeClient.Get(ctx, client.ObjectKey{
		Name:      s.helmRepoName,
		Namespace: s.helmRepoNamespace,
	}, helmRepo)

	if err != nil {
		errMsg := fmt.Sprintf("cannot find HelmRepository %q/%q", s.helmRepoNamespace, s.helmRepoName)
		s.log.Error(err, errMsg)

		return &httpbody.HttpBody{
				ContentType: "application/json",
				Data:        []byte{},
			}, &grpcruntime.HTTPStatusError{
				Err: errors.New(errMsg),
				//TODO: why do we return 200?
				HTTPStatus: http.StatusOK,
			}
	}

	s.chartClient.SetRepository(helmRepo)
	if err := s.chartClient.UpdateCache(ctx); err != nil {
		return nil, fmt.Errorf("failed to update Helm cache: %w", err)
	}

	sourceRef := helmv2beta1.CrossNamespaceObjectReference{
		APIVersion: helmRepo.TypeMeta.APIVersion,
		Kind:       helmRepo.TypeMeta.Kind,
		Name:       helmRepo.ObjectMeta.Name,
		Namespace:  helmRepo.ObjectMeta.Namespace,
	}
	ref := &charts.ChartReference{Chart: msg.ProfileName, Version: msg.ProfileVersion, SourceRef: sourceRef}
	valuesBytes, err := s.chartClient.FileFromChart(ctx, ref, chartutil.ValuesfileName)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve values file from Helm chart '%s' (%s): %w", msg.ProfileName, msg.ProfileVersion, err)
	}

	var acceptHeader string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if accept, ok := md["accept"]; ok {
			acceptHeader = strings.Join(accept, ",")
		}
	}

	if strings.Contains(acceptHeader, octetStreamType) {
		return &httpbody.HttpBody{
			ContentType: octetStreamType,
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
		ContentType: jsonType,
		Data:        res,
	}, nil
}
