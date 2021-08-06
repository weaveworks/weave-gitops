package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/utils"
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

type AuthService struct {
	logger     logger.Logger
	fluxClient flux.Flux
	// Note that this is a k8s go-client, NOT a wego kube.Kube interface.
	// That interface wasn't providing any valuable abstraction for this service.
	k8sClient   client.Client
	gitProvider gitproviders.GitProvider
	gitClient   git.Git
}

// NewAuthService constructs an auth service for doing git and GitProvider operations with authenticated clients.
func NewAuthService(fluxClient flux.Flux, k8sClient client.Client, providerName gitproviders.GitProviderName, l logger.Logger, token string) (*AuthService, error) {
	provider, err := gitproviders.New(gitproviders.Config{Provider: providerName, Token: token})
	if err != nil {
		return nil, fmt.Errorf("error creating git provider client: %w", err)
	}
	return &AuthService{
		logger:      l,
		fluxClient:  fluxClient,
		k8sClient:   k8sClient,
		gitProvider: provider,
		// We do not have a gitClient yet. We need to get or create deploy keys first.
		// This will get populated by the SetupDeployKeys call.
		// Listing it here for the sake of readability (it would still be nil without this entry).
		gitClient: nil,
	}, nil
}

// SetupDeployKey instruments the AuthService with the neccesary data to provide a git client that uses deploy keys.
// The user MUST call this function before calling .GitClient(), else an error will be returned.
// The work of setting up deploy keys was NOT done in the constructor for the sake being explicit about what is happening.
func (a *AuthService) SetupDeployKey(ctx context.Context, name types.NamespacedName, targetName string, repoUrl string) error {
	repoUrl = utils.SanitizeRepoUrl(repoUrl)

	owner, err := utils.GetOwnerFromUrl(repoUrl)
	if err != nil {
		return fmt.Errorf("error getting owner from URL: %w", err)
	}

	repoName := utils.UrlToRepoName(repoUrl)

	accountType, err := a.gitProvider.GetAccountType(owner)
	if err != nil {
		return fmt.Errorf("error getting account type: %w", err)
	}

	repoInfo, err := a.gitProvider.GetRepoInfo(accountType, owner, repoName)
	if err != nil {
		return fmt.Errorf("error getting repo info: %w", err)
	}

	if repoInfo.Visibility != nil && *repoInfo.Visibility == gitprovider.RepositoryVisibilityPublic {
		// This is a public repo. We don't need to add deploy keys to it.
		a.gitClient = git.New(nil)
		return nil
	}

	deployKeyExists, err := a.gitProvider.DeployKeyExists(owner, repoName)
	if err != nil {
		return fmt.Errorf("failed check for existing deploy key: %w", err)
	}

	if deployKeyExists {
		a.logger.Println("Existing deploy key found")
		secretName := types.NamespacedName{
			Name:      utils.CreateAppSecretName(targetName, repoUrl),
			Namespace: name.Namespace,
		}
		// The deploy key was found on the Git Provider, fetch it from the cluster.
		secret, err := a.retreiveDeployKey(ctx, secretName)
		if apierrors.IsNotFound(err) {
			// Edge case where the deploy key exists on the Git Provider, but not on the cluster.
			// Users might end up here if we uploaded the deploy key, but it failed to save on the cluster.
			// TODO: What should we do to help the user out here? We can't read the key data from the Git Provider,
			// do we delete and recreate?
			return fmt.Errorf("deploy key does not exist on cluster: %w", err)
		} else if err != nil {
			return fmt.Errorf("error retrieving deploy key: %w", err)
		}

		b := extractPrivateKey(secret)

		pubKey, err := makePublicKey(b)
		if err != nil {
			return fmt.Errorf("could not create public key from secret: %w", err)
		}

		// Set the git client to use the existing deploy key.
		a.gitClient = git.New(pubKey)

		return nil
	}

	deployKey, secret, err := a.generateDeployKey(targetName, name, repoUrl)
	if err != nil {
		return fmt.Errorf("error generating deploy key: %w", err)
	}

	publicKeyBytes := extractPublicKey(secret)

	if err := a.gitProvider.UploadDeployKey(owner, repoName, publicKeyBytes); err != nil {
		return fmt.Errorf("error uploading deploy key: %w", err)
	}

	if err := a.storeDeployKey(ctx, secret); err != nil {
		return fmt.Errorf("error storing deploy key: %w", err)
	}

	a.logger.Println("Deploy key generated and uploaded to git provider")
	// Set the gitClient to use the newly generated deploy key
	a.gitClient = git.New(deployKey)

	return nil

}

var ErrNoDeployKeysSetup = errors.New("the git client has not been intialized. Run SetupDeployKeys() on this object first.")

// GitClient returns an instrumented git client that will use the deploy key any git operations.
// SetupDeployKeys MUST be called before this, else an error will be returned.
func (a *AuthService) GitClient() (git.Git, error) {
	if a.gitClient == nil {
		return nil, ErrNoDeployKeysSetup
	}

	return a.gitClient, nil
}

// GitProvider returns a GitProvider that is already constructed with access tokens.
// This can be used for GitProvider API operations, like opening pull requests or adding deploy keys.
func (a *AuthService) GitProvider() gitproviders.GitProvider {
	return a.gitProvider
}

// Generates an ssh keypair for upload to the Git Provider and for use in a git.Git client.
func (a *AuthService) generateDeployKey(targetName string, secretName types.NamespacedName, repoUrl string) (*ssh.PublicKeys, *corev1.Secret, error) {
	secret, err := a.createKeyPairSecret(targetName, secretName, repoUrl)
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
func (a *AuthService) storeDeployKey(ctx context.Context, secret *corev1.Secret) error {
	if err := a.k8sClient.Create(ctx, secret); err != nil {
		return fmt.Errorf("could not store secret: %w", err)
	}

	return nil
}

// Wrapper to abstract how a key is fetched.
func (a *AuthService) retreiveDeployKey(ctx context.Context, name types.NamespacedName) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	if err := a.k8sClient.Get(ctx, name, secret); err != nil {
		return nil, fmt.Errorf("error getting deploy key secret: %w", err)
	}

	return secret, nil
}

// Uses flux to create a ssh key pair secret.
func (a *AuthService) createKeyPairSecret(targetName string, name types.NamespacedName, repoUrl string) (*corev1.Secret, error) {
	secretname := utils.CreateAppSecretName(targetName, repoUrl)
	secretData, err := a.fluxClient.CreateSecretGit(secretname, repoUrl, name.Namespace)
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
