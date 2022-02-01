package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/controller-runtime/pkg/client"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/server/internal"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"github.com/weaveworks/weave-gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/services/applicationv2"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

const DefaultPort = "9001"

var (
	ErrEmptyAccessToken = errors.New("access token is empty")
	ErrBadProvider      = errors.New("wrong provider name")
)

// Flux owner labels
var (
	KustomizeNameKey      = fmt.Sprintf("%s/name", kustomizev2.GroupVersion.Group)
	KustomizeNamespaceKey = fmt.Sprintf("%s/namespace", kustomizev2.GroupVersion.Group)
	HelmNameKey           = fmt.Sprintf("%s/name", helmv2.GroupVersion.Group)
	HelmNamespaceKey      = fmt.Sprintf("%s/namespace", helmv2.GroupVersion.Group)
)

type applicationServer struct {
	pb.UnimplementedApplicationsServer

	factory        services.Factory
	jwtClient      auth.JWTClient
	log            logr.Logger
	ghAuthClient   auth.GithubAuthClient
	fetcherFactory applicationv2.FetcherFactory
	glAuthClient   auth.GitlabAuthClient
	clientGetter   kube.ClientGetter
	kubeGetter     kube.KubeGetter
}

// An ApplicationsConfig allows for the customization of an ApplicationsServer.
// Use the DefaultConfig() to use the default dependencies.
type ApplicationsConfig struct {
	Logger           logr.Logger
	Factory          services.Factory
	JwtClient        auth.JWTClient
	GithubAuthClient auth.GithubAuthClient
	FetcherFactory   applicationv2.FetcherFactory
	GitlabAuthClient auth.GitlabAuthClient
	ClusterConfig    kube.ClusterConfig
}

var _ applicationv2.FetcherFactory = &DefaultFetcherFactory{}

// DefaultFetcherFactory creates applicationv2.Fetcher objects from a Kubernetes
// client
type DefaultFetcherFactory struct {
}

// NewDefaultFetcherFactory returns a new DefaultFetcherFactory
func NewDefaultFetcherFactory() applicationv2.FetcherFactory {
	return &DefaultFetcherFactory{}
}

// Create uses a Kubernetes client to create an applicationv2.Fetcher object.
func (f *DefaultFetcherFactory) Create(client client.Client) applicationv2.Fetcher {
	return applicationv2.NewFetcher(client)
}

// NewApplicationsServer creates a grpc Applications server
func NewApplicationsServer(cfg *ApplicationsConfig, setters ...ApplicationsOption) pb.ApplicationsServer {
	configGetter := NewImpersonatingConfigGetter(cfg.ClusterConfig.DefaultConfig, false)
	clientGetter := kube.NewDefaultClientGetter(configGetter, cfg.ClusterConfig.ClusterName)
	kubeGetter := kube.NewDefaultKubeGetter(configGetter, cfg.ClusterConfig.ClusterName)

	args := &ApplicationsOptions{
		ClientGetter: clientGetter,
		KubeGetter:   kubeGetter,
	}

	for _, setter := range setters {
		setter(args)
	}

	return &applicationServer{
		jwtClient:      cfg.JwtClient,
		log:            cfg.Logger,
		factory:        cfg.Factory,
		ghAuthClient:   cfg.GithubAuthClient,
		fetcherFactory: cfg.FetcherFactory,
		glAuthClient:   cfg.GitlabAuthClient,
		clientGetter:   args.ClientGetter,
		kubeGetter:     args.KubeGetter,
	}
}

// DefaultApplicationsConfig creates a populated config with the dependencies for a Server
func DefaultApplicationsConfig() (*ApplicationsConfig, error) {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("could not create zap logger: %v", err)
	}

	logr := zapr.NewLogger(zapLog)

	rand.Seed(time.Now().UnixNano())
	secretKey := rand.String(20)
	envSecretKey := os.Getenv("GITOPS_JWT_ENCRYPTION_SECRET")

	if envSecretKey != "" {
		secretKey = envSecretKey
	}

	jwtClient := auth.NewJwtClient(secretKey)

	rest, clusterName, err := kube.RestConfig()
	if err != nil {
		return nil, fmt.Errorf("could not create client config: %w", err)
	}

	fluxClient := flux.New(osys.New(), &runner.CLIRunner{})

	return &ApplicationsConfig{
		Logger:           logr,
		Factory:          services.NewFactory(fluxClient, internal.NewApiLogger(zapLog)),
		JwtClient:        jwtClient,
		FetcherFactory:   NewDefaultFetcherFactory(),
		GithubAuthClient: auth.NewGithubAuthClient(http.DefaultClient),
		GitlabAuthClient: auth.NewGitlabAuthClient(http.DefaultClient),
		ClusterConfig: kube.ClusterConfig{
			DefaultConfig: rest,
			ClusterName:   clusterName,
		},
	}, nil
}

