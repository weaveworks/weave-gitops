package server_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/server"
	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestListGitopsRuntimeObjects(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	tests := []struct {
		description string
		objects     []runtime.Object
		assertions  func(*pb.ListFluxRuntimeObjectsResponse)
	}{
		{
			"no gitops runtime",
			[]runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
			},
			func(res *pb.ListFluxRuntimeObjectsResponse) {
				g.Expect(res.Errors[0].Message).To(Equal(server.ErrFluxNamespaceNotFound.Error()))
				g.Expect(res.Errors[0].Namespace).To(BeEmpty())
				g.Expect(res.Errors[0].ClusterName).To(Equal(cluster.DefaultCluster))
			},
		},
		{
			"flux namespace label, with controllers",
			[]runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "flux-ns", Labels: map[string]string{
					coretypes.PartOfLabel: server.FluxNamespacePartOf,
				}}},
				newDeployment("kustomize-controller", "flux-ns", map[string]string{coretypes.PartOfLabel: server.FluxNamespacePartOf}),
				newDeployment("weave-gitops-enterprise-mccp-cluster-service", "flux-ns", map[string]string{coretypes.PartOfLabel: server.PartOfWeaveGitops}),
				newDeployment("other-controller-in-flux-ns", "flux-ns", map[string]string{}),
			},
			func(res *pb.ListFluxRuntimeObjectsResponse) {
				g.Expect(res.Deployments).To(HaveLen(2), "expected deployments in the flux namespace to be returned")
				g.Expect(res.Deployments[0].Name).To(Equal("kustomize-controller"))
				g.Expect(res.Deployments[1].Name).To(Equal("weave-gitops-enterprise-mccp-cluster-service"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			scheme, err := kube.CreateScheme()
			g.Expect(err).To(BeNil())
			client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(tt.objects...).Build()
			cfg := makeServerConfig(client, t, "")
			c := makeServer(cfg, t)
			res, err := c.ListFluxRuntimeObjects(ctx, &pb.ListFluxRuntimeObjectsRequest{})
			g.Expect(err).NotTo(HaveOccurred())
			tt.assertions(res)
		})
	}
}
