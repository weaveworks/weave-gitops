package run

import (
	runclient "github.com/fluxcd/pkg/runtime/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

var _ = Describe("GetKubeClient", func() {

	It("returns kube client", func() {
		kubeconfigArgs := genericclioptions.NewConfigFlags(false)
		kubeclientOptions := new(runclient.Options)

		namespace := "test-namespace"

		kubeconfigArgs.Namespace = &namespace

		_, err := kubeconfigArgs.ToRESTConfig()
		Expect(err).NotTo(HaveOccurred())

		contextName := "test-context"

		kubeClient, err := GetKubeClient(logger.From(logr.Discard()), contextName, k8sEnv.Rest, kubeclientOptions)

		Expect(err).NotTo(HaveOccurred())
		Expect(kubeClient).ToNot(BeNil())
		Expect(kubeClient.ClusterName).To(Equal(contextName))
	})
})
