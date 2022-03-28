package clustersmngr_test

import (
	"context"
	"testing"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	. "github.com/onsi/gomega"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func TestClientGet(t *testing.T) {
	clientsPool := clustersmngr.NewClustersClientsPool()

	err := clientsPool.Add(&auth.UserPrincipal{}, clustersmngr.Cluster{
		Name:      "test",
		Server:    k8sEnv.Rest.Host,
		TLSConfig: k8sEnv.Rest.TLSClientConfig,
	})

	g := NewGomegaWithT(t)
	g.Expect(err).To(BeNil())

	clustersClient := clustersmngr.NewClient(clientsPool)

	kust := &kustomizev1.Kustomization{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
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

	g.Expect(clustersClient.Get(ctx, "test", types.NamespacedName{Name: "test", Namespace: "default"}, k)).To(Succeed())
	g.Expect(k.Name).To(Equal("test"))
}

func TestClientGenericList(t *testing.T) {
	clientsPool := clustersmngr.NewClustersClientsPool()

	err := clientsPool.Add(&auth.UserPrincipal{}, clustersmngr.Cluster{
		Name:      "test",
		Server:    k8sEnv.Rest.Host,
		TLSConfig: k8sEnv.Rest.TLSClientConfig,
	})

	g := NewGomegaWithT(t)
	g.Expect(err).To(BeNil())

	clustersClient := clustersmngr.NewClient(clientsPool)

	kust := &kustomizev1.Kustomization{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
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

	g.Expect(clustersClient.List(ctx, cklist)).To(Succeed())
	g.Expect(cklist.List("test").Items).To(HaveLen(1))
	g.Expect(cklist.List("test").Items[0].Name).To(Equal("test"))

	bucket := &sourcev1.Bucket{
		ObjectMeta: v1.ObjectMeta{
			Name:      "my-bucket",
			Namespace: "default",
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
	g.Expect(cblist.List("test").Items).To(HaveLen(1))
	g.Expect(cblist.List("test").Items[0].Name).To(Equal("my-bucket"))
}
