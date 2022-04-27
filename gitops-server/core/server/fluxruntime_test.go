package server_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/gitops-server/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/gitops-server/core/server"
	stypes "github.com/weaveworks/weave-gitops/gitops-server/core/server/types"
	pb "github.com/weaveworks/weave-gitops/gitops-server/pkg/api/core"
	"github.com/weaveworks/weave-gitops/common/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestGetReconciledObjects(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	automationName := "my-automation"
	ns := newNamespace(ctx, k, g)

	reconciledObj := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: ns.Name,
			Labels: map[string]string{
				server.KustomizeNameKey:      automationName,
				server.KustomizeNamespaceKey: ns.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": automationName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": automationName},
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

	g.Expect(k.Create(ctx, &reconciledObj)).Should(Succeed())

	res, err := c.GetReconciledObjects(ctx, &pb.GetReconciledObjectsRequest{
		AutomationName: automationName,
		Namespace:      ns.Name,
		AutomationKind: pb.AutomationKind_KustomizationAutomation,
		Kinds:          []*pb.GroupVersionKind{{Group: "apps", Version: "v1", Kind: "Deployment"}},
		ClusterName:    clustersmngr.DefaultCluster,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Objects).To(HaveLen(1))

	first := res.Objects[0]
	g.Expect(first.GroupVersionKind.Kind).To(Equal("Deployment"))
	g.Expect(first.Name).To(Equal(reconciledObj.Name))
}

func TestGetChildObjects(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	automationName := "my-automation"
	ns := newNamespace(ctx, k, g)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: ns.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": automationName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": automationName},
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
	g.Expect(k.Create(ctx, deployment)).Should(Succeed())
	// Get after Create to get a populated UID
	g.Expect(k.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, deployment)).To(Succeed())

	rs := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-123abcd", automationName),
			Namespace: ns.Name,
		},
		Spec: appsv1.ReplicaSetSpec{
			Template: deployment.Spec.Template,
			Selector: deployment.Spec.Selector,
		},
	}

	rs.SetOwnerReferences([]metav1.OwnerReference{{
		UID:        deployment.UID,
		APIVersion: appsv1.SchemeGroupVersion.String(),
		Kind:       "Deployment",
		Name:       deployment.Name,
	}})

	g.Expect(k.Create(ctx, rs)).Should(Succeed())

	res, err := c.GetChildObjects(ctx, &pb.GetChildObjectsRequest{
		ParentUid: string(deployment.UID),
		Namespace: ns.Name,
		GroupVersionKind: &pb.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "ReplicaSet",
		},
		ClusterName: clustersmngr.DefaultCluster,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Objects).To(HaveLen(1))

	first := res.Objects[0]
	g.Expect(first.GroupVersionKind.Kind).To(Equal("ReplicaSet"))
	g.Expect(first.Name).To(Equal(rs.Name))
}

func TestListFluxRuntimeObjects(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	name := "random-flux-controller"
	ns := newNamespace(ctx, k, g)

	fluxDep := newDeployment(name, ns.Name)
	fluxDep.ObjectMeta.Labels = map[string]string{
		stypes.PartOfLabel: server.FluxNamespacePartOf,
	}
	g.Expect(k.Create(ctx, fluxDep)).To(Succeed())

	otherDep := newDeployment("other-deployment", ns.Name)
	g.Expect(k.Create(ctx, otherDep)).To(Succeed())

	res, err := c.ListFluxRuntimeObjects(ctx, &pb.ListFluxRuntimeObjectsRequest{Namespace: ns.Name})
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(res.Deployments).To(HaveLen(1), "expected only deployments with the part-of label to be returned")
	g.Expect(res.Deployments[0].Name).To(Equal(fluxDep.Name))
}

func TestListFluxRuntimeObjects_inMultipleNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cfg := makeGRPCServer(k8sEnv.Rest, t)

	k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	existingDeploymentsNo := func() int {
		res, err := c.ListFluxRuntimeObjects(ctx, &pb.ListFluxRuntimeObjectsRequest{})
		g.Expect(err).NotTo(HaveOccurred())

		return len(res.Deployments)
	}()

	name := "random-flux-controller"
	ns := newNamespace(ctx, k, g)
	ns2 := newNamespace(ctx, k, g)

	fluxDep := newDeployment(name, ns.Name)
	fluxDep.ObjectMeta.Labels = map[string]string{
		stypes.PartOfLabel: server.FluxNamespacePartOf,
	}
	g.Expect(k.Create(ctx, fluxDep)).To(Succeed())

	otherDep := newDeployment(name, ns2.Name)
	otherDep.ObjectMeta.Labels = map[string]string{
		stypes.PartOfLabel: server.FluxNamespacePartOf,
	}
	g.Expect(k.Create(ctx, otherDep)).To(Succeed())

	updateNamespaceCache(cfg)

	res, err := c.ListFluxRuntimeObjects(ctx, &pb.ListFluxRuntimeObjectsRequest{})
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(res.Deployments).To(HaveLen(existingDeploymentsNo + 2))
	g.Expect(res.Deployments[0].Name).To(Equal(fluxDep.Name))
}

func newDeployment(name, ns string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": name},
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