func (s *applicationServer) ListApplications(ctx context.Context, msg *pb.ListApplicationsRequest) (*pb.ListApplicationsResponse, error) {
	rawClient, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	fetcher := s.fetcherFactory.Create(rawClient)

	apps, err := fetcher.List(ctx, msg.Namespace)
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
	kubeClient, err := s.kubeGetter.Kube(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create kube service: %w", err)
	}

	app, err := kubeClient.GetApplication(ctx, types.NamespacedName{Name: msg.Name, Namespace: msg.Namespace})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, grpcStatus.Errorf(codes.NotFound, "not found: %s", err.Error())
		}

		return nil, fmt.Errorf("could not get application %q: %w", msg.Name, err)
	}

	src, deployment, err := findFluxObjects(app)
	if err != nil {
		return nil, fmt.Errorf("could not get flux objects for application %q: %w", app.Name, err)
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

	var (
		kust            *kustomizev2.Kustomization
		helmRelease     *helmv2.HelmRelease
		deploymentType  pb.AutomationKind
		reconciledKinds []*pb.GroupVersionKind
	)

	if deployment != nil {
		// Same as a src. Deployment may not be created at this point.
		switch at := deployment.(type) {
		case *kustomizev2.Kustomization:
			kust = at
			deploymentType = pb.AutomationKind_Kustomize
			reconciledKinds, err = getKustomizeInventory(at)

			if err != nil {
				return nil, err
			}
		case *helmv2.HelmRelease:
			helmRelease = at
			deploymentType = pb.AutomationKind_Helm
			reconciledKinds, err = getHelmInventory(at, kubeClient)

			if err != nil {
				return nil, err
			}
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

func (s *applicationServer) AddApplication(ctx context.Context, msg *pb.AddApplicationRequest) (*pb.AddApplicationResponse, error) {
	token, err := middleware.ExtractProviderToken(ctx)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Unauthenticated, "token error: %s", err.Error())
	}

	kubeClient, err := s.kubeGetter.Kube(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create kube service: %w", err)
	}

	appUrl, err := gitproviders.NewRepoURL(msg.Url, false)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "unable to parse app url %q: %s", msg.Url, err)
	}

	wegoConfig, err := kubeClient.GetWegoConfig(ctx, msg.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed getting wego config")
	}

	configRepo, err := gitproviders.NewRepoURL(wegoConfig.ConfigRepo, true)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "unable to parse config url %q: %s", wegoConfig.ConfigRepo, err)
	}

	appSrv, err := s.factory.GetAppService(ctx, kubeClient)
	if err != nil {
		return nil, fmt.Errorf("could not create app service: %w", err)
	}

	params := app.AddParams{
		Name:             msg.Name,
		Namespace:        msg.Namespace,
		Url:              appUrl.String(),
		Path:             msg.Path,
		GitProviderToken: token.AccessToken,
		Branch:           msg.Branch,
		AutoMerge:        msg.AutoMerge,
		ConfigRepo:       configRepo.String(),
	}

	client := internal.NewGitProviderClient(token.AccessToken)

	gitClient, gitProvider, err := s.factory.GetGitClients(ctx, kubeClient, client, services.GitConfigParams{
		URL:        msg.Url,
		ConfigRepo: wegoConfig.ConfigRepo,
		Namespace:  msg.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get git clients: %w", err)
	}

	if err := appSrv.Add(gitClient, gitProvider, params); err != nil {
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

func (s *applicationServer) RemoveApplication(ctx context.Context, msg *pb.RemoveApplicationRequest) (*pb.RemoveApplicationResponse, error) {
	token, err := middleware.ExtractProviderToken(ctx)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Unauthenticated, "token error: %s", err.Error())
	}

	kubeClient, err := s.kubeGetter.Kube(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create kube service: %w", err)
	}

	application, err := kubeClient.GetApplication(ctx, types.NamespacedName{Name: msg.Name, Namespace: msg.Namespace})
	if err != nil {
		return nil, fmt.Errorf("could not get application %q: %w", msg.Name, err)
	}

	appSrv, err := s.factory.GetAppService(ctx, kubeClient)
	if err != nil {
		return nil, fmt.Errorf("could not create app service: %w", err)
	}

	client := internal.NewGitProviderClient(token.AccessToken)

	gitClient, gitProvider, err := s.factory.GetGitClients(ctx, kubeClient, client, services.GitConfigParams{
		URL:        application.Spec.URL,
		ConfigRepo: application.Spec.ConfigRepo,
		Namespace:  msg.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get git clients: %w", err)
	}

	removeParams := app.RemoveParams{
		Name:             msg.Name,
		Namespace:        msg.Namespace,
		DryRun:           false,
		GitProviderToken: token.AccessToken,
		AutoMerge:        msg.AutoMerge,
	}

	if err := appSrv.Remove(gitClient, gitProvider, removeParams); err != nil {
		return nil, fmt.Errorf("error removing app: %w", err)
	}

	return &pb.RemoveApplicationResponse{Success: true, RepoUrl: application.Spec.ConfigRepo}, nil
}

func (s *applicationServer) SyncApplication(ctx context.Context, msg *pb.SyncApplicationRequest) (*pb.SyncApplicationResponse, error) {
	kubeClient, err := s.kubeGetter.Kube(ctx)
	if err != nil {
		return &pb.SyncApplicationResponse{
			Success: false,
		}, fmt.Errorf("failed to create kube service: %w", err)
	}

	appSrv := &app.AppSvc{
		Kube:  kubeClient,
		Clock: clock.New(),
	}
	if err := appSrv.Sync(app.SyncParams{Name: msg.Name, Namespace: msg.Namespace}); err != nil {
		return &pb.SyncApplicationResponse{
			Success: false,
		}, fmt.Errorf("error syncing app: %w", err)
	}

	return &pb.SyncApplicationResponse{
		Success: true,
	}, nil
}

//Until the middleware is done this function will not be able to get the token and will fail
func (s *applicationServer) ListCommits(ctx context.Context, msg *pb.ListCommitsRequest) (*pb.ListCommitsResponse, error) {
	providerToken, err := middleware.ExtractProviderToken(ctx)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Unauthenticated, "error listing commits: %s", err.Error())
	}

	cl, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	kubeClient, err := s.kubeGetter.Kube(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create kube service: %w", err)
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

	application := &wego.Application{}
	if err := cl.Get(ctx, types.NamespacedName{Name: msg.Name, Namespace: msg.Namespace}, application); err != nil {
		return nil, fmt.Errorf("could not get app %q in namespace %q: %w", msg.Name, msg.Namespace, err)
	}

	appService, err := s.factory.GetAppService(ctx, kubeClient)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Unauthenticated, "failed to create app service: %s", err.Error())
	}

	client := internal.NewGitProviderClient(providerToken.AccessToken)

	_, gitProvider, err := s.factory.GetGitClients(ctx, kubeClient, client, services.GitConfigParams{
		URL:        application.Spec.URL,
		ConfigRepo: application.Spec.ConfigRepo,
		Namespace:  msg.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get git clients: %w", err)
	}

	commits, err := appService.GetCommits(gitProvider, params, application)
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

func (s *applicationServer) GetReconciledObjects(ctx context.Context, msg *pb.GetReconciledObjectsReq) (*pb.GetReconciledObjectsRes, error) {
	cl, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	var opts client.MatchingLabels

	switch msg.AutomationKind {
	case pb.AutomationKind_Kustomize:
		opts = client.MatchingLabels{
			KustomizeNameKey:      msg.AutomationName,
			KustomizeNamespaceKey: msg.AutomationNamespace,
		}
	case pb.AutomationKind_Helm:
		opts = client.MatchingLabels{
			HelmNameKey:      msg.AutomationName,
			HelmNamespaceKey: msg.AutomationNamespace,
		}
	default:
		return nil, fmt.Errorf("unsupported application kind: %s", msg.AutomationKind.String())
	}

	result := []unstructured.Unstructured{}

	for _, gvk := range msg.Kinds {
		list := unstructured.UnstructuredList{}

		list.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   gvk.Group,
			Kind:    gvk.Kind,
			Version: gvk.Version,
		})

		if err := cl.List(ctx, &list, opts); err != nil {
			return nil, fmt.Errorf("could not get unstructured list: %s", err)
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
	cl, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	list := unstructured.UnstructuredList{}

	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   msg.GroupVersionKind.Group,
		Version: msg.GroupVersionKind.Version,
		Kind:    msg.GroupVersionKind.Kind,
	})

	if err := cl.List(ctx, &list); err != nil {
		return nil, fmt.Errorf("could not get unstructured object: %s", err)
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

func (s *applicationServer) ParseRepoURL(ctx context.Context, msg *pb.ParseRepoURLRequest) (*pb.ParseRepoURLResponse, error) {
	u, err := gitproviders.NewRepoURL(msg.Url, false)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "could not parse url: %s", err.Error())
	}

	return &pb.ParseRepoURLResponse{
		Name:     u.RepositoryName(),
		Owner:    u.Owner(),
		Provider: toProtoProvider(u.Provider()),
	}, nil
}

