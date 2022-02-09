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
	var (
		clusterFactory vcluster.Factory
		clusterName    string
	)

	BeforeEach(func() {
		var err error
		clusterName = "test-" + rand.String(10)
		clusterFactory, err = vcluster.NewFactory()
		Expect(err).To(BeNil(), "creating new factory")
		client, err := clusterFactory.Create(context.TODO(), clusterName)
		Expect(err).To(BeNil(), "creating new cluster")

		namespaceObj := &corev1.Namespace{}
		namespaceObj.Name = "test"
		Expect(client.Create(context.TODO(), namespaceObj)).To(Succeed())
	})

	AfterEach(func() {
		Expect(clusterFactory.Delete(context.TODO(), clusterName)).To(Succeed())
	})

	It("Testing creation and deletion of a vcluster", func() {

	})
})
