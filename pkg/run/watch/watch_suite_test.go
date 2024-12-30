package watch

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

var (
	k8sClient client.Client
	k8sEnv    *testutils.K8sTestEnv
)

func TestRun(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Run Suite")
}

var cleanupK8s func()

var _ = BeforeSuite(func() {
	var err error
	k8sEnv, err = testutils.StartK8sTestEnvironment([]string{
		"../../../tools/testcrds",
	})
	Expect(err).NotTo(HaveOccurred())

	cleanupK8s = k8sEnv.Stop
	k8sClient = k8sEnv.Client
})

var _ = AfterSuite(func() {
	cleanupK8s()
})
