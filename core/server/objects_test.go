package server_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/run/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetObject(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	appName := "myapp"

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: sourcev1.GitRepositoryKind,
			},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, kust).Build()

	cfg := makeServerConfig(fakeClient, t, "")
	c := makeServer(cfg, t)

	res, err := c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        appName,
		Namespace:   ns.Name,
		Kind:        kustomizev1.KustomizationKind,
		ClusterName: "Default",
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Object.ClusterName).To(Equal("Default"))
	g.Expect(res.Object.Payload).NotTo(BeEmpty())

	_, err = c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        appName,
		Namespace:   ns.Name,
		Kind:        helmv2.HelmReleaseKind,
		ClusterName: "Default",
	})
	g.Expect(err).To(HaveOccurred())

	_, err = c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        appName,
		Namespace:   ns.Name,
		Kind:        kustomizev1.KustomizationKind,
		ClusterName: "Other",
	})
	g.Expect(err).To(HaveOccurred())

	_, err = c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        "other name",
		Namespace:   ns.Name,
		Kind:        kustomizev1.KustomizationKind,
		ClusterName: "Defauult",
	})
	g.Expect(err).To(HaveOccurred())

	_, err = c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        appName,
		Namespace:   "other ns",
		Kind:        kustomizev1.KustomizationKind,
		ClusterName: "Defauult",
	})
	g.Expect(err).To(HaveOccurred())
}

func TestGetObjectOtherKinds(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	appName := "myapp"

	dep := newDeployment(appName, ns.Name, map[string]string{})

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(&ns, dep).Build()
	cfg := makeServerConfig(client, t, "")

	c := makeServer(cfg, t)

	_, err = c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        appName,
		Namespace:   ns.Name,
		Kind:        "deployment",
		ClusterName: "Default",
	})
	g.Expect(err).To(HaveOccurred())

	err = cfg.PrimaryKinds.Add("deployment", schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})
	g.Expect(err).NotTo(HaveOccurred())

	c = makeServer(cfg, t)

	res, err := c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        appName,
		Namespace:   ns.Name,
		Kind:        "deployment",
		ClusterName: "Default",
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Object.ClusterName).To(Equal("Default"))
	g.Expect(res.Object.Payload).NotTo(BeEmpty())
}

func TestGetObject_HelmReleaseWithInventory(t *testing.T) {
	g := NewGomegaWithT(t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).NotTo(HaveOccurred())

	ctx := context.Background()

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
			LastReleaseRevision: 1,
		},
	}
	// Create helm storage.
	storage := types.HelmReleaseStorage{
		Name:     "",
		Manifest: helmReleaseInventoryObjects(),
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

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, helm1, secret).Build()
	cfg := makeServerConfig(client, t, "")
	c := makeServer(cfg, t)

	res, err := c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        helm1.Name,
		Namespace:   ns.Name,
		Kind:        helmv2.HelmReleaseKind,
		ClusterName: "Default",
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Object.Inventory).To(HaveLen(2))
}

func helmReleaseInventoryObjects() string {
	return `
---
apiVersion: v1
kind: Secret
metadata:
  name: test
  namespace: default
immutable: true
stringData:
  key: "private-key"
---
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: crossplane-provider-aws2
spec:
  package: crossplane/provider-aws:v0.23.0
  controllerConfigRef:
    name: provider-aws
`
}

func TestGetObject_HelmReleaseWithCompressedInventory(t *testing.T) {
	g := NewGomegaWithT(t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).NotTo(HaveOccurred())

	ctx := context.Background()

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
			LastReleaseRevision: 1,
		},
	}
	// Create helm storage.
	storage := types.HelmReleaseStorage{
		Name:     "",
		Manifest: helmReleaseInventoryObjects(),
	}

	storageData, _ := json.Marshal(storage)

	var compressed bytes.Buffer

	compressor := gzip.NewWriter(&compressed)
	_, _ = compressor.Write(storageData)
	compressor.Close()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sh.helm.release.v1.first-helm-name.v1",
			Namespace: ns.Name,
		},
		Data: map[string][]byte{
			"release": []byte(base64.StdEncoding.EncodeToString(compressed.Bytes())),
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, helm1, secret).Build()
	cfg := makeServerConfig(client, t, "")
	c := makeServer(cfg, t)

	res, err := c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        helm1.Name,
		Namespace:   ns.Name,
		Kind:        helmv2.HelmReleaseKind,
		ClusterName: "Default",
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Object.Inventory).To(HaveLen(2))
}

