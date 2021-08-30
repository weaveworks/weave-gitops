// +build !unittest

package acceptance

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
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
	SetDefaultEventuallyTimeout(EVENTUALLY_DEFAULT_TIME_OUT)
	DEFAULT_SSH_KEY_PATH = os.Getenv("HOME") + "/.ssh/id_rsa"
	GITHUB_ORG = os.Getenv("GITHUB_ORG")
	WEGO_BIN_PATH = os.Getenv("WEGO_BIN_PATH")
	if WEGO_BIN_PATH == "" {
		WEGO_BIN_PATH = "/usr/local/bin/wego"
	}
	log.Infof("WEGO Binary Path: %s", WEGO_BIN_PATH)
})

var _ = AfterSuite(func() {
	if webDriver != nil {
		Expect(webDriver.Stop()).To(Succeed())
	}
})
