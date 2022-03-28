package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/internal"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
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
	case gitproviders.GitProviderGitLab:
		authFlow, err := NewGitlabAuthFlow(internal.GitlabRedirectUriCLI, http.DefaultClient)
		if err != nil {
			return nil, fmt.Errorf("could not create gitlab auth flow for CLI: %w", err)
		}

		return NewGitlabAuthFlowHandler(http.DefaultClient, authFlow), nil
	}

	return nil, fmt.Errorf("unsupported auth provider \"%s\"", name)
}

type ProviderTokenValidator interface {
	ValidateToken(ctx context.Context, token string) error
}

type SecretName struct {
	Name      models.GeneratedSecretName
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
	CreateGitClient(ctx context.Context, repoUrl gitproviders.RepoURL, namespace string, dryRun bool) (git.Git, error)
	GetGitProvider() gitproviders.GitProvider
	SetupDeployKey(ctx context.Context, namespace string, repo gitproviders.RepoURL) (*ssh.PublicKeys, error)
}

type authSvc struct {
	log        logger.Logger
	fluxClient flux.Flux
	// Note that this is a k8s go-client, NOT a wego kube.Kube interface.
	// That interface wasn't providing any valuable abstraction for this service.
	k8sClient   client.Client
	gitProvider gitproviders.GitProvider
}

// NewAuthService constructs an auth service for doing git operations with an authenticated client.
func NewAuthService(fluxClient flux.Flux, k8sClient client.Client, provider gitproviders.GitProvider, log logger.Logger) (AuthService, error) {
	return &authSvc{
		log:         log,
		fluxClient:  fluxClient,
		k8sClient:   k8sClient,
		gitProvider: provider,
	}, nil
}

// GetGitProvider returns the GitProvider associated with the AuthService instance
func (a *authSvc) GetGitProvider() gitproviders.GitProvider {
	return a.gitProvider
}

// CreateGitClient creates a git.Git client instrumented with existing or generated deploy keys.
// This ensures that git operations are done with stored deploy keys instead of a user's local ssh-agent or equivalent.
func (a *authSvc) CreateGitClient(ctx context.Context, repoUrl gitproviders.RepoURL, namespace string, dryRun bool) (git.Git, error) {
	if dryRun {
		d, _ := makePublicKey([]byte(""))
		return git.New(d, wrapper.NewGoGit()), nil
	}

	pubKey, keyErr := a.SetupDeployKey(ctx, namespace, repoUrl)
	if keyErr != nil {
		return nil, fmt.Errorf("error setting up deploy keys: %w", keyErr)
	}

	if pubKey == nil {
		// Don't return git.New(pubkey, wrapper.NewGoGit()), nil here. It will fail
		// "nil" of type *ssh.PublicKeys does not behave correctly
		return git.New(nil, wrapper.NewGoGit()), nil
	}

	// Set the git client to use the existing deploy key.
	return git.New(pubKey, wrapper.NewGoGit()), nil
}

// SetupDeployKey creates a git.Git client instrumented with existing or generated deploy keys.
// This ensures that git operations are done with stored deploy keys instead of a user's local ssh-agent or equivalent.
func (a *authSvc) SetupDeployKey(ctx context.Context, namespace string, repo gitproviders.RepoURL) (*ssh.PublicKeys, error) {
	secretName := SecretName{
		Name:      models.CreateRepoSecretName(repo),
		Namespace: namespace,
	}

	deployKeyExists, err := a.gitProvider.DeployKeyExists(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("failed check for existing deploy key: %w", err)
	}

	if deployKeyExists {
		// The deploy key was found on the Git Provider, fetch it from the cluster.
		secret, err := a.retrieveDeployKey(ctx, secretName)
		if apierrors.IsNotFound(err) {
			// Edge case where the deploy key exists on the Git Provider, but not on the cluster.
			// Users might end up here if we uploaded the deploy key, but it failed to save on the cluster,
			// or if a cluster was destroyed during development work.
			// Create and upload a new deploy key.
			a.log.Warningf("A deploy key named %s was found on the git provider, but not in the cluster.", secretName.Name)
			return a.provisionDeployKey(ctx, secretName, repo)
		} else if err != nil {
			return nil, fmt.Errorf("error retrieving deploy key: %w", err)
		}

		b := extractPrivateKey(secret)

		pubKey, err := makePublicKey(b)
		if err != nil {
			return nil, fmt.Errorf("could not create public key from secret: %w", err)
		}

		// Set the git client to use the existing deploy key.
		return pubKey, nil
	}

	return a.provisionDeployKey(ctx, secretName, repo)
}

func (a *authSvc) provisionDeployKey(ctx context.Context, name SecretName, repo gitproviders.RepoURL) (*ssh.PublicKeys, error) {
	deployKey, secret, err := a.generateDeployKey(name, repo)
	if err != nil {
		return nil, fmt.Errorf("error generating deploy key: %w", err)
	}

	publicKeyBytes := extractPublicKey(secret)

	if err := a.gitProvider.UploadDeployKey(ctx, repo, publicKeyBytes); err != nil {
		return nil, fmt.Errorf("error uploading deploy key: %w", err)
	}

	if err := a.storeDeployKey(ctx, secret); err != nil {
		return nil, fmt.Errorf("error storing deploy key: %w", err)
	}

	a.log.Println("Deploy key generated and uploaded to git provider")

	return deployKey, nil
}

// Generates an ssh keypair for upload to the Git Provider and for use in a git.Git client.
func (a *authSvc) generateDeployKey(secretName SecretName, repo gitproviders.RepoURL) (*ssh.PublicKeys, *corev1.Secret, error) {
	secret, err := a.createKeyPairSecret(secretName, repo)
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
func (a *authSvc) createKeyPairSecret(name SecretName, repo gitproviders.RepoURL) (*corev1.Secret, error) {
	secretData, err := a.fluxClient.CreateSecretGit(name.Name.String(), repo, name.Namespace)
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
