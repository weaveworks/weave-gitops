package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// BlockingCLIAuthHandler takes over the terminal experience and returns a token when the user completes the flow.
type BlockingCLIAuthHandler func(context.Context, io.Writer) (string, error)

func NewAuthCLIHandler(name gitproviders.GitProviderName) (BlockingCLIAuthHandler, error) {
	switch name {
	case gitproviders.GitProviderGitHub:
		return NewGithubDeviceFlowHandler(http.DefaultClient), nil

	}

	return nil, fmt.Errorf("unsupported auth provider \"%s\"", name)
}

type SecretName struct {
	Name      app.GeneratedSecretName
	Namespace string
}

func (sn SecretName) String() string {
	nn := types.NamespacedName{
		Namespace: sn.Namespace,
		Name:      sn.Name.String(),
	}
	return nn.String()
}

func (sn SecretName) NamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: sn.Namespace,
		Name:      sn.Name.String(),
	}
}

type AuthService interface {
	SetupDeployKey(ctx context.Context, name SecretName, targetName string, repo gitproviders.NormalizedRepoURL) (git.Git, error)
}

type authSvc struct {
	logger     logger.Logger
	fluxClient flux.Flux
	// Note that this is a k8s go-client, NOT a wego kube.Kube interface.
	// That interface wasn't providing any valuable abstraction for this service.
	k8sClient   client.Client
	gitProvider gitproviders.GitProvider
}

// NewAuthService constructs an auth service for doing git operations with an authenticated client.
func NewAuthService(ctx context.Context, fluxClient flux.Flux, k8sClient client.Client, osysClient *osys.OsysClient, providerName gitproviders.GitProviderName, l logger.Logger) (AuthService, string, error) {
	token, err := getGitProviderToken(ctx, osysClient, providerName)
	if err != nil {
		return nil, "", fmt.Errorf("error getting git provider token: %w", err)
	}

	provider, err := gitproviders.New(gitproviders.Config{Provider: providerName, Token: token})
	if err != nil {
		return nil, "", fmt.Errorf("error creating git provider client: %w", err)
	}
	return &authSvc{
		logger:      l,
		fluxClient:  fluxClient,
		k8sClient:   k8sClient,
		gitProvider: provider,
	}, token, nil
}

// SetupDeployKey creates a git.Git client instrumented with existing or generated deploy keys.
// This ensures that git operations are done with stored deploy keys instead of a user's local ssh-agent or equivalent.
func (a *authSvc) SetupDeployKey(ctx context.Context, name SecretName, targetName string, repo gitproviders.NormalizedRepoURL) (git.Git, error) {
	owner := repo.Owner()
	repoName := repo.RepositoryName()
	accountType, err := a.gitProvider.GetAccountType(repo.Owner())
	if err != nil {
		return nil, fmt.Errorf("error getting account type: %w", err)
	}

	repoInfo, err := a.gitProvider.GetRepoInfo(accountType, repo.Owner(), repo.RepositoryName())
	if err != nil {
		return nil, fmt.Errorf("error getting repo info: %w", err)
	}

	if repoInfo.Visibility != nil && *repoInfo.Visibility == gitprovider.RepositoryVisibilityPublic {
		// This is a public repo. We don't need to add deploy keys to it.
		return git.New(nil, wrapper.NewGoGit()), nil
	}

	deployKeyExists, err := a.gitProvider.DeployKeyExists(owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed check for existing deploy key: %w", err)
	}

	if deployKeyExists {
		a.logger.Println("Existing deploy key found")
		// The deploy key was found on the Git Provider, fetch it from the cluster.
		secret, err := a.retrieveDeployKey(ctx, name)
		if apierrors.IsNotFound(err) {
			// Edge case where the deploy key exists on the Git Provider, but not on the cluster.
			// Users might end up here if we uploaded the deploy key, but it failed to save on the cluster,
			// or if a cluster was destroyed during development work.
			// Create and upload a new deploy key.
			a.logger.Warningf("A deploy key named %s was found on the git provider, but not in the cluster.", name.Name)
			return a.provisionDeployKey(ctx, targetName, name, repo)
		} else if err != nil {
			return nil, fmt.Errorf("error retrieving deploy key: %w", err)
		}

		b := extractPrivateKey(secret)

		pubKey, err := makePublicKey(b)
		if err != nil {
			return nil, fmt.Errorf("could not create public key from secret: %w", err)
		}

		// Set the git client to use the existing deploy key.
		return git.New(pubKey, wrapper.NewGoGit()), nil
	}

	return a.provisionDeployKey(ctx, targetName, name, repo)
}

