package profiles

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const (
	// ManifestFileName contains the manifests of all installed Profiles
	ManifestFileName = "profiles.yaml"

	wegoServiceName = "wego-app"
	getProfilesPath = "/v1/profiles"
)

type ProfilesService interface {
	// Add installs a profile on a cluster
	Add(ctx context.Context, gitProvider gitproviders.GitProvider, opts AddOptions) error
	// Get lists all the available profiles in a cluster
	Get(ctx context.Context, opts GetOptions) error
	// Update updates a profile
	Update(ctx context.Context, gitProvider gitproviders.GitProvider, opts UpdateOptions) error
}

type ProfilesSvc struct {
	ClientSet kubernetes.Interface
	Logger    logger.Logger
}

func NewService(clientSet kubernetes.Interface, log logger.Logger) *ProfilesSvc {
	return &ProfilesSvc{
		ClientSet: clientSet,
		Logger:    log,
	}
}

func (s *ProfilesSvc) discoverHelmRepository(ctx context.Context, opts GetOptions) (types.NamespacedName, string, error) {
	availableProfile, version, err := s.GetProfile(ctx, opts)
	if err != nil {
		return types.NamespacedName{}, "", fmt.Errorf("failed to get profiles from cluster: %w", err)
	}

	return types.NamespacedName{
		Name:      availableProfile.HelmRepository.Name,
		Namespace: availableProfile.HelmRepository.Namespace,
	}, version, nil
}
