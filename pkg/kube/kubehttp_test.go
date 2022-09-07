package kube_test

import (
	"context"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/yaml"
)

var _ = Describe("KubeHTTP", func() {
	var (
		namespace *corev1.Namespace
	)

	BeforeEach(func() {
		namespace = &corev1.Namespace{}
		namespace.Name = "kube-test-" + rand.String(5)
		Expect(k8sClient.Create(context.Background(), namespace)).To(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(context.Background(), namespace)).To(Succeed())
	})

	Describe("Getting client with override kubeconfig", func() {
		var (
			origkc string
			dir    string
		)

		BeforeEach(func() {
			origkc = os.Getenv("KUBECONFIG")

			var err error
			dir, err = ioutil.TempDir("", "kube-http-test-")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			os.Setenv("KUBECONFIG", origkc)
			Expect(os.RemoveAll(dir)).To(Succeed())
		})

		It("valid kubeconfig", func() {
			tests := []struct {
				name        string
				expName     string
				clusterName string
			}{
				{"foo", "foo", "foo"},
				{"weave-upa-admin@weave-upa", "weave-upa", "weave-upa"},
				{"user@weave.works@market.eu-west-2.eksctl.io-podinfo", "market.eu-west-2.eksctl.io-podinfo", "user@weave.works@market.eu-west-2.eksctl.io-podinfo"},
				{"user@market.eu-west-2.eksctl.io-podinfo", "market.eu-west-2.eksctl.io-podinfo", "user@market.eu-west-2.eksctl.io-podinfo"},
				{"user@market.eu-west-2.eksctl.io-podinfo", "market.eu-west-2.eksctl.io-podinfo", "market.eu-west-2.eksctl.io-podinfo"},
				{"cluster_name", "cluster-name", "cluster_name"},
			}

			for _, test := range tests {
				createKubeconfig(test.name, test.clusterName, dir, true)

				kube.InClusterConfig = func() (*rest.Config, error) { return nil, rest.ErrNotInCluster }

				_, cname, err := kube.RestConfig()
				Expect(err).NotTo(HaveOccurred(), "Failed to get a kube config")

				Expect(cname).To(Equal(test.expName))
			}
		})

		It("errors when pointing at a missing kubeconfig file", func() {
			t, err := ioutil.TempFile(dir, "not_a_kubeconfig")
			Expect(err).ToNot(HaveOccurred())

			kube.InClusterConfig = func() (*rest.Config, error) { return nil, rest.ErrNotInCluster }

			os.Setenv("KUBECONFIG", t.Name())

			_, _, err = kube.RestConfig()
			Expect(err).To(HaveOccurred())
		})

		It("errors when no current_context set in kubeconfig file", func() {
			kube.InClusterConfig = func() (*rest.Config, error) { return nil, rest.ErrNotInCluster }

			createKubeconfig("foo", "bar", dir, false)

			_, _, err := kube.RestConfig()
			Expect(err).To(HaveOccurred())
		})

		It("returns a sensisble clusterName inCluster", func() {
			kube.InClusterConfig = func() (*rest.Config, error) { return nil, nil }

			_, clusterName, err := kube.RestConfig()
			Expect(err).ToNot(HaveOccurred())

			Expect(clusterName).To(Equal("default"))
		})

		It("derives the clusterName from the env inCluster", func() {
			kube.InClusterConfig = func() (*rest.Config, error) { return nil, nil }
			os.Setenv("CLUSTER_NAME", "foo")

			_, clusterName, err := kube.RestConfig()
			Expect(err).ToNot(HaveOccurred())

			Expect(clusterName).To(Equal("foo"))
		})
	})
})

func createKubeconfig(name, clusterName, dir string, setCurContext bool) {
	f, err := ioutil.TempFile(dir, "test.kubeconfig")
	Expect(err).ToNot(HaveOccurred())

	defer f.Close()

	c := clientcmdapi.Config{}

	if setCurContext {
		c.CurrentContext = name
	}

	c.APIVersion = "v1"
	c.Kind = "Config"
	c.Contexts = append(c.Contexts, clientcmdapi.NamedContext{Name: name, Context: clientcmdapi.Context{Cluster: clusterName}})
	c.Clusters = append(c.Clusters, clientcmdapi.NamedCluster{Name: clusterName, Cluster: clientcmdapi.Cluster{Server: "127.0.0.1"}})

	kubeconfig, err := yaml.Marshal(&c)
	Expect(err).ToNot(HaveOccurred())

	_, err = f.Write(kubeconfig)
	Expect(err).ToNot(HaveOccurred())

	os.Setenv("KUBECONFIG", f.Name())
}
