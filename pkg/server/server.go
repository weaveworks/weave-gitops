package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	appspb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	commitpb "github.com/weaveworks/weave-gitops/pkg/api/commits"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/middleware"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type applicationServer struct {
	appspb.UnimplementedApplicationsServer

	kube kube.Kube
	log  logr.Logger
}

type commitServer struct {
	commitpb.UnimplementedCommitsServer

	kube kube.Kube
	log  logr.Logger
}

// An ServerConfig allows for the customization of an ApplicationsServer.
// Use the DefaultConfig() to use the default dependencies.
type ServerConfig struct {
	Logger     logr.Logger
	KubeClient kube.Kube
}

// NewApplicationsServer creates a grpc Applications server
func NewApplicationsServer(cfg *ServerConfig) appspb.ApplicationsServer {
	return &applicationServer{
		kube: cfg.KubeClient,
		log:  cfg.Logger,
	}
}

// NewCommitsServer creates a grpc Commits server
func NewCommitsServer(cfg *ServerConfig) commitpb.CommitsServer {
	return &commitServer{
		kube: cfg.KubeClient,
		log:  cfg.Logger,
	}
}

// DefaultConfig creates a populated config with the dependencies for a Server
func DefaultConfig() (*ServerConfig, error) {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("could not create zap logger: %v", err)
	}
	logr := zapr.NewLogger(zapLog)

	kubeClient, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("could not create kube http client: %w", err)
	}

	return &ServerConfig{
		Logger:     logr,
		KubeClient: kubeClient,
	}, nil
}

// NewServerHandler allow for other applications to embed the Weave GitOps HTTP API.
// This handler can be muxed with other services or used as a standalone service.
func NewServerHandler(ctx context.Context, cfg *ServerConfig, opts ...runtime.ServeMuxOption) (http.Handler, error) {
	appsSrv := NewApplicationsServer(cfg)
	commitSrv := NewCommitsServer(cfg)

	mux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(cfg.Logger))
	httpHandler := middleware.WithLogging(cfg.Logger, mux)

	if err := appspb.RegisterApplicationsHandlerServer(ctx, mux, appsSrv); err != nil {
		return nil, fmt.Errorf("could not register application: %w", err)
	}

	if err := commitpb.RegisterCommitsHandlerServer(ctx, mux, commitSrv); err != nil {
		return nil, fmt.Errorf("could not register commit: %w", err)
	}

	return httpHandler, nil
}

func (s *applicationServer) ListApplications(ctx context.Context, msg *appspb.ListApplicationsRequest) (*appspb.ListApplicationsResponse, error) {
	apps, err := s.kube.GetApplications(ctx, msg.GetNamespace())
	if err != nil {
		return nil, err
	}

	if apps == nil {
		return &appspb.ListApplicationsResponse{
			Applications: []*appspb.Application{},
		}, nil
	}

	list := []*appspb.Application{}
	for _, a := range apps {
		list = append(list, &appspb.Application{Name: a.Name})
	}
	return &appspb.ListApplicationsResponse{
		Applications: list,
	}, nil
}

func (s *applicationServer) GetApplication(ctx context.Context, msg *appspb.GetApplicationRequest) (*appspb.GetApplicationResponse, error) {
	app, err := s.kube.GetApplication(ctx, types.NamespacedName{Name: msg.Name, Namespace: msg.Namespace})
	if err != nil {
		return nil, fmt.Errorf("could not get application \"%s\": %w", msg.Name, err)
	}

	src, deployment, err := findFluxObjects(app)
	if err != nil {
		return nil, fmt.Errorf("could not get flux objects for application \"%s\": %w", app.Name, err)
	}

	name := types.NamespacedName{Name: app.Name, Namespace: app.Namespace}

	if err := s.kube.GetResource(ctx, name, src); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("could not get source for app %s: %w", app.Name, err)
	}

	if err := s.kube.GetResource(ctx, name, deployment); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("could not get deployment for app %s: %w", app.Name, err)
	}

	// A Source is just an abstract interface, we need to get the underlying implementation.
	var srcK8sConditions []metav1.Condition
	var srcConditions []*appspb.Condition

	if src != nil {
		// An src might be nil if it is not reconciled yet,
		// in which case return nil in the response for the source_conditions key.
		switch st := src.(type) {
		case *sourcev1.GitRepository:
			srcK8sConditions = st.Status.Conditions
		case *sourcev1.HelmRepository:
			srcK8sConditions = st.Status.Conditions
		}

		srcConditions = mapConditions(srcK8sConditions)
	}

	var deploymentK8sConditions []metav1.Condition
	var deploymentConditions []*appspb.Condition
	if deployment != nil {
		// Same as a src. Deployment may not be created at this point.
		switch at := deployment.(type) {
		case *kustomizev1.Kustomization:
			deploymentK8sConditions = at.Status.Conditions
		case *helmv2.HelmRelease:
			deploymentK8sConditions = at.Status.Conditions
		}

		deploymentConditions = mapConditions(deploymentK8sConditions)
	}

	return &appspb.GetApplicationResponse{Application: &appspb.Application{
		Name:                 app.Name,
		Url:                  app.Spec.URL,
		Path:                 app.Spec.Path,
		SourceConditions:     srcConditions,
		DeploymentConditions: deploymentConditions,
	}}, nil
}