func (s *applicationServer) GetGitlabAuthURL(ctx context.Context, msg *pb.GetGitlabAuthURLRequest) (*pb.GetGitlabAuthURLResponse, error) {
	u, err := s.glAuthClient.AuthURL(ctx, msg.RedirectUri)
	if err != nil {
		return nil, fmt.Errorf("could not get gitlab auth url: %w", err)
	}

	return &pb.GetGitlabAuthURLResponse{Url: u.String()}, nil
}

func (s *applicationServer) AuthorizeGitlab(ctx context.Context, msg *pb.AuthorizeGitlabRequest) (*pb.AuthorizeGitlabResponse, error) {
	tokenState, err := s.glAuthClient.ExchangeCode(ctx, msg.RedirectUri, msg.Code)
	if err != nil {
		return nil, fmt.Errorf("could not exchange code: %w", err)
	}

	token, err := s.jwtClient.GenerateJWT(tokenState.ExpiresInSeconds, gitproviders.GitProviderGitLab, tokenState.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("could not generate token: %w", err)
	}

	return &pb.AuthorizeGitlabResponse{Token: token}, nil
}

func (s *applicationServer) ValidateProviderToken(ctx context.Context, msg *pb.ValidateProviderTokenRequest) (*pb.ValidateProviderTokenResponse, error) {
	token, err := middleware.ExtractProviderToken(ctx)
	if err != nil {
		return nil, grpcStatus.Error(codes.Unauthenticated, err.Error())
	}

	v, err := findValidator(msg.Provider, s)
	if err != nil {
		return nil, grpcStatus.Error(codes.InvalidArgument, err.Error())
	}

	if err := v.ValidateToken(ctx, token.AccessToken); err != nil {
		return nil, grpcStatus.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.ValidateProviderTokenResponse{
		Valid: true,
	}, nil
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

func mapKustomizationSpecToResponse(kust *kustomizev2.Kustomization) *pb.Kustomization {
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
		return nil, nil, fmt.Errorf("invalid source type %q", st)
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
		deployment = &kustomizev2.Kustomization{}
	}

	if deployment == nil {
		return nil, nil, fmt.Errorf("invalid deployment type %q", at)
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

func toProtoProvider(p gitproviders.GitProviderName) pb.GitProvider {
	switch p {
	case gitproviders.GitProviderGitHub:
		return pb.GitProvider_GitHub
	case gitproviders.GitProviderGitLab:
		return pb.GitProvider_GitLab
	}

	return pb.GitProvider_Unknown
}

func findValidator(provider pb.GitProvider, s *applicationServer) (auth.ProviderTokenValidator, error) {
	switch provider {
	case pb.GitProvider_GitHub:
		return s.ghAuthClient, nil
	case pb.GitProvider_GitLab:
		return s.glAuthClient, nil
	}

	return nil, fmt.Errorf("unknown git provider %s", provider)
}