func TestGetObject_HelmReleaseCantGetSecret(t *testing.T) {
	g := NewGomegaWithT(t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).NotTo(HaveOccurred())

	ctx := context.Background()

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
			LastReleaseRevision: 1,
		},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sh.helm.release.v1.first-helm-name.v1",
			Namespace: ns.Name,
		},
		// No data, so that we throw an error trying to decode it
		// This should behave the same as if we're rejected for RBAC reasons
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, helm1, secret).Build()
	cfg := makeServerConfig(client, t, "")
	c := makeServer(cfg, t)

	res, err := c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        helm1.Name,
		Namespace:   ns.Name,
		Kind:        helmv2.HelmReleaseKind,
		ClusterName: "Default",
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Object.Inventory).To(BeEmpty())
}

func TestGetObjectSecret(t *testing.T) {
	g := NewGomegaWithT(t)

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "schhhhh-dont-tell-anybody",
			Namespace: ns.Name,
		},
		Data: map[string][]byte{
			"key": []byte("value"),
		},
	}
	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, secret).Build()

	cfg := makeServerConfig(fakeClient, t, "")
	c := makeServer(cfg, t)

	res, err := c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        secret.Name,
		Namespace:   ns.Name,
		Kind:        "Secret",
		ClusterName: "Default",
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Object.ClusterName).To(Equal("Default"))

	var data map[string]interface{}
	err = json.Unmarshal([]byte(res.Object.Payload), &data)
	g.Expect(err).To(BeNil())
	g.Expect(data["kind"]).To(Equal("Secret"))
	g.Expect(data["metadata"].(map[string]interface{})["name"]).To(Equal(secret.Name))
	g.Expect(data["data"]).To(Equal(map[string]interface{}{"redacted": nil}))
}

func TestListObjectSingle(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
			Labels: map[string]string{
				"toolkit.fluxcd.io/tenant": "Neil",
			},
		},
	}
	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kust-name",
			Namespace: ns.Name,
			UID:       "not a real uid",
		},
		Spec: kustomizev1.KustomizationSpec{},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, kust).Build()
	cfg := makeServerConfig(client, t, "")
	c := makeServer(cfg, t)

	res, err := c.ListObjects(ctx, &pb.ListObjectsRequest{
		Namespace: ns.Name,
		Kind:      kustomizev1.KustomizationKind,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Errors).To(BeEmpty())
	g.Expect(res.Objects).To(HaveLen(1))
	g.Expect(res.Objects[0].ClusterName).To(Equal("Default"))
	g.Expect(res.Objects[0].Payload).To(ContainSubstring("kust-name"))
	g.Expect(res.Objects[0].Uid).To(Equal("not a real uid"))
	g.Expect(res.Objects[0].Tenant).To(Equal("Neil"))
}

func TestListObjectMultiple(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kust-name",
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{},
	}
	helm1 := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "first-helm-name",
			Namespace: ns.Name,
		},
		Spec: helmv2.HelmReleaseSpec{},
	}
	helm2 := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "second-helm-name",
			Namespace: ns.Name,
		},
		Spec: helmv2.HelmReleaseSpec{},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, kust, helm1, helm2).Build()
	cfg := makeServerConfig(client, t, "")
	c := makeServer(cfg, t)

	res, err := c.ListObjects(ctx, &pb.ListObjectsRequest{
		Namespace: ns.Name,
		Kind:      helmv2.HelmReleaseKind,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Errors).To(BeEmpty())
	g.Expect(res.Objects).To(HaveLen(2))
	g.Expect(res.Objects[0].ClusterName).To(Equal("Default"))
	g.Expect(res.Objects[0].Payload).To(ContainSubstring("helm-name"))
	g.Expect(res.Objects[1].Payload).To(ContainSubstring("helm-name"))
}

