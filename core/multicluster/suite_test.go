package multicluster_test

import (
	"context"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/multicluster"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var k8sEnv *testutils.K8sTestEnv

func TestMain(m *testing.M) {
	os.Setenv("KUBEBUILDER_ASSETS", "../../tools/bin/envtest")

	var err error
	k8sEnv, err = testutils.StartK8sTestEnvironment([]string{
		"../../manifests/crds",
		"../../tools/testcrds",
	})

	if err != nil {
		panic(err)
	}

	code := m.Run()

	k8sEnv.Stop()

	os.Exit(code)
}

func makeLeafCluster(t *testing.T) multicluster.Cluster {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	namespace := &corev1.Namespace{}
	namespace.Name = "kube-test-" + rand.String(5)

	_, err := controllerutil.CreateOrUpdate(ctx, k8sEnv.Client, namespace, func() error {
		return nil
	})
	g.Expect(err).To(BeNil())

	svcAcctSecret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{Name: "weave-gitops-server-token", Namespace: namespace.Name},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, k8sEnv.Client, svcAcctSecret, func() error {
		return nil
	})
	g.Expect(err).To(BeNil())

	svcAcct := &corev1.ServiceAccount{
		ObjectMeta: v1.ObjectMeta{Name: "weave-gitops-server", Namespace: namespace.Name},
		Secrets: []corev1.ObjectReference{
			{Name: svcAcctSecret.Name, Namespace: namespace.Name},
		},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, k8sEnv.Client, svcAcct, func() error {
		return nil
	})
	g.Expect(err).To(BeNil())

	return multicluster.Cluster{
		Server: k8sEnv.Rest.Host,
		Name:   "leaf-cluster",
	}
}
