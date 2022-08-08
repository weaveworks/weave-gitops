package run_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/run"
)

var _ = Describe("GetKubeClient", func() {
	var fakeLogger *loggerfakes.FakeLogger

	BeforeEach(func() {
		fakeLogger = &loggerfakes.FakeLogger{}
	})

	It("returns kube client", func() {
		kubeConfigArgs := run.GetKubeConfigArgs()

		namespace := "test-namespace"

		kubeConfigArgs.Namespace = &namespace

		_, err := kubeConfigArgs.ToRESTConfig()
		Expect(err).NotTo(HaveOccurred())

		kubeClientOpts := run.GetKubeClientOptions()

		contextName := "test-context"

		kubeClient, err := run.GetKubeClient(fakeLogger, contextName, k8sEnv.Rest, kubeClientOpts)

		Expect(err).NotTo(HaveOccurred())
		Expect(kubeClient).ToNot(BeNil())
		Expect(kubeClient.ClusterName).To(Equal(contextName))
	})
})