func TestListObjectSingleWithClusterName(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
			Labels: map[string]string{
				"toolkit.fluxcd.io/tenant": "Neil",
			},
		},
	}
	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kust-name",
			Namespace: ns.Name,
			UID:       "not a real uid",
		},
		Spec: kustomizev1.KustomizationSpec{},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, kust).Build()
	cfg := makeServerConfig(client, t, "")
	c := makeServer(cfg, t)

	res, err := c.ListObjects(ctx, &pb.ListObjectsRequest{
		Namespace:   ns.Name,
		Kind:        kustomizev1.KustomizationKind,
		ClusterName: "Default",
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Errors).To(BeEmpty())
	g.Expect(res.Objects).To(HaveLen(1))
	g.Expect(res.Objects[0].ClusterName).To(Equal("Default"))
	g.Expect(res.Objects[0].Payload).To(ContainSubstring("kust-name"))
	g.Expect(res.Objects[0].Uid).To(Equal("not a real uid"))
	g.Expect(res.Objects[0].Tenant).To(Equal("Neil"))
}

func TestListObjectMultipleWithClusterName(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kust-name",
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{},
	}
	helm1 := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "first-helm-name",
			Namespace: ns.Name,
		},
		Spec: helmv2.HelmReleaseSpec{},
	}
	helm2 := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "second-helm-name",
			Namespace: ns.Name,
		},
		Spec: helmv2.HelmReleaseSpec{},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, kust, helm1, helm2).Build()
	cfg := makeServerConfig(client, t, "")
	c := makeServer(cfg, t)

	res, err := c.ListObjects(ctx, &pb.ListObjectsRequest{
		Namespace:   ns.Name,
		Kind:        helmv2.HelmReleaseKind,
		ClusterName: "Default",
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Errors).To(BeEmpty())
	g.Expect(res.Objects).To(HaveLen(2))
	g.Expect(res.Objects[0].ClusterName).To(Equal("Default"))
	g.Expect(res.Objects[0].Payload).To(ContainSubstring("helm-name"))
	g.Expect(res.Objects[1].Payload).To(ContainSubstring("helm-name"))
}
func TestListObject_HelmReleaseWithInventory(t *testing.T) {
	g := NewGomegaWithT(t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).NotTo(HaveOccurred())

	ctx := context.Background()

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
			LastReleaseRevision: 1,
		},
	}
	// Create helm storage.
	storage := types.HelmReleaseStorage{
		Name:     "",
		Manifest: helmReleaseInventoryObjects(),
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

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, helm1, secret).Build()
	cfg := makeServerConfig(client, t, "")
	c := makeServer(cfg, t)

	res, err := c.ListObjects(ctx, &pb.ListObjectsRequest{
		Namespace: ns.Name,
		Kind:      helmv2.HelmReleaseKind,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Errors).To(BeEmpty())
	g.Expect(res.Objects[0].Inventory).To(HaveLen(2))
}

func TestListObject_HelmReleaseCantGetSecret(t *testing.T) {
	g := NewGomegaWithT(t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).NotTo(HaveOccurred())

	ctx := context.Background()

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
			LastReleaseRevision: 1,
		},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sh.helm.release.v1.first-helm-name.v1",
			Namespace: ns.Name,
		},
		// No data, so that we throw an error trying to decode it
		// This should behave the same as if we're rejected for RBAC reasons
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, helm1, secret).Build()
	cfg := makeServerConfig(client, t, "")
	c := makeServer(cfg, t)

	res, err := c.ListObjects(ctx, &pb.ListObjectsRequest{
		Namespace: ns.Name,
		Kind:      helmv2.HelmReleaseKind,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Errors).To(HaveLen(1))
	g.Expect(res.Objects).To(HaveLen(1))
}

func TestListObjectsSecret(t *testing.T) {
	g := NewGomegaWithT(t)

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "schhhhh-dont-tell-anybody",
			Namespace: ns.Name,
		},
		Data: map[string][]byte{
			"key": []byte("value"),
		},
	}
	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, secret).Build()

	cfg := makeServerConfig(fakeClient, t, "")
	c := makeServer(cfg, t)

	res, err := c.ListObjects(ctx, &pb.ListObjectsRequest{
		Kind:        "Secret",
		ClusterName: "Default",
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Objects).To(HaveLen(1))
	g.Expect(res.Objects[0].ClusterName).To(Equal("Default"))

	var data map[string]interface{}
	err = json.Unmarshal([]byte(res.Objects[0].Payload), &data)
	g.Expect(err).To(BeNil())
	g.Expect(data["kind"]).To(Equal("Secret"))
	g.Expect(data["metadata"].(map[string]interface{})["name"]).To(Equal(secret.Name))
	g.Expect(data["data"]).To(Equal(map[string]interface{}{"redacted": nil}))
}

