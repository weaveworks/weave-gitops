package cluster

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	machnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/cli-utils/pkg/flowcontrol"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const (
	// DefaultCluster name
	DefaultCluster = "Default"
	// ClientQPS is the QPS to use while creating the k8s clients (actually a float32)
	ClientQPS = 1000
	// ClientBurst is the burst to use while creating the k8s clients
	ClientBurst = 2000

	kubeClientDialTimeout   = 5 * time.Second
	kubeClientDialKeepAlive = 30 * time.Second
)

var (
	kubeClientTimeout = getEnvDuration("WEAVE_GITOPS_KUBE_CLIENT_TIMEOUT", 30*time.Second)
	// DefaultKubeConfigOptions = []KubeConfigOption{WithFlowControl}
	DefaultKubeConfigOptions = []KubeConfigOption{}
)

type KubeConfigOption func(*rest.Config) (*rest.Config, error)

// Cluster is an abstraction around a connection to a specific k8s cluster
// It's effectively a pair of (name, rest.Config), with some helpers
//
//counterfeiter:generate . Cluster
type Cluster interface {
	// GetName gets the name weave-gitops has given this cluster.
	GetName() string
	// GetHost gets the host of the cluster - this should match what's set in the `rest.Config`s below
	GetHost() string
	// GetServerClient gets a un-impersonated client for this cluster
	GetServerClient() (client.Client, error)
	// GetUserClient gets an appropriately impersonated client for the user on this cluster
	GetUserClient(*auth.UserPrincipal) (client.Client, error)
	// GetServerClientset gets an un-impersonated clientset for this cluster
	GetServerClientset() (kubernetes.Interface, error)
	// GetUserClientset gets an appropriately impersonated clientset for the user on this cluster
	GetUserClientset(*auth.UserPrincipal) (kubernetes.Interface, error)
	// GetServerConig gets an un-impersonated rest config for this cluster.
	// Note that using this might bypass e.g. caches.
	GetServerConfig() (*rest.Config, error)
}

func WithFlowControl(config *rest.Config) (*rest.Config, error) {
	// flowcontrol.IsEnabled makes a request to the K8s API of the cluster stored in the config.
	// It does a HEAD request to /livez/ping which uses the config.Dial timeout. We can use this
	// function to error early rather than wait to call client.New.
	enabled, err := flowcontrol.IsEnabled(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("error querying cluster for flowcontrol config: %w", err)
	}

	if enabled {
		// Enabled & negative QPS and Burst indicate that the client would use the rate limit set by the server.
		// Ref: https://github.com/kubernetes/kubernetes/blob/v1.24.0/staging/src/k8s.io/client-go/rest/config.go#L354-L364
		config.QPS = -1
		config.Burst = -1

		return config, nil
	}

	config.QPS = ClientQPS
	config.Burst = ClientBurst

	// From https://github.com/weaveworks/weave-gitops-enterprise/issues/3189
	// Suggested in https://github.com/kubernetes/kubernetes/issues/118703#issuecomment-1595072383
	// TODO: Revert or adapt when upstream fix is available
	config.Proxy = machnet.NewProxierWithNoProxyCIDR(http.ProxyFromEnvironment)

	return config, nil
}

func getEnvDuration(key string, defaultDuration time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultDuration
	}

	d, err := time.ParseDuration(val)

	// on error return the default duration
	if err != nil {
		return defaultDuration
	}

	return d
}
