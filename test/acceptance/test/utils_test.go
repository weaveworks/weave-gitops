package acceptance

import (
	"fmt"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestFileExists(t *testing.T) {
	require.True(t, FileExists("utils_test.go"))
	require.False(t, FileExists("imaginaryfile.txt"))
}

// Resetting namespace is an expensive operation, only use this when absolutely necessary
func ResetNamespace(namespace string) {
	log.Infof("Resetting namespace for controllers...")

	By("And there's no previous wego installation", func() {
		command := exec.Command("sh", "-c", fmt.Sprintf("%s install --namespace %s| kubectl --ignore-not-found=true delete -f -", WEGO_BIN_PATH, namespace))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, INSTALL_RESET_TIMEOUT).Should(gexec.Exit())
	})
}
