package server_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta/testrestmapper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/server"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

func TestGetInventoryKustomization(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := t.Context()

	automationName := "my-automation"

	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
			Labels: map[string]string{
				"toolkit.fluxcd.io/tenant": "tenant",
			},
		},
	}

	anotherNs := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "another-namespace",
			Labels: map[string]string{
				"toolkit.fluxcd.io/tenant": "tenant",
			},
		},
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: "another-namespace",
			UID:       "this-is-not-an-uid",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					types.AppLabel: automationName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{types.AppLabel: automationName},
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

	rs := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-123abcd", automationName),
			Namespace: "another-namespace",
		},
		Spec: appsv1.ReplicaSetSpec{
			Template: deployment.Spec.Template,
			Selector: deployment.Spec.Selector,
		},
		Status: appsv1.ReplicaSetStatus{
			Replicas: 1,
		},
	}

	rs.SetOwnerReferences([]metav1.OwnerReference{{
		UID:        deployment.UID,
		APIVersion: appsv1.SchemeGroupVersion.String(),
		Kind:       "Deployment",
		Name:       deployment.Name,
	}})

	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      automationName,
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: sourcev1.GitRepositoryKind,
			},
		},
		Status: kustomizev1.KustomizationStatus{
			Inventory: &kustomizev1.ResourceInventory{
				Entries: []kustomizev1.ResourceRef{
					{
						ID:      fmt.Sprintf("%s_%s_apps_Deployment", "another-namespace", deployment.Name),
						Version: "v1",
					},
				},
			},
		},
	}

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(&ns, &anotherNs, kust, deployment, rs).Build()
	cfg := makeServerConfig(t, client, "")
	c := makeServer(ctx, t, cfg)

	res, err := c.GetInventory(ctx, &pb.GetInventoryRequest{
		Namespace:    ns.Name,
		ClusterName:  cluster.DefaultCluster,
		Kind:         "Kustomization",
		Name:         kust.Name,
		WithChildren: true,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Entries).To(HaveLen(1))

	g.Expect(res.Entries[0].Children).To(HaveLen(1))
	g.Expect(res.Entries[0].Tenant).To(Equal("tenant"))
}

func TestGetBlankInventoryKustomization(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := t.Context()

	automationName := "my-automation"
	ns := "test-namespace"

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: ns,
			UID:       "this-is-not-an-uid",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					types.AppLabel: automationName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{types.AppLabel: automationName},
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

	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      automationName,
			Namespace: ns,
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: sourcev1.GitRepositoryKind,
			},
		},
		Status: kustomizev1.KustomizationStatus{
			Inventory: nil, // blank inventory
		},
	}

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(kust, deployment).Build()
	cfg := makeServerConfig(t, client, "")
	c := makeServer(ctx, t, cfg)

	res, err := c.GetInventory(ctx, &pb.GetInventoryRequest{
		Namespace:    ns,
		ClusterName:  cluster.DefaultCluster,
		Kind:         "Kustomization",
		Name:         kust.Name,
		WithChildren: true,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Entries).To(HaveLen(0))
}

func TestGetInventoryHelmRelease(t *testing.T) {
	g := NewGomegaWithT(t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).NotTo(HaveOccurred())

	ctx := t.Context()

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	helm1 := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "first-helm-name",
			Namespace: ns.Name,
		},
		Spec: helmv2.HelmReleaseSpec{},
		Status: helmv2.HelmReleaseStatus{
			History: helmv2.Snapshots{{
				Name:      "first-helm-name",
				Version:   1,
				Namespace: ns.Name,
			}},
		},
	}

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "config-map",
			Namespace: ns.Name,
		},
		Data: map[string]string{
			"key": "value",
		},
	}

	cmData, err := json.Marshal(cm)
	g.Expect(err).NotTo(HaveOccurred())

	// Create helm storage.
	storage := types.HelmReleaseStorage{
		Name:     "",
		Manifest: string(cmData),
	}

	storageData, _ := json.Marshal(storage)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sh.helm.release.v1.first-helm-name.v1",
			Namespace: ns.Name,
		},
		Data: map[string][]byte{
			"release": []byte(base64.StdEncoding.EncodeToString(storageData)),
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, helm1, secret, cm).Build()
	cfg := makeServerConfig(t, client, "")
	c := makeServer(ctx, t, cfg)

	res, err := c.GetInventory(ctx, &pb.GetInventoryRequest{
		Namespace:    ns.Name,
		ClusterName:  cluster.DefaultCluster,
		Kind:         "HelmRelease",
		Name:         helm1.Name,
		WithChildren: true,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Entries).To(HaveLen(1))
}