func TestListObjectsLabels(t *testing.T) {
	g := NewGomegaWithT(t)

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	deployment1 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "the-deployment-i-want",
			Namespace: ns.Name,
			Labels: map[string]string{
				"key": "the-value",
			},
		},
	}
	deployment2 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "the-deployment-i-dont-want",
			Namespace: ns.Name,
			Labels: map[string]string{
				"key": "another-value",
			},
		},
	}

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, deployment1, deployment2).Build()

	cfg := makeServerConfig(fakeClient, t, "")
	c := makeServer(cfg, t)

	res, err := c.ListObjects(ctx, &pb.ListObjectsRequest{
		Kind:        "Deployment",
		ClusterName: "Default",
		Labels: map[string]string{
			"key": "the-value",
		},
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Objects).To(HaveLen(1))

	var data map[string]interface{}
	err = json.Unmarshal([]byte(res.Objects[0].Payload), &data)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(data["metadata"].(map[string]interface{})["name"]).To(Equal(deployment1.Name))
}

func TestListObjectsGitOpsRunSessions(t *testing.T) {
	g := NewGomegaWithT(t)

	const (
		testNS      = "test-namespace"
		testCluster = "test-cluster"
	)

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNS,
		},
	}

	session1 := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "session-1",
			Namespace: ns.Name,
			Labels: map[string]string{
				"app":                       "vcluster",
				"app.kubernetes.io/part-of": "gitops-run",
			},
			Annotations: map[string]string{
				"run.weave.works/automation-kind": "ks",
			},
		},
		Spec: appsv1.StatefulSetSpec{},
	}

	session2 := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "session-2",
			Namespace: ns.Name,
			Labels: map[string]string{
				"app":                       "vcluster",
				"app.kubernetes.io/part-of": "gitops-run",
			},
			Annotations: map[string]string{
				"run.weave.works/automation-kind": "helm",
			},
		},
		Spec: appsv1.StatefulSetSpec{},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, session1, session2).Build()
	cfg := makeServerConfig(fakeClient, t, testCluster)
	c := makeServer(cfg, t)

	res, err := c.ListObjects(ctx, &pb.ListObjectsRequest{
		Namespace:   testNS,
		Kind:        "StatefulSet",
		ClusterName: testCluster,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Errors).To(BeEmpty())
	g.Expect(res.Objects).To(HaveLen(2))
	g.Expect(res.Objects[0].ClusterName).To(Equal(testCluster))
	g.Expect(res.Objects[1].ClusterName).To(Equal(testCluster))
	g.Expect(res.Objects[0].Payload).To(ContainSubstring("session-1"))
	g.Expect(res.Objects[1].Payload).To(ContainSubstring("session-2"))
}

func TestGetObjectSessionObjects(t *testing.T) {
	g := NewGomegaWithT(t)

	const (
		testNS      = "test-namespace"
		testCluster = "test-cluster"
	)

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNS,
		},
	}

	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevKsName,
			Namespace: testNS,
		},
		Spec: kustomizev1.KustomizationSpec{},
	}

	helm := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevHelmName,
			Namespace: testNS,
		},
		Spec: helmv2.HelmReleaseSpec{},
	}

	bucket := &sourcev1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevBucketName,
			Namespace: testNS,
		},
		Spec: sourcev1.BucketSpec{},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, kust, helm, bucket).Build()
	cfg := makeServerConfig(fakeClient, t, testCluster)
	c := makeServer(cfg, t)

	res, err := c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        constants.RunDevKsName,
		Namespace:   testNS,
		Kind:        kustomizev1.KustomizationKind,
		ClusterName: testCluster,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Object.ClusterName).To(ContainSubstring(testCluster))
	g.Expect(res.Object.Payload).To(ContainSubstring(constants.RunDevKsName))

	res, err = c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        constants.RunDevHelmName,
		Namespace:   testNS,
		Kind:        helmv2.HelmReleaseKind,
		ClusterName: testCluster,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Object.ClusterName).To(ContainSubstring(testCluster))
	g.Expect(res.Object.Payload).To(ContainSubstring(constants.RunDevHelmName))

	res, err = c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        constants.RunDevBucketName,
		Namespace:   testNS,
		Kind:        sourcev1.BucketKind,
		ClusterName: testCluster,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Object.ClusterName).To(ContainSubstring(testCluster))
	g.Expect(res.Object.Payload).To(ContainSubstring(constants.RunDevBucketName))
}
