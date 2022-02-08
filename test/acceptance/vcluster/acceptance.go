package vcluster

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/test/vcluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

var _ = Describe("Acceptance PoC", func() {
	BeforeEach(func() {
		clusterFactory, err := vcluster.NewFactory()
		Expect(err).To(BeNil(), "creating new factory")
		client, err := clusterFactory.Create(context.TODO(), "test-"+rand.String(10))
		Expect(err).To(BeNil(), "creating new cluster")

		namespaceObj := &corev1.Namespace{}
		namespaceObj.Name = "test"
		Expect(client.Create(context.TODO(), namespaceObj)).To(Succeed())
	})

	It("Verify that gitops-flux can print out the version of flux", func() {

	})
})
