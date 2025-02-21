package kube_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

var (
	k8sClient  client.Client
	k8sTestEnv *testutils.K8sTestEnv
)

func TestKube(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kube Suite")
}

var cleanupK8s func()

var _ = BeforeSuite(func() {
	var err error
	k8sTestEnv, err = testutils.StartK8sTestEnvironment([]string{
		"../../tools/testcrds",
	})
	Expect(err).NotTo(HaveOccurred())

	cleanupK8s = k8sTestEnv.Stop
	k8sClient = k8sTestEnv.Client
})

var _ = AfterSuite(func() {
	cleanupK8s()
})