func TestGetInventoryHelmReleaseNoNSResources(t *testing.T) {
	g := NewGomegaWithT(t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).NotTo(HaveOccurred())

	ctx := t.Context()

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	helm1 := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "first-helm-name",
			Namespace: ns.Name,
		},
		Spec: helmv2.HelmReleaseSpec{
			TargetNamespace: "test-ns",
		},
		Status: helmv2.HelmReleaseStatus{
			StorageNamespace: ns.Name,
			History: helmv2.Snapshots{
				{
					Name:      "first-helm-name",
					Version:   1,
					Namespace: "test-ns",
				},
			},
		},
	}

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "config-map-without-ns",
		},
		Data: map[string]string{
			"key": "value",
		},
	}

	cmData, err := json.Marshal(cm)
	g.Expect(err).NotTo(HaveOccurred())

	// The version in the helm release manifest manifest has no Namespace.
	// This is set to the namespace from the History data.
	cm.SetNamespace("test-ns")

	// Create helm storage.
	storage := types.HelmReleaseStorage{
		Name:     "",
		Manifest: string(cmData),
	}

	storageData, _ := json.Marshal(storage)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sh.helm.release.v1.first-helm-name.v1",
			Namespace: helm1.GetStorageNamespace(),
		},
		Data: map[string][]byte{
			"release": []byte(base64.StdEncoding.EncodeToString(storageData)),
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).
		WithRESTMapper(testrestmapper.TestOnlyStaticRESTMapper(scheme)).
		WithRuntimeObjects(ns, helm1, secret, cm).Build()
	cfg := makeServerConfig(t, client, "")
	c := makeServer(ctx, t, cfg)

	res, err := c.GetInventory(ctx, &pb.GetInventoryRequest{
		Namespace:    ns.Name,
		ClusterName:  cluster.DefaultCluster,
		Kind:         "HelmRelease",
		Name:         helm1.Name,
		WithChildren: true,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Entries).To(HaveLen(1))
}

func TestGetInventoryHelmReleaseWithKubeconfig(t *testing.T) {
	g := NewGomegaWithT(t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).NotTo(HaveOccurred())

	ctx := t.Context()

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	helm1 := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "first-helm-name",
			Namespace: ns.Name,
		},
		Spec: helmv2.HelmReleaseSpec{
			KubeConfig: &meta.KubeConfigReference{
				SecretRef: &meta.SecretKeyReference{
					Name: "kubeconfig",
				},
			},
		},
		Status: helmv2.HelmReleaseStatus{
			History: helmv2.Snapshots{{
				Name:      "first-helm-name",
				Version:   1,
				Namespace: ns.Name,
			}},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, helm1).Build()
	cfg := makeServerConfig(t, client, "")
	c := makeServer(ctx, t, cfg)

	res, err := c.GetInventory(ctx, &pb.GetInventoryRequest{
		Namespace:    ns.Name,
		ClusterName:  cluster.DefaultCluster,
		Kind:         "HelmRelease",
		Name:         helm1.Name,
		WithChildren: true,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Entries).To(HaveLen(0))
}

func TestResourceRefToUnstructured(t *testing.T) {
	testCases := []struct {
		name        string
		id          string
		version     string
		expected    *unstructured.Unstructured
		expectedErr string
	}{
		{
			name:    "valid id",
			id:      "test-namespace_test-name_apps_Deployment",
			version: "v1",
			expected: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name":      "test-name",
						"namespace": "test-namespace",
					},
				},
			},
		},
		{
			name:        "invalid id",
			id:          "foo",
			version:     "v1",
			expectedErr: "unable to parse stored object metadata:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			res, err := server.ResourceRefToUnstructured(tc.id, tc.version)
			if tc.expectedErr != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tc.expectedErr)))
			} else {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(&res).To(Equal(tc.expected))
			}
		})
	}
}
