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

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"

	"github.com/weaveworks/weave-gitops/pkg/services/auth"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/apputils"
	"github.com/weaveworks/weave-gitops/pkg/middleware"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"go.uber.org/zap"
	grpcStatus "google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kstatus/status"
)

var ErrEmptyAccessToken = fmt.Errorf("access token is empty")

type applicationServer struct {
	pb.UnimplementedApplicationsServer

	appFactory   apputils.AppFactory
	jwtClient    auth.JWTClient
	log          logr.Logger
	kube         client.Client
	ghAuthClient auth.GithubAuthClient
}

// An ApplicationsConfig allows for the customization of an ApplicationsServer.
// Use the DefaultConfig() to use the default dependencies.
type ApplicationsConfig struct {
	Logger           logr.Logger
	AppFactory       apputils.AppFactory
	JwtClient        auth.JWTClient
	KubeClient       client.Client
	GithubAuthClient auth.GithubAuthClient
}

// NewApplicationsServer creates a grpc Applications server
func NewApplicationsServer(cfg *ApplicationsConfig) pb.ApplicationsServer {
	return &applicationServer{
		jwtClient:    cfg.JwtClient,
		log:          cfg.Logger,
		appFactory:   cfg.AppFactory,
		kube:         cfg.KubeClient,
		ghAuthClient: cfg.GithubAuthClient,
	}
}

// DefaultConfig creates a populated config with the dependencies for a Server
func DefaultConfig() (*ApplicationsConfig, error) {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("could not create zap logger: %v", err)
	}

	logr := zapr.NewLogger(zapLog)

	rand.Seed(time.Now().UnixNano())
	secretKey := rand.String(20)
	jwtClient := auth.NewJwtClient(secretKey)

	_, rawClient, err := kube.NewKubeHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("could not create kube http client: %w", err)
	}

	return &ApplicationsConfig{
		Logger:           logr,
		AppFactory:       &apputils.DefaultAppFactory{},
		JwtClient:        jwtClient,
		KubeClient:       rawClient,
		GithubAuthClient: auth.NewGithubAuthProvider(http.DefaultClient),
	}, nil
}

// NewApplicationsHandler allow for other applications to embed the Weave GitOps HTTP API.
// This handler can be muxed with other services or used as a standalone service.
func NewApplicationsHandler(ctx context.Context, cfg *ApplicationsConfig, opts ...runtime.ServeMuxOption) (http.Handler, error) {
	appsSrv := NewApplicationsServer(cfg)

	mux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(cfg.Logger))
	httpHandler := middleware.WithLogging(cfg.Logger, mux)
	httpHandler = middleware.WithProviderToken(cfg.JwtClient, httpHandler, cfg.Logger)

	if err := pb.RegisterApplicationsHandlerServer(ctx, mux, appsSrv); err != nil {
		return nil, fmt.Errorf("could not register application: %w", err)
	}

	return httpHandler, nil
}

func (s *applicationServer) ListApplications(ctx context.Context, msg *pb.ListApplicationsRequest) (*pb.ListApplicationsResponse, error) {
	kubeService, kubeErr := s.appFactory.GetKubeService()
	if kubeErr != nil {
		return nil, fmt.Errorf("failed to create kube service: %w", kubeErr)
	}

	apps, err := kubeService.GetApplications(ctx, msg.Namespace)
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
	kubeClient, kubeErr := s.appFactory.GetKubeService()
	if kubeErr != nil {
		return nil, fmt.Errorf("failed to create kube service: %w", kubeErr)
	}

	app, err := kubeClient.GetApplication(ctx, types.NamespacedName{Name: msg.Name, Namespace: msg.Namespace})
	if err != nil {
		return nil, fmt.Errorf("could not get application \"%s\": %w", msg.Name, err)
	}

	src, deployment, err := findFluxObjects(app)
	if err != nil {
		return nil, fmt.Errorf("could not get flux objects for application \"%s\": %w", app.Name, err)
	}

	name := types.NamespacedName{Name: app.Name, Namespace: app.Namespace}

	if err := kubeClient.GetResource(ctx, name, src); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("could not get source for app %s: %w", app.Name, err)
	}

	if err := kubeClient.GetResource(ctx, name, deployment); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("could not get deployment for app %s: %w", app.Name, err)
	}

	reconciledKinds := []*pb.GroupVersionKind{}

	var (
		kust           *kustomizev1.Kustomization
		helmRelease    *helmv2.HelmRelease
		deploymentType pb.AutomationKind
	)

	if deployment != nil {
		// Same as a src. Deployment may not be created at this point.
		switch at := deployment.(type) {
		case *kustomizev1.Kustomization:
			kust = at
			reconciledKinds = addReconciledKinds(reconciledKinds, at)
			deploymentType = pb.AutomationKind_Kustomize
		case *helmv2.HelmRelease:
			helmRelease = at
			deploymentType = pb.AutomationKind_Helm
		}
	}

	return &pb.GetApplicationResponse{Application: &pb.Application{
		Name:                  app.Name,
		Namespace:             app.Namespace,
		Url:                   app.Spec.URL,
		Path:                  app.Spec.Path,
		DeploymentType:        deploymentType,
		Kustomization:         mapKustomizationSpecToResponse(kust),
		HelmRelease:           mapHelmReleaseSpecToResponse(helmRelease),
		Source:                mapSourceSpecToReponse(src),
		ReconciledObjectKinds: reconciledKinds,
	}}, nil
}

