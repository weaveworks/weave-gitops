package server_test

import (
	"context"
	"testing"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/cache/cachefakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/core/server"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Not using a counterfeit: I want the real methods on the provided
// `client.Client` to be invoked.
type clientMock struct {
	client.Client
}

func (c clientMock) RestConfig() *rest.Config {
	return &rest.Config{}
}

func TestStartServer(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()
	log := logr.Discard()
	cacheContainer := &cachefakes.FakeContainer{}
	cfg := server.NewCoreConfig(log, &rest.Config{}, cacheContainer, "test-cluster")
	svc, err := server.NewCoreServer(cfg)
	g.Expect(err).NotTo(HaveOccurred())

	appName := "my app"
	ns := &corev1.Namespace{}
	ns.Name = "test"
	kust := &kustomizev1.Kustomization{
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: "GitRepository",
			},
		},
	}
	kust.Name = appName
	kust.Namespace = ns.Name

	client := fake.NewClientBuilder().
		WithScheme(kube.CreateScheme()).
		WithRuntimeObjects(kust, ns).
		Build()

	clientsPool := clustersmngrfakes.FakeClientsPool{}
	clientsPool.ClientsReturns(map[string]clustersmngr.ClusterClient{"default": clientMock{client}})
	clientsPool.ClientReturns(clientMock{client}, nil)

	clusterClient := clustersmngr.NewClient(&clientsPool)
	ctx = context.WithValue(ctx, clustersmngr.ClustersClientCtxKey, clusterClient)

	resp, err := svc.ListKustomizations(ctx, &pb.ListKustomizationsRequest{
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(resp.Kustomizations).To(HaveLen(1))
	g.Expect(resp.Kustomizations[0].Namespace).To(Equal(ns.Name))
	g.Expect(resp.Kustomizations[0].Name).To(Equal(appName))
}
