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
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/middleware"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type server struct {
	pb.UnimplementedApplicationsServer

	kube kube.Kube
	log  logr.Logger
}

// An ApplicationsConfig allows for the customization of an ApplicationsServer.
// Use the DefaultConfig() to use the default dependencies.
type ApplicationsConfig struct {
	Logger     logr.Logger
	KubeClient kube.Kube
}

// NewApplicationsServer creates a grpc Applications server
func NewApplicationsServer(cfg *ApplicationsConfig) pb.ApplicationsServer {
	return &server{
		kube: cfg.KubeClient,
		log:  cfg.Logger,
	}
}

// DefaultConfig creates a populated config with the dependencies for an ApplicationsServer
func DefaultConfig() (*ApplicationsConfig, error) {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("could not create zap logger: %v", err)
	}
	logr := zapr.NewLogger(zapLog)

	kubeClient, err := kube.NewKubeHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("could not create kube http client: %w", err)
	}

	return &ApplicationsConfig{
		Logger:     logr,
		KubeClient: kubeClient,
	}, nil
}

// NewApplicationsHandler allow for other applications to embed the Weave GitOps Applications HTTP API.
// This handler can be muxed with other services or used as a standalone service.
func NewApplicationsHandler(ctx context.Context, cfg *ApplicationsConfig, opts ...runtime.ServeMuxOption) (http.Handler, error) {
	appsSrv := NewApplicationsServer(cfg)

	mux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(cfg.Logger))
	httpHandler := middleware.WithLogging(cfg.Logger, mux)

	if err := pb.RegisterApplicationsHandlerServer(ctx, mux, appsSrv); err != nil {
		return nil, fmt.Errorf("could not register application: %w", err)
	}

	return httpHandler, nil
}

func (s *server) ListApplications(ctx context.Context, msg *pb.ListApplicationsRequest) (*pb.ListApplicationsResponse, error) {
	apps, err := s.kube.GetApplications(ctx, msg.GetNamespace())
	if err != nil {
		return nil, err
	}

	if apps == nil {
		return &pb.ListApplicationsResponse{
			Applications: []*pb.Application{},
		}, nil
	}

	list := []*pb.Application{}
	for _, a := range apps {
		list = append(list, &pb.Application{Name: a.Name})
	}
	return &pb.ListApplicationsResponse{
		Applications: list,
	}, nil
}

func (s *server) GetApplication(ctx context.Context, msg *pb.GetApplicationRequest) (*pb.GetApplicationResponse, error) {
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
	var srcConditions []*pb.Condition

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
	var deploymentConditions []*pb.Condition
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

	return &pb.GetApplicationResponse{Application: &pb.Application{
		Name:                 app.Name,
		Url:                  app.Spec.URL,
		Path:                 app.Spec.Path,
		SourceConditions:     srcConditions,
		DeploymentConditions: deploymentConditions,
	}}, nil
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
func mapConditions(conditions []metav1.Condition) []*pb.Condition {
	out := []*pb.Condition{}

	for _, c := range conditions {
		out = append(out, &pb.Condition{
			Type:      c.Type,
			Status:    string(c.Status),
			Reason:    c.Reason,
			Message:   c.Message,
			Timestamp: int32(c.LastTransitionTime.Unix()),
		})
	}

	return out
}
