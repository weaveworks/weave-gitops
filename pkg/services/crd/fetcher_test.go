package crd_test

import (
	"testing"

	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/pkg/kube"
)

func TestFetcher_IsAvailable(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	ctx := t.Context()

	service, err := newService(ctx, k8sEnv)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: func() *runtime.Scheme {
			scheme, _ := kube.CreateScheme()
			return scheme
		}(),
	})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	var found bool

	found = service.IsAvailable(ctx, defaultClusterName, "customobjects.example.com")
	g.Expect(found).To(gomega.BeFalse(), "customobjects crd should not be defined in %s cluster", defaultClusterName)

	newCRD(ctx, g, k,
		CRDInfo{
			Singular: "customobject",
			Group:    "example.com",
			Plural:   "customobjects",
			Kind:     "CustomObject",
		})

	service.UpdateCRDList(ctx)

	found = service.IsAvailable(ctx, defaultClusterName, "customobjects.example.com")
	g.Expect(found).To(gomega.BeTrue(), "customobjects crd should be defined in %s cluster", defaultClusterName)

	found = service.IsAvailable(ctx, defaultClusterName, "somethingelse.example.com")
	g.Expect(found).To(gomega.BeFalse(), "somethingelse crd should not be defined in %s Cluster", defaultClusterName)

	found = service.IsAvailable(ctx, "Other", "customobjects.example.com")
	g.Expect(found).To(gomega.BeFalse(), "customobjects crd should not be defined in Other cluster")
}

func TestFetcher_IsAvailableOnClusters(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	ctx := t.Context()

	service, err := newService(ctx, k8sEnv)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: func() *runtime.Scheme {
			scheme, _ := kube.CreateScheme()
			return scheme
		}(),
	})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	newCRD(ctx, g, k,
		CRDInfo{
			Singular: "xclustercustomon",
			Group:    "example.com",
			Plural:   "xclustercustomons",
			Kind:     "CrossClusterCustomObject",
		},
	)

	crdName := "xclustercustomons.example.com"

	service.UpdateCRDList(ctx)

	response := service.IsAvailableOnClusters(ctx, crdName)

	g.Expect(response).To(gomega.HaveLen(1), "cluster list should contain one entry")
	g.Expect(response).To(gomega.HaveKey(defaultClusterName), "cluster list should contain info about %s cluster", defaultClusterName)
	g.Expect(response[defaultClusterName]).To(gomega.BeTrue(), "%s should be available on %s cluster", crdName, defaultClusterName)

	crdName = "xclusterothercustomons.example.com"

	response = service.IsAvailableOnClusters(ctx, crdName)

	g.Expect(response).To(gomega.HaveLen(1), "cluster list should contain one entry")
	g.Expect(response).To(gomega.HaveKey(defaultClusterName), "cluster list should contain info about %s cluster", defaultClusterName)
	g.Expect(response[defaultClusterName]).To(gomega.BeFalse(), "%s shouldn't be available on %s cluster", crdName, defaultClusterName)
}
