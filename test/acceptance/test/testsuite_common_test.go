//go:build !unittest
// +build !unittest

package acceptance

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

const (
	gitlabTokenEnvVar       = "GITLAB_TOKEN"
	gitlabOrgEnvVar         = "GITLAB_ORG"
	gitlabPublicGroupEnvVar = "GITLAB_PUBLIC_GROUP"
	gitlabSubgroupEnvVar    = "GITLAB_SUBGROUP"
	gitlabKeyEnvVar         = "GITLAB_KEY"

	githubTokenEnvVar = "GITHUB_TOKEN"
	githubOrgEnvVar   = "GITHUB_ORG"

	gitopsBinaryPathEnvVar = "WEGO_BIN_PATH"
)

func TestAcceptance(t *testing.T) {
	defer func() {
		err := ShowItems("")
		if err != nil {
			log.Infof("Failed to print the cluster resources")
		}

		err = ShowItems("GitRepositories")
		if err != nil {
			log.Infof("Failed to print the GitRepositories")
		}

		ShowWegoControllerLogs(WEGO_DEFAULT_NAMESPACE)
	}()

	if testing.Short() {
		t.Skip("Skip User Acceptance Tests")
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Weave GitOps User Acceptance Tests")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(EVENTUALLY_DEFAULT_TIMEOUT)
	sshKeyPath = getEnvVar("HOME") + "/.ssh/id_rsa"
	gitopsBinaryPath = getEnvVar(gitopsBinaryPathEnvVar)

	githubOrg = getEnvVar(githubOrgEnvVar)
	githubToken = getEnvVar(githubTokenEnvVar)

	gitlabOrg = getEnvVar(gitlabOrgEnvVar)
	gitlabToken = getEnvVar(gitlabTokenEnvVar)
	gitlabKey = getEnvVar(gitlabKeyEnvVar)
	gitlabPublicGroup = getEnvVar(gitlabPublicGroupEnvVar)
	gitlabSubgroup = getEnvVar(gitlabSubgroupEnvVar)

	if gitopsBinaryPath == "" {
		gitopsBinaryPath = "/usr/local/bin/gitops"
	}
	log.Infof("GITOPS Binary Path: %s", gitopsBinaryPath)

	gitProvider, gitOrg, gitProviderName = getGitProviderInfo()

})

func getEnvVar(envVar string) string {
	value := os.Getenv(envVar)
	ExpectWithOffset(1, value).NotTo(BeEmpty(), fmt.Sprintf("Please ensure %s environment variable is set", envVar))

	return value
}
