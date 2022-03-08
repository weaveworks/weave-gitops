package kube_test

import (
	"math/rand"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	testClustername = "test-cluster"
	k8sClient       client.Client
	k               kube.Kube
	k8sTestEnv      *testutils.K8sTestEnv
)

func TestKube(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kube Suite")
}

var cleanupK8s func()

var _ = BeforeSuite(func() {
	var err error
	k8sTestEnv, err = testutils.StartK8sTestEnvironment([]string{
		"../../manifests/crds",
		"../../tools/testcrds",
	})
	Expect(err).NotTo(HaveOccurred())

	cleanupK8s = k8sTestEnv.Stop
	k8sClient = k8sTestEnv.Client
})

var _ = AfterSuite(func() {
	cleanupK8s()
})

func init() {
	rand.Seed(time.Now().UnixNano())
}
