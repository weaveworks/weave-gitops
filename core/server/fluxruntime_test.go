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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestListFluxRuntimeObjects(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	tests := []struct {
		description string
		objects     []runtime.Object
		assertions  func(*pb.ListFluxRuntimeObjectsResponse)
	}{
		{
			"no flux runtime",
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
				newDeployment("random-flux-controller", "flux-ns", map[string]string{coretypes.PartOfLabel: server.FluxNamespacePartOf}),
				newDeployment("other-controller-in-flux-ns", "flux-ns", map[string]string{}),
			},
			func(res *pb.ListFluxRuntimeObjectsResponse) {
				g.Expect(res.Deployments).To(HaveLen(1), "expected deployments in the flux namespace to be returned")
				g.Expect(res.Deployments[0].Name).To(Equal("random-flux-controller"))
			},
		},
		{
			"use flux-system namespace when no namespace label available",
			[]runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "flux-system"}},
				newDeployment("random-flux-controller", "flux-system", map[string]string{coretypes.PartOfLabel: server.FluxNamespacePartOf}),
				newDeployment("other-controller-in-flux-ns", "flux-system", map[string]string{}),
			},
			func(res *pb.ListFluxRuntimeObjectsResponse) {
				g.Expect(res.Deployments).To(HaveLen(1), "expected deployments in the default flux namespace to be returned")
				g.Expect(res.Deployments[0].Name).To(Equal("random-flux-controller"))
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

func newDeployment(name, ns string, labels map[string]string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					coretypes.AppLabel: name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{coretypes.AppLabel: name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "nginx",
						Image: "nginx",
					}},
				},
			},
		},
	}
}

func TestListFluxCrds(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	crd1 := &apiextensions.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{
		Name:   "crd1",
		Labels: map[string]string{coretypes.PartOfLabel: "flux"},
	}, Spec: apiextensions.CustomResourceDefinitionSpec{
		Group:    "group",
		Names:    apiextensions.CustomResourceDefinitionNames{Plural: "plural", Kind: "kind"},
		Versions: []apiextensions.CustomResourceDefinitionVersion{},
	}}
	crd2 := &apiextensions.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{
		Name:   "crd2",
		Labels: map[string]string{coretypes.PartOfLabel: "flux"},
	}, Spec: apiextensions.CustomResourceDefinitionSpec{
		Group: "group",
		Versions: []apiextensions.CustomResourceDefinitionVersion{
			{Name: "0"},
			// "Active" version in etcd, use this one.
			{Name: "1", Storage: true},
		},
	}}
	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(crd1, crd2).Build()
	cfg := makeServerConfig(client, t, "")
	c := makeServer(cfg, t)

	res, err := c.ListFluxCrds(ctx, &pb.ListFluxCrdsRequest{})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Crds).To(HaveLen(2))

	first := res.Crds[0]
	g.Expect(first.Version).To(Equal(""))
	g.Expect(first.Name.Plural).To(Equal("plural"))
	g.Expect(first.Name.Group).To(Equal("group"))
	g.Expect(first.Kind).To(Equal("kind"))
	g.Expect(first.ClusterName).To(Equal(cluster.DefaultCluster))
	g.Expect(res.Crds[1].Version).To(Equal("1"))
}
