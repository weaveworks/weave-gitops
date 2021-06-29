package kube_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
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
var ns corev1.Namespace

var _ = BeforeSuite(func() {
	testenv = &envtest.Environment{CRDDirectoryPaths: []string{"../../manifests/crds"}}

	cfg, err = testenv.Start()

	Expect(err).NotTo(HaveOccurred())

	scheme = kube.CreateScheme()

	ns = corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "wego-system"},
	}

	testclient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())

	err = testclient.Create(context.Background(), &ns)
	Expect(err).NotTo(HaveOccurred())
})

var _ = BeforeEach(func() {
	k = &kube.KubeHTTP{
		Client:      testclient,
		ClusterName: testClustername,
	}
})

var _ = Describe("KubeHTTP", func() {
	It("GetApplication", func() {
		ctx := context.Background()
		name := "my-app"
		app := &wego.Application{ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: kube.WeGONamespace,
		}}

		err = testclient.Create(ctx, app)
		Expect(err).NotTo(HaveOccurred())

		a, err := k.GetApplication(ctx, name)
		Expect(err).NotTo(HaveOccurred())
		Expect(a.Name).To(Equal(name))
	})
})
