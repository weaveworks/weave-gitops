package clustersmngr_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
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

func TestClientList(t *testing.T) {
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

	clist := &clustersmngr.ClusteredKustomizationList{}

	g.Expect(clustersClient.List(ctx, clist)).To(Succeed())
	g.Expect(clist.Lists["test"].Items).To(HaveLen(1))
	g.Expect(clist.Lists["test"].Items[0].Name).To(Equal("test"))
}
