package server_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestListHelmReleases(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	appName := "myapp"
	ns := newNamespace()

	release := newHelmRelease(appName, ns.Name)

	k := fake.NewClientBuilder().
		WithScheme(kube.CreateScheme()).
		WithRuntimeObjects(&release, ns).
		Build()

	c := makeGRPCServer(k, t)

	res, err := c.ListHelmReleases(ctx, &pb.ListHelmReleasesRequest{
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmReleases).To(HaveLen(1))
	g.Expect(res.HelmReleases[0].Name).To(Equal(appName))
}

func TestGetHelmRelease(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	appName := "myapp" + rand.String(5)
	ns1 := newNamespace()
	ns2 := newNamespace()
	ns3 := newNamespace()

	release1 := newHelmRelease(appName, ns1.Name)
	release2 := newHelmRelease(appName, ns2.Name)

	k := fake.NewClientBuilder().
		WithScheme(kube.CreateScheme()).
		WithRuntimeObjects(&release1, &release2, ns1, ns2, ns3).
		Build()

	c := makeGRPCServer(k, t)

	// Get app from ns1.
	response, err := c.GetHelmRelease(ctx, &pb.GetHelmReleaseRequest{
		Name:        appName,
		Namespace:   ns1.Name,
		ClusterName: clustersmngr.DefaultCluster,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(response.HelmRelease.Name).To(Equal(appName))
	g.Expect(response.HelmRelease.Namespace).To(Equal(ns1.Name))

	// Get app from ns2.
	response, err = c.GetHelmRelease(ctx, &pb.GetHelmReleaseRequest{
		Name:        appName,
		Namespace:   ns2.Name,
		ClusterName: clustersmngr.DefaultCluster,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(response.HelmRelease.Name).To(Equal(appName))
	g.Expect(response.HelmRelease.Namespace).To(Equal(ns2.Name))

	// Get app from ns3, should fail.
	_, err = c.GetHelmRelease(ctx, &pb.GetHelmReleaseRequest{
		Name:        appName,
		Namespace:   ns3.Name,
		ClusterName: clustersmngr.DefaultCluster,
	})

	g.Expect(err).To(HaveOccurred())
}

func TestGetHelmRelease_withInventory(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	appName := "myapp" + rand.String(5)
	ns1 := newNamespace()

	// Create helm release.
	helmRelease := helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ns1.Name,
		},
		Status: helmv2.HelmReleaseStatus{
			LastReleaseRevision: 1,
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind: "GitRepository",
						Name: "somesource",
					},
				},
			},
		},
	}
	helmRelease.ManagedFields = []metav1.ManagedFieldsEntry{
		{
			Manager:   "helmrealease-controller",
			Operation: "Apply",
		},
	}

	// Create helm storage.
	storage := types.HelmReleaseStorage{
		Name:     "",
		Manifest: helmReleaseInventoryObjects(),
	}

	storageData, _ := json.Marshal(storage)

	// Create secret.
	storageSecret := v1.Secret{}
	storageSecret.Namespace = ns1.Name
	storageSecret.Name = fmt.Sprintf(
		"sh.helm.release.v1.%s.v%v",
		helmRelease.GetName(),
		helmRelease.Status.LastReleaseRevision,
	)
	storageSecret.Data = map[string][]byte{
		"release": []byte(base64.StdEncoding.EncodeToString(storageData)),
	}

	k := fake.NewClientBuilder().
		WithScheme(kube.CreateScheme()).
		WithRuntimeObjects(&storageSecret, &helmRelease, ns1).
		Build()

	c := makeGRPCServer(k, t)

	response, err := c.GetHelmRelease(ctx, &pb.GetHelmReleaseRequest{
		Name:        appName,
		Namespace:   ns1.Name,
		ClusterName: clustersmngr.DefaultCluster,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(response.HelmRelease.Name).To(Equal(appName))
	g.Expect(response.HelmRelease.Namespace).To(Equal(ns1.Name))
	g.Expect(response.HelmRelease.Inventory).To(HaveLen(2))
}

func TestGetHelmRelease_withInventoryCompressed(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	appName := "myapp" + rand.String(5)
	ns1 := newNamespace()

	// Create helm release.
	helmRelease := helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ns1.Name,
		},
		Status: helmv2.HelmReleaseStatus{
			LastReleaseRevision: 1,
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind: "GitRepository",
						Name: "somesource",
					},
				},
			},
		},
	}
	helmRelease.ManagedFields = []metav1.ManagedFieldsEntry{
		{
			Manager:   "helmrealease-controller",
			Operation: "Apply",
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

	// Create secret.
	storageSecret := v1.Secret{}
	storageSecret.Namespace = ns1.Name
	storageSecret.Name = fmt.Sprintf(
		"sh.helm.release.v1.%s.v%v",
		helmRelease.GetName(),
		helmRelease.Status.LastReleaseRevision,
	)
	storageSecret.Data = map[string][]byte{
		"release": []byte(base64.StdEncoding.EncodeToString(compressed.Bytes())),
	}

	k := fake.NewClientBuilder().
		WithScheme(kube.CreateScheme()).
		WithRuntimeObjects(&storageSecret, &helmRelease, ns1).
		Build()

	c := makeGRPCServer(k, t)

	response, err := c.GetHelmRelease(ctx, &pb.GetHelmReleaseRequest{
		Name:        appName,
		Namespace:   ns1.Name,
		ClusterName: clustersmngr.DefaultCluster,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(response.HelmRelease.Name).To(Equal(appName))
	g.Expect(response.HelmRelease.Namespace).To(Equal(ns1.Name))
	g.Expect(response.HelmRelease.Inventory).To(HaveLen(2))
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

func newHelmRelease(
	appName, nsName string,
) helmv2.HelmRelease {
	release := helmv2.HelmRelease{
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind: "GitRepository",
						Name: "somesource",
					},
				},
			},
		},
	}
	release.Name = appName
	release.Namespace = nsName

	return release
}