func mapHelmReleaseSpecToResponse(helm *helmv2.HelmRelease) *pb.HelmRelease {
	if helm == nil {
		return nil
	}

	return &pb.HelmRelease{
		Name:            helm.Name,
		Namespace:       helm.Namespace,
		TargetNamespace: helm.Spec.TargetNamespace,
		Conditions:      mapConditions(helm.Status.Conditions),
		Chart: &pb.HelmChart{
			Chart:       helm.Spec.Chart.Spec.Chart,
			Version:     helm.Spec.Chart.Spec.Version,
			ValuesFiles: helm.Spec.Chart.Spec.ValuesFiles,
		},
	}
}

func mapKustomizationSpecToResponse(kust *kustomizev1.Kustomization) *pb.Kustomization {
	if kust == nil {
		return nil
	}

	return &pb.Kustomization{
		Name:                kust.Name,
		Namespace:           kust.Namespace,
		TargetNamespace:     kust.Spec.TargetNamespace,
		Path:                kust.Spec.Path,
		Conditions:          mapConditions(kust.Status.Conditions),
		Interval:            kust.Spec.Interval.Duration.String(),
		Prune:               kust.Spec.Prune,
		LastAppliedRevision: kust.Status.LastAppliedRevision,
	}
}

func mapSourceSpecToReponse(src client.Object) *pb.Source {
	// An src might be nil if it is not reconciled yet,
	// in which case return nil in the response for the source_conditions key.
	source := &pb.Source{}
	if src == nil {
		return source
	}

	switch st := src.(type) {
	case *sourcev1.GitRepository:
		source.Name = st.Name
		source.Namespace = st.Namespace
		source.Url = st.Spec.URL
		source.Type = pb.Source_Git
		source.Interval = st.Spec.Interval.Duration.String()
		source.Suspend = st.Spec.Suspend

		if st.Spec.Timeout != nil {
			source.Timeout = st.Spec.Timeout.Duration.String()
		}

		if st.Spec.Reference != nil {
			source.Reference = st.Spec.Reference.Branch
		}

		source.Conditions = mapConditions(st.Status.Conditions)
	case *sourcev1.HelmRepository:
		source.Name = st.Name
		source.Namespace = st.Namespace
		source.Url = st.Spec.URL
		source.Type = pb.Source_Helm
		source.Interval = st.Spec.Interval.Duration.String()
		source.Suspend = st.Spec.Suspend

		if st.Spec.Timeout != nil {
			source.Timeout = st.Spec.Timeout.Duration.String()
		}

		source.Conditions = mapConditions(st.Status.Conditions)
	}

	return source
}

