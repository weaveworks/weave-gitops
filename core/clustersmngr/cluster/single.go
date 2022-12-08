package cluster

import (
	"fmt"
	"net"

	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type singleCluster struct {
	name       string
	restConfig *rest.Config
	scheme     *apiruntime.Scheme
}

// NewSingleCluster creates and returns a Cluster that uses the provided
// rest.Config.
func NewSingleCluster(name string, config *rest.Config, scheme *apiruntime.Scheme, restConfigOptions ...RESTConfigOption) (Cluster, error) {
	// TODO: why does the cluster care about options?
	config.Timeout = kubeClientTimeout
	config.Dial = (&net.Dialer{
		Timeout: kubeClientDialTimeout,
		// KeepAlive is default to 30s within client-go.
		KeepAlive: kubeClientDialKeepAlive,
	}).DialContext

	var err error

	for _, opt := range restConfigOptions {
		config, err = opt(config)
		if err != nil {
			return nil, err
		}
	}

	return &singleCluster{
		name:       name,
		restConfig: config,
		scheme:     scheme,
	}, nil
}

func (c *singleCluster) GetName() string {
	return c.name
}

func (c *singleCluster) GetHost() string {
	return c.restConfig.Host
}

func (c *singleCluster) GetUserClient(user *auth.UserPrincipal) (client.Client, error) {
	cfg, err := getImpersonatedConfig(c.restConfig, user)
	if err != nil {
		return nil, err
	}

	client, err := getClientFromConfig(cfg, c.scheme)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *singleCluster) GetServerClient(opts ...RESTConfigOption) (client.Client, error) {
	cfg, err := applyOptsToRESTConfig(c.restConfig, opts...)
	if err != nil {
		return nil, err
	}
	client, err := getClientFromConfig(cfg, c.scheme)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *singleCluster) GetUserClientset(user *auth.UserPrincipal) (kubernetes.Interface, error) {
	cfg, err := getImpersonatedConfig(c.restConfig, user)
	if err != nil {
		return nil, err
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("making clientset: %w", err)
	}

	return cs, nil
}

func (c *singleCluster) GetServerClientset() (kubernetes.Interface, error) {
	cs, err := kubernetes.NewForConfig(c.restConfig)
	if err != nil {
		return nil, fmt.Errorf("making clientset: %w", err)
	}

	return cs, nil
}

func (c *singleCluster) GetServerConfig() (*rest.Config, error) {
	return c.restConfig, nil
}

func getClientFromConfig(config *rest.Config, scheme *apiruntime.Scheme) (client.Client, error) {
	mapper, err := apiutil.NewDiscoveryRESTMapper(config)
	if err != nil {
		return nil, fmt.Errorf("could not create RESTMapper from config: %w", err)
	}

	client, err := client.New(config, client.Options{
		Scheme: scheme,
		Mapper: mapper,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create leaf client: %w", err)
	}

	return client, nil
}

func getImpersonatedConfig(config *rest.Config, user *auth.UserPrincipal) (*rest.Config, error) {
	cfg := rest.CopyConfig(config)

	if !user.Valid() {
		return nil, fmt.Errorf("no user ID or Token found in UserPrincipal")
	} else if tok := user.Token(); tok != "" {
		cfg.BearerToken = tok
	} else {
		cfg.Impersonate = rest.ImpersonationConfig{
			UserName: user.ID,
			Groups:   user.Groups,
		}
	}

	return cfg, nil
}

// applyOptsToRESTConfig applies the provided options to a copy of the provided
// rest.Config struct.
func applyOptsToRESTConfig(cfg *rest.Config, opts ...RESTConfigOption) (*rest.Config, error) {
	updated := rest.CopyConfig(cfg)
	for _, opt := range opts {
		changed, err := opt(updated)
		if err != nil {
			return nil, err
		}
		updated = changed
	}

	return updated, nil
}
