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
	"k8s.io/apimachinery/pkg/util/rand"
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

	scheme = kube.CreateScheme()
	cfg, err = testenv.Start()

	Expect(err).NotTo(HaveOccurred())

	ns = corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: kube.WeGONamespace},
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
	var (
		namespace *corev1.Namespace
		err       error
	)

	BeforeEach(func() {
		namespace = &corev1.Namespace{}
		namespace.Name = "kube-test-" + rand.String(5)
		err = k8sClient.Create(context.Background(), namespace)
		Expect(err).NotTo(HaveOccurred(), "failed to create test namespace")

		k = &kube.KubeHTTP{Client: k8sClient, ClusterName: testClustername}
	})
	AfterEach(func() {
		err = k8sClient.Delete(context.Background(), namespace)
		Expect(err).NotTo(HaveOccurred(), "failed to delete test namespace")
	})
	It("GetClusterName", func() {
		name, err := k.GetClusterName(context.Background())
		Expect(err).NotTo(HaveOccurred())

		Expect(name).To(Equal(testClustername))
	})
	It("GetClusterStatus", func() {
		ctx := context.Background()
		status := k.GetClusterStatus(ctx)

		Expect(status.String()).To(Equal(kube.Unknown.String()))
		// At present, tests are execute in the same testenv instance.
		// As a result tests will interfere with each other.
		// For that reason, only the initial unknown case is testable.
		// TODO: ensure test are all isolated.

	})
	It("FluxPresent", func() {
		ctx := context.Background()

		exists1, err := k.FluxPresent(ctx)
		Expect(err).NotTo(HaveOccurred())

		// Flux doesn't exist yet
		Expect(exists1).To(BeFalse())

		fluxNs := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: kube.FluxNamespace,
			},
		}

		// Create the namespace
		err = testclient.Create(ctx, &fluxNs)
		Expect(err).NotTo(HaveOccurred())

		exists2, err := k.FluxPresent(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(exists2).To(BeTrue())

	})

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
	It("SecretPresent", func() {
		name := "my-secret"
		ns := "default"
		ctx := context.Background()
		secret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		}

		err = testclient.Create(ctx, &secret)
		Expect(err).NotTo(HaveOccurred())

		exists, err := k.SecretPresent(ctx, name, ns)
		Expect(err).NotTo(HaveOccurred())

		Expect(exists).To(BeTrue())
	})
	It("GetApplications", func() {
		ctx := context.Background()
		name := "my-app"
		// TODO: this currently relies on the previous GetApplication test case.
		// This is very bad and I intend on fixing the
		// entire test environment isolation issue in a later PR.

		list, err := k.GetApplications(ctx, name)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(list)).To(Equal(1))
		Expect(list[0].Name).To(Equal(name))

	})
})
