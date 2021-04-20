package acceptance

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var WEGO_BIN_PATH string

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Suite")
}

var _ = BeforeSuite(func() {
	WEGO_BIN_PATH = os.Getenv("WEGO_BIN_PATH")
	fmt.Printf("WEGO Binary Path: %s", WEGO_BIN_PATH)
	if WEGO_BIN_PATH == "" {
		WEGO_BIN_PATH = "/usr/local/bin/wego"
	}
})
