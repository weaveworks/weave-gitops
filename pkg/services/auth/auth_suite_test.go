package auth

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var k8sClient client.Client

func TestGitProviderAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Auth Suite")
}

var cleanupK8s func()

var _ = BeforeSuite(func() {
	k, stop, err := testutils.StartK8sTestEnvironment()
	Expect(err).NotTo(HaveOccurred())
	cleanupK8s = stop
	k8sClient = k
})

var _ = AfterSuite(func() {
	cleanupK8s()
})