func (a *authSvc) provisionDeployKey(ctx context.Context, targetName string, name SecretName, repo gitproviders.NormalizedRepoURL) (git.Git, error) {
	deployKey, secret, err := a.generateDeployKey(targetName, name, repo)
	if err != nil {
		return nil, fmt.Errorf("error generating deploy key: %w", err)
	}

	publicKeyBytes := extractPublicKey(secret)

	if err := a.gitProvider.UploadDeployKey(repo.Owner(), repo.RepositoryName(), publicKeyBytes); err != nil {
		return nil, fmt.Errorf("error uploading deploy key: %w", err)
	}

	if err := a.storeDeployKey(ctx, secret); err != nil {
		return nil, fmt.Errorf("error storing deploy key: %w", err)
	}

	a.logger.Println("Deploy key generated and uploaded to git provider")

	return git.New(deployKey, wrapper.NewGoGit()), nil
}

// Generates an ssh keypair for upload to the Git Provider and for use in a git.Git client.
func (a *authSvc) generateDeployKey(targetName string, secretName SecretName, repo gitproviders.NormalizedRepoURL) (*ssh.PublicKeys, *corev1.Secret, error) {
	secret, err := a.createKeyPairSecret(targetName, secretName, repo)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create key-pair secret: %w", err)
	}

	privKeyBytes := extractPrivateKey(secret)

	deployKey, err := makePublicKey(privKeyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating new public keys: %w", err)
	}

	return deployKey, secret, nil

}

// Wrapper to abstract how the key is stored in case we want to change this later.
func (a *authSvc) storeDeployKey(ctx context.Context, secret *corev1.Secret) error {
	if err := a.k8sClient.Create(ctx, secret); err != nil {
		return fmt.Errorf("could not store secret: %w", err)
	}

	return nil
}

// Wrapper to abstract how a key is fetched.
func (a *authSvc) retrieveDeployKey(ctx context.Context, name SecretName) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	if err := a.k8sClient.Get(ctx, name.NamespacedName(), secret); err != nil {
		return nil, fmt.Errorf("error getting deploy key secret: %w", err)
	}

	return secret, nil
}

// Uses flux to create a ssh key pair secret.
func (a *authSvc) createKeyPairSecret(targetName string, name SecretName, repo gitproviders.NormalizedRepoURL) (*corev1.Secret, error) {
	secretData, err := a.fluxClient.CreateSecretGit(name.Name.String(), repo.String(), name.Namespace)
	if err != nil {
		return nil, fmt.Errorf("could not create git secret: %w", err)
	}

	var secret corev1.Secret
	err = yaml.Unmarshal(secretData, &secret)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal created secret: %w", err)
	}

	return &secret, nil
}

func makePublicKey(pemBytes []byte) (*ssh.PublicKeys, error) {
	return ssh.NewPublicKeys("git", pemBytes, "")
}

// Helper to standardize how we extract data from a ssh key pair secret.
func extractSecretPart(secret *corev1.Secret, key string) []byte {
	var data []byte
	var ok bool
	data, ok = secret.Data[string(key)]
	if !ok {
		// StringData is a write-only field, flux generates secrets on disk with StringData
		// Once they get applied on the cluster, Kubernetes populates Data and removes StringData.
		// Handle this case here to be able to extract data no matter the "state" of the object.
		data = []byte(secret.StringData[string(key)])
	}

	return data
}

func extractPrivateKey(secret *corev1.Secret) []byte {
	return extractSecretPart(secret, "identity")
}

func extractPublicKey(secret *corev1.Secret) []byte {
	return extractSecretPart(secret, "identity.pub")
}

func getGitProviderToken(ctx context.Context, osysClient *osys.OsysClient, providerName gitproviders.GitProviderName) (string, error) {

	token, tokenErr := osysClient.GetGitProviderToken()

	if tokenErr == osys.ErrNoGitProviderTokenSet {
		// No provider token set, we need to do the auth flow.
		authHandler, err := NewAuthCLIHandler(providerName)
		if err != nil {
			return "", fmt.Errorf("could not get auth handler for provider %s: %w", providerName, err)
		}

		token, err = authHandler(ctx, osysClient.Stdout())
		if err != nil {
			return "", fmt.Errorf("could not complete auth flow: %w", err)
		}
	} else if tokenErr != nil {
		// We didn't detect a NoGitProviderSet error, something else went wrong.
		return "", fmt.Errorf("could not get access token: %w", tokenErr)
	}
	return token, nil
}
