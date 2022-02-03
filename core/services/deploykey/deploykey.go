package deploykey

import (
	"context"
	"fmt"
	"net/http"

	"github.com/weaveworks/weave-gitops/core/adapters/fluxbin"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Manager interface {
	Create(ctx context.Context, name types.NamespacedName, repoURL gitproviders.RepoURL, token string) (models.GeneratedSecretName, error)
}

func NewManager(k8s client.Client, rt http.RoundTripper) Manager {
	return manager{k8s: k8s, http: rt}
}

type manager struct {
	k8s  client.Client
	http http.RoundTripper
}

func (m manager) Create(ctx context.Context, name types.NamespacedName, repoURL gitproviders.RepoURL, token string) (models.GeneratedSecretName, error) {
	generatedName := models.GeneratedSecretName(name.Name)

	secret, err := fluxbin.CreateSecretGit(string(generatedName), name.Namespace, repoURL)
	if err != nil {
		return "", fmt.Errorf("creating git secret: %w", err)
	}

	p, err := gitproviders.New(gitproviders.Config{
		Provider:     repoURL.Provider(),
		Token:        token,
		Hostname:     repoURL.URL().Host,
		RoundTripper: m.http,
	}, repoURL.Owner(), gitproviders.GetAccountType)
	if err != nil {
		return generatedName, fmt.Errorf("creating git provider: %w", err)
	}

	b := auth.ExtractPublicKey(&secret)

	if err := p.UploadDeployKey(ctx, repoURL, b); err != nil {
		return generatedName, fmt.Errorf("uploading deploy key: %w", err)
	}

	if err := m.k8s.Create(ctx, &secret); err != nil {
		return generatedName, fmt.Errorf("adding secret to cluster: %w", err)
	}

	return generatedName, nil
}