func (s *commitServer) ListCommits(ctx context.Context, msg *commitpb.ListCommitsRequest) (*commitpb.ListCommitsResponse, error) {
	logger := logger.NewApiLogger()

	appService := app.New(logger, nil, nil, s.kube, nil)

	a, err := s.kube.GetApplication(ctx, types.NamespacedName{Name: msg.Name, Namespace: msg.Namespace})
	if err != nil {
		return nil, fmt.Errorf("unable to get application for %s %w", msg.Name, err)
	}

	token, err := app.DoAppRepoCLIAuth(a.Spec.URL, logger)
	if err != nil {
		return &commitpb.ListCommitsResponse{
			Commits: []*commitpb.Commit{},
		}, fmt.Errorf("could not complete auth flow: %w", err)
	}

	pageToken := 0
	if msg.PageToken != nil {
		pageToken = int(*msg.PageToken)
	}

	params := app.CommitParams{Name: msg.Name, Namespace: msg.Namespace, GitProviderToken: token, PageSize: int(msg.PageSize), PageToken: pageToken}

	commits, err := appService.GetCommits(params)
	if err != nil {
		return &commitpb.ListCommitsResponse{
			Commits: []*commitpb.Commit{},
		}, err
	}

	if commits == nil {
		return &commitpb.ListCommitsResponse{
			Commits: []*commitpb.Commit{},
		}, nil
	}

	list := []*commitpb.Commit{}
	for _, commit := range commits {
		list = append(list, &commitpb.Commit{
			Author:     commit.Get().Author,
			Message:    utils.CleanCommitMessage(commit.Get().Message),
			CommitHash: commit.Get().Sha,
			Date:       commit.Get().CreatedAt.String(),
		})
	}
	nextPageToken := int32(pageToken + 1)
	return &commitpb.ListCommitsResponse{
		Commits:       list,
		NextPageToken: nextPageToken,
	}, nil
}

// Returns k8s objects that can be used to find the cluster objects.
// The first return argument is the source, the second is the deployment
func findFluxObjects(app *wego.Application) (client.Object, client.Object, error) {
	st := app.Spec.SourceType
	if st == "" {
		// Apps that were created before the SourceType field exists will not have a SourceType defined.
		// Assume git, since thats what the CLI defaults to.
		st = wego.SourceTypeGit
	}

	var src client.Object
	switch st {
	case wego.SourceTypeGit:
		src = &sourcev1.GitRepository{}
	case wego.SourceTypeHelm:
		src = &sourcev1.HelmRepository{}
	}

	if src == nil {
		return nil, nil, fmt.Errorf("invalid source type \"%s\"", st)
	}

	at := app.Spec.DeploymentType
	if at == "" {
		// Same as above, default to kustomize to match CLI default.
		at = wego.DeploymentTypeKustomize
	}
	var deployment client.Object
	switch at {
	case wego.DeploymentTypeHelm:
		deployment = &helmv2.HelmRelease{}
	case wego.DeploymentTypeKustomize:
		deployment = &kustomizev1.Kustomization{}
	}

	if deployment == nil {
		return nil, nil, fmt.Errorf("invalid deployment type \"%s\"", at)
	}

	return src, deployment, nil
}

// Convert k8s conditions to protobuf conditions
func mapConditions(conditions []metav1.Condition) []*appspb.Condition {
	out := []*appspb.Condition{}

	for _, c := range conditions {
		out = append(out, &appspb.Condition{
			Type:      c.Type,
			Status:    string(c.Status),
			Reason:    c.Reason,
			Message:   c.Message,
			Timestamp: int32(c.LastTransitionTime.Unix()),
		})
	}

	return out
}
