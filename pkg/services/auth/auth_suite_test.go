package auth

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var k8sClient client.Client

func TestGitProviderAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Auth Suite")
}

var _ = BeforeSuite(func() {
	k, err := utils.StartK8sTestEnvironment()
	Expect(err).NotTo(HaveOccurred())
	k8sClient = k
})
