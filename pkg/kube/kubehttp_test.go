package kube_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var testClustername = "test-cluster"
var testenv *envtest.Environment
var cfg *rest.Config
var testclient client.Client
var err error
var scheme *apiruntime.Scheme
var k kube.Kube

var _ = BeforeSuite(func() {
	testenv = &envtest.Environment{CRDDirectoryPaths: []string{"../../config/crd"}}

	cfg, err = testenv.Start()

	Expect(err).NotTo(HaveOccurred())

	scheme = kube.CreateScheme()
})

var _ = BeforeEach(func() {
	testclient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())

	k = &kube.KubeHTTP{
		Client:      testclient,
		ClusterName: testClustername,
	}

	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("KubeHTTP", func() {
	It("GetApplication", func() {
		ctx := context.Background()
		name := "my-app"
		app := &wego.Application{ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		}}

		err = testclient.Create(ctx, app)
		Expect(err).NotTo(HaveOccurred())

		a, err := k.GetApplication(ctx, name)
		Expect(err).NotTo(HaveOccurred())
		Expect(a.Name).To(Equal(name))
	})
})
