package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"

	"github.com/weaveworks/weave-gitops/pkg/services/auth"

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
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type key int

const tokenKey key = iota

var ErrEmptyAccessToken = fmt.Errorf("access token is empty")

type applicationServer struct {
	pb.UnimplementedApplicationsServer

	log       logr.Logger
	app       *app.App
	jwtClient auth.JWTClient
}

// An ApplicationConfig allows for the customization of an ApplicationsServer.
// Use the DefaultConfig() to use the default dependencies.
type ApplicationConfig struct {
	Logger    logr.Logger
	App       *app.App
	JwtClient auth.JWTClient
}

//Remove when middleware is done
type contextVals struct {
	Token *oauth2.Token
}

// NewApplicationsServer creates a grpc Applications server
func NewApplicationsServer(cfg *ApplicationConfig) pb.ApplicationsServer {
	return &applicationServer{
		jwtClient: cfg.JwtClient,
		log:       cfg.Logger,
		app:       cfg.App,
	}
}

// DefaultConfig creates a populated config with the dependencies for a Server
func DefaultConfig() (*ApplicationConfig, error) {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("could not create zap logger: %v", err)
	}
	logr := zapr.NewLogger(zapLog)

	kubeClient, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("could not create kube http client: %w", err)
	}

	appSrv := app.New(nil, nil, nil, kubeClient, nil)

	rand.Seed(time.Now().UnixNano())
	secretKey := rand.String(20)

	jwtClient := auth.NewJwtClient(secretKey)

	return &ApplicationConfig{
		Logger:    logr,
		App:       appSrv,
		JwtClient: jwtClient,
	}, nil
}

// NewApplicationsHandler allow for other applications to embed the Weave GitOps HTTP API.
// This handler can be muxed with other services or used as a standalone service.
func NewApplicationsHandler(ctx context.Context, cfg *ApplicationConfig, opts ...runtime.ServeMuxOption) (http.Handler, error) {
	appsSrv := NewApplicationsServer(cfg)

	mux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(cfg.Logger))
	httpHandler := middleware.WithLogging(cfg.Logger, mux)

	if err := pb.RegisterApplicationsHandlerServer(ctx, mux, appsSrv); err != nil {
		return nil, fmt.Errorf("could not register application: %w", err)
	}

	return httpHandler, nil
}

func (s *applicationServer) ListApplications(ctx context.Context, msg *pb.ListApplicationsRequest) (*pb.ListApplicationsResponse, error) {
	apps, err := s.app.Kube.GetApplications(ctx, msg.GetNamespace())
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

func (s *applicationServer) GetApplication(ctx context.Context, msg *pb.GetApplicationRequest) (*pb.GetApplicationResponse, error) {
	app, err := s.app.Kube.GetApplication(ctx, types.NamespacedName{Name: msg.Name, Namespace: msg.Namespace})
	if err != nil {
		return nil, fmt.Errorf("could not get application \"%s\": %w", msg.Name, err)
	}

	src, deployment, err := findFluxObjects(app)
	if err != nil {
		return nil, fmt.Errorf("could not get flux objects for application \"%s\": %w", app.Name, err)
	}

	name := types.NamespacedName{Name: app.Name, Namespace: app.Namespace}

	if err := s.app.Kube.GetResource(ctx, name, src); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("could not get source for app %s: %w", app.Name, err)
	}

	if err := s.app.Kube.GetResource(ctx, name, deployment); err != nil {
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

//Temporary solution to get this to build until middleware is done
func extractToken(ctx context.Context) (string, error) {
	c := ctx.Value(tokenKey)

	vals, ok := c.(contextVals)
	if !ok {
		return "", errors.New("could not get token from context")
	}

	if vals.Token == nil || vals.Token.AccessToken == "" {
		return "", errors.New("no token specified")
	}

	return vals.Token.AccessToken, nil
}

//Until the middleware is done this function will not be able to get the token and will fail
func (s *applicationServer) ListCommits(ctx context.Context, msg *pb.ListCommitsRequest) (*pb.ListCommitsResponse, error) {
	vals := contextVals{Token: &oauth2.Token{AccessToken: "temptoken"}}
	ctx = context.WithValue(ctx, tokenKey, vals)

	token, err := extractToken(ctx)
	if err != nil {
		return nil, err
	}

	pageToken := 0
	if msg.PageToken != nil {
		pageToken = int(*msg.PageToken)
	}

	params := app.CommitParams{
		Name:             msg.Name,
		Namespace:        msg.Namespace,
		GitProviderToken: token,
		PageSize:         int(msg.PageSize),
		PageToken:        pageToken,
	}

	application, err := s.app.Kube.GetApplication(ctx, types.NamespacedName{Name: params.Name, Namespace: params.Namespace})
	if err != nil {
		return nil, fmt.Errorf("unable to get application for %s %w", params.Name, err)
	}

	if application.Spec.SourceType == wego.SourceTypeHelm {
		return nil, fmt.Errorf("unable to get commits for a helm chart")
	}

	commits, err := s.app.GetCommits(params, application)
	if err != nil {
		return nil, err
	}

	list := []*pb.Commit{}
	for _, commit := range commits {
		list = append(list, &pb.Commit{
			Author:     commit.Get().Author,
			Message:    utils.CleanCommitMessage(commit.Get().Message),
			CommitHash: commit.Get().Sha,
			Date:       commit.Get().CreatedAt.String(),
		})
	}
	nextPageToken := int32(pageToken + 1)
	return &pb.ListCommitsResponse{
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

var ErrBadProvider = errors.New("wrong provider name")

// Authenticate generates and returns a jwt token using git provider name and git provider token
func (s *applicationServer) Authenticate(_ context.Context, msg *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {

	if !strings.HasPrefix(github.DefaultDomain, msg.ProviderName) &&
		!strings.HasPrefix(gitlab.DefaultDomain, msg.ProviderName) {
		return nil, status.Errorf(codes.InvalidArgument, "%s expected github or gitlab, got %s", ErrBadProvider, msg.ProviderName)
	}

	if msg.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyAccessToken.Error())
	}

	token, err := s.jwtClient.GenerateJWT(auth.ExpirationTime, gitproviders.GitProviderName(msg.GetProviderName()), msg.GetAccessToken())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error generating jwt token. %s", err)
	}

	return &pb.AuthenticateResponse{Token: token}, nil
}
