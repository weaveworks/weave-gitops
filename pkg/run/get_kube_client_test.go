package run

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
)

var _ = Describe("GetKubeClient", func() {
	var fakeLogger *loggerfakes.FakeLogger

	BeforeEach(func() {
		fakeLogger = &loggerfakes.FakeLogger{}
	})

	It("returns kube client", func() {
		kubeConfigArgs := GetKubeConfigArgs()

		namespace := "test-namespace"

		kubeConfigArgs.Namespace = &namespace

		_, err := kubeConfigArgs.ToRESTConfig()
		Expect(err).NotTo(HaveOccurred())

		kubeClientOpts := GetKubeClientOptions()

		contextName := "test-context"

		kubeClient, err := GetKubeClient(fakeLogger, contextName, k8sEnv.Rest, kubeClientOpts)

		Expect(err).NotTo(HaveOccurred())
		Expect(kubeClient).ToNot(BeNil())
		Expect(kubeClient.ClusterName).To(Equal(contextName))
	})
})
