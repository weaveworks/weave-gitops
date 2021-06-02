// +build !unittest

package acceptance

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	ginkgo "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func GomegaFail(message string, callerSkip ...int) {
	//Show all resources
	err := ShowItems("")
	if err != nil {
		log.Infof("Failed to print the cluster resources")
	}

	err = ShowItems("GitRepositories")
	if err != nil {
		log.Infof("Failed to print the GitRepositories")
	}

	ShowWegoControllerLogs(WEGO_DEFAULT_NAMESPACE)

	//Pass this down to the default handler for onward processing
	ginkgo.Fail(message, callerSkip...)
}

func TestAcceptance(t *testing.T) {

	if testing.Short() {
		t.Skip("Skip User Acceptance Tests")
	}

	RegisterFailHandler(GomegaFail)
	RunSpecs(t, getSuiteTitle())
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(EVENTUALLY_DEFAULT_TIME_OUT)
	WEGO_BIN_PATH = os.Getenv("WEGO_BIN_PATH")
	if WEGO_BIN_PATH == "" {
		WEGO_BIN_PATH = "/usr/local/bin/wego"
	}
	log.Infof("WEGO Binary Path: %s", WEGO_BIN_PATH)
})
