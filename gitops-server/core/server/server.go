package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/gitops-server/core/cache"
	"github.com/weaveworks/weave-gitops/gitops-server/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/gitops-server/core/nsaccess"
	pb "github.com/weaveworks/weave-gitops/gitops-server/pkg/api/core"
	"github.com/weaveworks/weave-gitops/gitops-server/pkg/kube"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Hydrate(ctx context.Context, mux *runtime.ServeMux, cfg CoreServerConfig) error {
	appsServer, err := NewCoreServer(cfg)
	if err != nil {
		return fmt.Errorf("unable to create new kube client: %w", err)
	}

	if err = pb.RegisterCoreHandlerServer(ctx, mux, appsServer); err != nil {
		return fmt.Errorf("could not register new app server: %w", err)
	}

	return nil
}

const temporarilyEmptyAppName = ""

type ClientGetterFn func(ctx context.Context) clustersmngr.Client
type coreServer struct {
	pb.UnimplementedCoreServer

	k8s            kube.ClientGetter
	cacheContainer cache.Container
	logger         logr.Logger
	nsChecker      nsaccess.Checker
}

type CoreServerConfig struct {
	log            logr.Logger
	RestCfg        *rest.Config
	clusterName    string
	NSAccess       nsaccess.Checker
	CacheContainer cache.Container
}

func NewCoreConfig(log logr.Logger, cfg *rest.Config, cacheContainer cache.Container, clusterName string) CoreServerConfig {
	return CoreServerConfig{
		log:            log.WithName("core-server"),
		RestCfg:        cfg,
		clusterName:    clusterName,
		NSAccess:       nsaccess.NewChecker(nsaccess.DefautltWegoAppRules),
		CacheContainer: cacheContainer,
	}
}

func NewCoreServer(cfg CoreServerConfig, setters ...CoreOption) (pb.CoreServer, error) {
	ctx := context.Background()
	cfgGetter := kube.NewImpersonatingConfigGetter(cfg.RestCfg, false)

	cfg.CacheContainer.Start(ctx)

	clientGetter := kube.NewDefaultClientGetter(cfgGetter, cfg.clusterName)

	args := &CoreOptions{
		ClientGetter: clientGetter,
	}

	for _, setter := range setters {
		setter(args)
	}

	return &coreServer{
		k8s:            args.ClientGetter,
		logger:         cfg.log,
		cacheContainer: cfg.CacheContainer,
		nsChecker:      cfg.NSAccess,
	}, nil
}

func list(ctx context.Context, k8s client.Client, appName, namespace string, list client.ObjectList, extraOpts ...client.ListOption) error {
	opts := []client.ListOption{
		getMatchingLabels(appName),
		client.InNamespace(namespace),
	}

	opts = append(opts, extraOpts...)
	err := k8s.List(ctx, list, opts...)
	err = wrapK8sAPIError("list resource", err)

	return err
}

func wrapK8sAPIError(msg string, err error) error {
	if k8serrors.IsUnauthorized(err) {
		return status.Errorf(codes.PermissionDenied, err.Error())
	} else if k8serrors.IsNotFound(err) {
		return status.Errorf(codes.NotFound, err.Error())
	} else if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	return nil
}

func doClientError(err error) error {
	return status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
}

func (cs *coreServer) namespaces() (map[string][]v1.Namespace, error) {
	objs, err := cs.cacheContainer.List(cache.NamespaceStorage)
	if err != nil {
		return nil, err
	}

	namespaces, ok := objs.(map[string][]v1.Namespace)
	if !ok {
		return nil, errors.New("could not convert objects to map[string][]v1.Namespace")
	}

	return namespaces, nil
}
