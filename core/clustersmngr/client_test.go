package clustersmngr_test

import (
	"context"
	"testing"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func TestClientGet(t *testing.T) {
	g := NewGomegaWithT(t)
	ns := createNamespace(g)

	appName := "myapp" + rand.String(5)

	clientsPool := clustersmngr.NewClustersClientsPool()

	err := clientsPool.Add(&auth.UserPrincipal{}, clustersmngr.Cluster{
		Name:      appName,
		Server:    k8sEnv.Rest.Host,
		TLSConfig: k8sEnv.Rest.TLSClientConfig,
	})

	g.Expect(err).To(BeNil())

	clustersClient := clustersmngr.NewClient(clientsPool)

	kust := &kustomizev1.Kustomization{
		ObjectMeta: v1.ObjectMeta{
			Name:      appName,
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: "GitRepository",
			},
		},
	}
	ctx := context.Background()
	g.Expect(k8sEnv.Client.Create(ctx, kust)).To(Succeed())

	k := &kustomizev1.Kustomization{}

	g.Expect(clustersClient.Get(ctx, appName, types.NamespacedName{Name: appName, Namespace: ns.Name}, k)).To(Succeed())
	g.Expect(k.Name).To(Equal(appName))
}

func TestClientGenericList(t *testing.T) {
	g := NewGomegaWithT(t)
	ns := createNamespace(g)

	appName := "myapp" + rand.String(5)

	clientsPool := clustersmngr.NewClustersClientsPool()

	err := clientsPool.Add(&auth.UserPrincipal{}, clustersmngr.Cluster{
		Name:      appName,
		Server:    k8sEnv.Rest.Host,
		TLSConfig: k8sEnv.Rest.TLSClientConfig,
	})
	g.Expect(err).To(BeNil())

	clustersClient := clustersmngr.NewClient(clientsPool)

	kust := &kustomizev1.Kustomization{
		ObjectMeta: v1.ObjectMeta{
			Name:      appName,
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: "GitRepository",
			},
		},
	}
	ctx := context.Background()
	g.Expect(k8sEnv.Client.Create(ctx, kust)).To(Succeed())

	cklist := clustersmngr.NewClusteredList(func() *kustomizev1.KustomizationList {
		return &kustomizev1.KustomizationList{}
	})

	g.Expect(clustersClient.List(ctx, cklist, client.InNamespace(ns.Name))).To(Succeed())
	g.Expect(cklist.List(appName).Items).To(HaveLen(1))
	g.Expect(cklist.List(appName).Items[0].Name).To(Equal(appName))

	bucket := &sourcev1.Bucket{
		ObjectMeta: v1.ObjectMeta{
			Name:      appName,
			Namespace: ns.Name,
		},
		Spec: sourcev1.BucketSpec{
			SecretRef: &meta.LocalObjectReference{
				Name: "somesecret",
			},
		},
	}

	g.Expect(k8sEnv.Client.Create(ctx, bucket)).To(Succeed())

	cblist := clustersmngr.NewClusteredList(func() *sourcev1.BucketList {
		return &sourcev1.BucketList{}
	})

	g.Expect(clustersClient.List(ctx, cblist)).To(Succeed())
	g.Expect(cblist.List(appName).Items).To(HaveLen(1))
	g.Expect(cblist.List(appName).Items[0].Name).To(Equal(appName))
}

func createNamespace(g *GomegaWithT) *corev1.Namespace {
	ns := &corev1.Namespace{}
	ns.Name = "kube-test-" + rand.String(5)

	g.Expect(k8sEnv.Client.Create(context.Background(), ns)).To(Succeed())

	return ns
}
