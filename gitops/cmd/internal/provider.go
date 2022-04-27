package internal

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/gitops/pkg/logger"
)

const missingTokenErr = "the %q environment variable needs to be set to a valid token"

type gitProviderClient struct {
	stdout        *os.File
	lookupEnvFunc func(key string) (string, bool)
	log           logger.Logger
}

func NewGitProviderClient(stdout *os.File, lookupEnvFunc func(key string) (string, bool), log logger.Logger) gitproviders.Client {
	return &gitProviderClient{
		stdout:        stdout,
		lookupEnvFunc: lookupEnvFunc,
		log:           log,
	}
}

// GetProvider returns a GitProvider containing the token stored in the <git provider>_TOKEN
func (c *gitProviderClient) GetProvider(repoUrl gitproviders.RepoURL, getAccountType gitproviders.AccountTypeGetter) (gitproviders.GitProvider, error) {
	token, err := GetToken(repoUrl, c.lookupEnvFunc)
	if err != nil {
		return nil, err
	}

	provider, err := gitproviders.New(gitproviders.Config{
		Provider: repoUrl.Provider(),
		Token:    token,
		Hostname: repoUrl.URL().Host,
	}, repoUrl.Owner(), getAccountType)
	if err != nil {
		return nil, fmt.Errorf("error creating git provider client: %w", err)
	}

	return provider, nil
}

func getTokenVarName(providerName gitproviders.GitProviderName) (string, error) {
	switch providerName {
	case gitproviders.GitProviderGitHub:
		return "GITHUB_TOKEN", nil
	case gitproviders.GitProviderGitLab:
		return "GITLAB_TOKEN", nil
	default:
		return "", fmt.Errorf("unknown git provider: %q", providerName)
	}
}

// GetToken returns the token stored in the <git provider>_TOKEN env var
func GetToken(repoUrl gitproviders.RepoURL, lookupEnvFunc func(key string) (string, bool)) (string, error) {
	tokenVarName, err := getTokenVarName(repoUrl.Provider())
	if err != nil {
		return "", fmt.Errorf("could not determine git provider token name: %w", err)
	}

	token, exists := lookupEnvFunc(tokenVarName)
	if !exists {
		return "", fmt.Errorf(missingTokenErr, tokenVarName)
	}

	return token, nil
}
