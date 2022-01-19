package profiles

import (
	"context"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"k8s.io/client-go/kubernetes"
)

const (
	wegoServiceName = "wego-app"
	getProfilesPath = "/v1/profiles"
)

type ProfilesService interface {
	// Add installs a profile on a cluster
	Add(ctx context.Context, gitProvider gitproviders.GitProvider, opts AddOptions) error
	// Get lists all the available profiles in a cluster
	Get(ctx context.Context, opts GetOptions) error
}

type ProfilesSvc struct {
	ClientSet kubernetes.Interface
}

func NewService(clientSet kubernetes.Interface) *ProfilesSvc {
	return &ProfilesSvc{
		ClientSet: clientSet,
	}
}