func (s *applicationServer) AddApplication(ctx context.Context, msg *pb.AddApplicationRequest) (*pb.AddApplicationResponse, error) {
	appUrl, err := gitproviders.NewNormalizedRepoURL(msg.Url)
	if err != nil {
		return nil, grpcStatus.Error(codes.InvalidArgument, fmt.Sprintf("unable to parse app url %q: %s", msg.Url, err))
	}

	var configUrl gitproviders.NormalizedRepoURL
	if msg.ConfigUrl != "" {
		configUrl, err = gitproviders.NewNormalizedRepoURL(msg.ConfigUrl)
		if err != nil {
			return nil, grpcStatus.Error(codes.InvalidArgument, fmt.Sprintf("unable to parse config url %q: %s", msg.ConfigUrl, err))
		}
	}

	appSrv, err := s.appFactory.GetAppServiceForAdd(ctx, apputils.AddServiceParams{
		URL:              appUrl.String(),
		ConfigURL:        configUrl.String(),
		Namespace:        msg.Namespace,
		IsHelmRepository: false,
		DryRun:           false,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create app service: %w", err)
	}

	token, err := middleware.ExtractProviderToken(ctx)
	if err != nil {
		return nil, grpcStatus.Error(codes.Unauthenticated, fmt.Sprintf("token error: %s", err.Error()))
	}

	params := app.AddParams{
		Name:             msg.Name,
		Namespace:        msg.Namespace,
		Url:              appUrl.String(),
		Path:             configUrl.String(),
		GitProviderToken: token.AccessToken,
		Branch:           msg.Branch,
		AutoMerge:        msg.AutoMerge,
		AppConfigUrl:     msg.ConfigUrl,
	}

	if err := appSrv.Add(params); err != nil {
		return nil, fmt.Errorf("error adding app: %w", err)
	}

	return &pb.AddApplicationResponse{
		Success: true,
		Application: &pb.Application{
			Name:      msg.Name,
			Namespace: msg.Namespace,
		},
	}, nil
}

//Until the middleware is done this function will not be able to get the token and will fail
func (s *applicationServer) ListCommits(ctx context.Context, msg *pb.ListCommitsRequest) (*pb.ListCommitsResponse, error) {
	providerToken, err := middleware.ExtractProviderToken(ctx)
	if err != nil {
		return nil, grpcStatus.Error(codes.Unauthenticated, fmt.Sprintf("error listing commits: %s", err.Error()))
	}

	pageToken := 0
	if msg.PageToken != nil {
		pageToken = int(*msg.PageToken)
	}

	params := app.CommitParams{
		Name:             msg.Name,
		Namespace:        msg.Namespace,
		GitProviderToken: providerToken.AccessToken,
		PageSize:         int(msg.PageSize),
		PageToken:        pageToken,
	}

	appService, appErr := s.appFactory.GetAppService(ctx, params.Name, params.Namespace)
	if appErr != nil {
		return nil, fmt.Errorf("failed to create app service: %w", appErr)
	}

	application, err := appService.Get(types.NamespacedName{Name: params.Name, Namespace: params.Namespace})
	if err != nil {
		return nil, fmt.Errorf("unable to get application for %s %w", params.Name, err)
	}

	commits, err := appService.GetCommits(params, application)
	if err != nil {
		return nil, err
	}

	list := []*pb.Commit{}

	for _, commit := range commits {
		c := commit.Get()

		list = append(list, &pb.Commit{
			Author:  c.Author,
			Message: utils.CleanCommitMessage(c.Message),
			Hash:    utils.ConvertCommitHashToShort(c.Sha),
			Date:    utils.CleanCommitCreatedAt(c.CreatedAt),
			Url:     utils.ConvertCommitURLToShort(c.URL),
		})
	}

	nextPageToken := int32(pageToken + 1)

	return &pb.ListCommitsResponse{
		Commits:       list,
		NextPageToken: nextPageToken,
	}, nil
}

const KustomizeNameKey string = "kustomize.toolkit.fluxcd.io/name"
const KustomizeNamespaceKey string = "kustomize.toolkit.fluxcd.io/namespace"

func (s *applicationServer) GetReconciledObjects(ctx context.Context, msg *pb.GetReconciledObjectsReq) (*pb.GetReconciledObjectsRes, error) {
	if msg.AutomationKind == pb.AutomationKind_Helm {
		return nil, grpcStatus.Error(codes.Unimplemented, "Helm is not currently supported for this method")
	}

	result := []unstructured.Unstructured{}

	for _, gvk := range msg.Kinds {
		list := unstructured.UnstructuredList{}

		list.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   gvk.Group,
			Kind:    gvk.Kind,
			Version: gvk.Version,
		})

		opts := client.MatchingLabels{
			KustomizeNameKey:      msg.AutomationName,
			KustomizeNamespaceKey: msg.AutomationNamespace,
		}

		if err := s.kube.List(ctx, &list, opts); err != nil {
			return nil, fmt.Errorf("could not get unstructured list: %s\n", err)
		}

		result = append(result, list.Items...)
	}

	objects := []*pb.UnstructuredObject{}

	for _, obj := range result {
		res, err := status.Compute(&obj)

		if err != nil {
			return nil, fmt.Errorf("could not get status for %s: %w", obj.GetName(), err)
		}

		objects = append(objects, &pb.UnstructuredObject{
			GroupVersionKind: &pb.GroupVersionKind{
				Group:   obj.GetObjectKind().GroupVersionKind().Group,
				Version: obj.GetObjectKind().GroupVersionKind().GroupVersion().Version,
				Kind:    obj.GetKind(),
			},
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
			Status:    res.Status.String(),
			Uid:       string(obj.GetUID()),
		})
	}

	return &pb.GetReconciledObjectsRes{Objects: objects}, nil
}

