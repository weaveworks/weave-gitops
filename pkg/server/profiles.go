package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/chartutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/helm"
	"sigs.k8s.io/controller-runtime/pkg/client"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

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

type ProfilesConfig struct {
	logr              logr.Logger
	helmRepoNamespace string
	helmRepoName      string
	kubeClient        client.Client
}

func NewProfilesConfig(kubeClient client.Client, helmRepoNamespace, helmRepoName string) ProfilesConfig {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("could not create zap logger: %v", err)
	}

	return ProfilesConfig{
		logr:              zapr.NewLogger(zapLog),
		helmRepoNamespace: helmRepoNamespace,
		helmRepoName:      helmRepoName,
		kubeClient:        kubeClient,
	}
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

func NewProfilesServer(config ProfilesConfig) (pb.ProfilesServer, error) {
	tempDir, err := ioutil.TempDir("", "helmrepocache")
	if err != nil {
		return nil, err
	}

	return &ProfilesServer{
		KubeClient:        config.kubeClient,
		Log:               config.logr,
		HelmChartManager:  helm.NewRepoManager(config.kubeClient, tempDir),
		HelmRepoNamespace: config.helmRepoNamespace,
		HelmRepoName:      config.helmRepoName,
		cacheDir:          tempDir,
	}, nil
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

		return nil, fmt.Errorf("failed to get HelmRepository %q/%q: %w", s.HelmRepoNamespace, s.HelmRepoName, err)
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
