// +build !unittest

package acceptance

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var WEGO_BIN_PATH string

func TestAcceptance(t *testing.T) {

	if testing.Short() {
		t.Skip("Skip User Acceptance Tests")
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(EVENTUALLY_DEFAULT_TIME_OUT)
	WEGO_BIN_PATH = os.Getenv("WEGO_BIN_PATH")
	if WEGO_BIN_PATH == "" {
		WEGO_BIN_PATH = "/usr/local/bin/wego"
	}
	log.Infof("WEGO Binary Path: %s", WEGO_BIN_PATH)
})