func (s *applicationServer) GetChildObjects(ctx context.Context, msg *pb.GetChildObjectsReq) (*pb.GetChildObjectsRes, error) {
	list := unstructured.UnstructuredList{}

	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   msg.GroupVersionKind.Group,
		Version: msg.GroupVersionKind.Version,
		Kind:    msg.GroupVersionKind.Kind,
	})

	if err := s.kube.List(ctx, &list); err != nil {
		return nil, fmt.Errorf("could not get unstructured object: %s\n", err)
	}

	objects := []*pb.UnstructuredObject{}

Items:
	for _, obj := range list.Items {

		refs := obj.GetOwnerReferences()

		for _, ref := range refs {
			if ref.UID != types.UID(msg.ParentUid) {
				// This is not the child we are looking for.
				// Skip the rest of the operations in Items loops.
				// The is effectively an early return.
				continue Items
			}
		}

		statusResult, err := status.Compute(&obj)
		if err != nil {
			return nil, fmt.Errorf("could not get status for %s: %w", obj.GetName(), err)
		}
		objects = append(objects, &pb.UnstructuredObject{
			GroupVersionKind: &pb.GroupVersionKind{
				Group:   obj.GetObjectKind().GroupVersionKind().Group,
				Version: obj.GetObjectKind().GroupVersionKind().GroupVersion().Version,
				Kind:    obj.GetKind(),
			},
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
			Status:    statusResult.Status.String(),
			Uid:       string(obj.GetUID()),
		})
	}

	return &pb.GetChildObjectsRes{Objects: objects}, nil
}

func (s *applicationServer) GetGithubDeviceCode(ctx context.Context, msg *pb.GetGithubDeviceCodeRequest) (*pb.GetGithubDeviceCodeResponse, error) {
	res, err := s.ghAuthClient.GetDeviceCode()
	if err != nil {
		return nil, fmt.Errorf("error doing github code request: %w", err)
	}

	return &pb.GetGithubDeviceCodeResponse{
		UserCode:      res.UserCode,
		ValidationURI: res.VerificationURI,
		DeviceCode:    res.DeviceCode,
		Interval:      int32(res.Interval),
	}, nil
}

func (s *applicationServer) GetGithubAuthStatus(ctx context.Context, msg *pb.GetGithubAuthStatusRequest) (*pb.GetGithubAuthStatusResponse, error) {
	token, err := s.ghAuthClient.GetDeviceCodeAuthStatus(msg.DeviceCode)
	if err == auth.ErrAuthPending {
		return nil, grpcStatus.Error(codes.Unauthenticated, err.Error())
	} else if err != nil {
		return nil, fmt.Errorf("error getting github device code status: %w", err)
	}

	t, err := s.jwtClient.GenerateJWT(auth.ExpirationTime, gitproviders.GitProviderGitHub, token)
	if err != nil {
		return nil, fmt.Errorf("could not generate token: %w", err)
	}

	return &pb.GetGithubAuthStatusResponse{AccessToken: t}, nil
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
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "%s expected github or gitlab, got %s", ErrBadProvider, msg.ProviderName)
	}

	if msg.AccessToken == "" {
		return nil, grpcStatus.Error(codes.InvalidArgument, ErrEmptyAccessToken.Error())
	}

	token, err := s.jwtClient.GenerateJWT(auth.ExpirationTime, gitproviders.GitProviderName(msg.GetProviderName()), msg.GetAccessToken())
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Internal, "error generating jwt token. %s", err)
	}

	return &pb.AuthenticateResponse{Token: token}, nil
}

func addReconciledKinds(arr []*pb.GroupVersionKind, kustomization *kustomizev1.Kustomization) []*pb.GroupVersionKind {
	if kustomization.Status.Snapshot == nil {
		return arr
	}

	found := map[string]bool{}

	for _, gvks := range kustomization.Status.Snapshot.NamespacedKinds() {
		for _, gvk := range gvks {
			s := gvk.String()

			if !found[s] {
				found[s] = true

				arr = append(arr, &pb.GroupVersionKind{
					Group:   gvk.Group,
					Version: gvk.Version,
					Kind:    gvk.Kind,
				})
			}
		}
	}

	for _, gvk := range kustomization.Status.Snapshot.NonNamespacedKinds() {
		s := gvk.String()
		if _, exists := found[s]; !exists {
			found[s] = true

			arr = append(arr, &pb.GroupVersionKind{
				Group:   gvk.Group,
				Version: gvk.Version,
				Kind:    gvk.Kind,
			})
		}
	}

	return arr
}
