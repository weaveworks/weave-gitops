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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListHelmReleases(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	newHelmRelease(ctx, appName, ns.Name, k, g)

	res, err := c.ListHelmReleases(ctx, &pb.ListHelmReleasesRequest{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmReleases).To(HaveLen(1))
	g.Expect(res.HelmReleases[0].Name).To(Equal(appName))
}

func TestListHelmReleases_inMultipleNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName1 := "myapp-1"
	ns1 := newNamespace(ctx, k, g)

	newHelmRelease(ctx, appName1, ns1.Name, k, g)

	appName2 := "myapp-2"
	ns2 := newNamespace(ctx, k, g)

	newHelmRelease(ctx, appName2, ns2.Name, k, g)

	res, err := c.ListHelmReleases(ctx, &pb.ListHelmReleasesRequest{})
	g.Expect(err).NotTo(HaveOccurred())

	releasesFound := 0

	for _, r := range res.HelmReleases {
		if r.Name == appName1 || r.Name == appName2 {
			releasesFound++
		}
	}

	g.Expect(releasesFound).To(Equal(2))
}

func TestGetHelmRelease(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp" + rand.String(5)
	ns1 := newNamespace(ctx, k, g)
	ns2 := newNamespace(ctx, k, g)
	ns3 := newNamespace(ctx, k, g)

	newHelmRelease(ctx, appName, ns1.Name, k, g)
	newHelmRelease(ctx, appName, ns2.Name, k, g)

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

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp" + rand.String(5)
	ns1 := newNamespace(ctx, k, g)

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
			LastReleaseRevision: 2,
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

	opt := []client.PatchOption{
		client.ForceOwnership,
		client.FieldOwner("helmrelease-controller"),
	}

	g.Expect(k.Patch(ctx, &helmRelease, client.Apply, opt...)).To(Succeed())

	helmRelease.Status.LastReleaseRevision = 1
	helmRelease.ManagedFields = nil

	g.Expect(k.Status().Patch(ctx, &helmRelease, client.Apply, opt...)).To(Succeed())

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
	g.Expect(k.Create(ctx, &storageSecret)).To(Succeed())

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

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp" + rand.String(5)
	ns1 := newNamespace(ctx, k, g)

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
			LastReleaseRevision: 2,
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

	opt := []client.PatchOption{
		client.ForceOwnership,
		client.FieldOwner("helmrelease-controller"),
	}

	g.Expect(k.Patch(ctx, &helmRelease, client.Apply, opt...)).To(Succeed())

	helmRelease.Status.LastReleaseRevision = 1
	helmRelease.ManagedFields = nil

	g.Expect(k.Status().Patch(ctx, &helmRelease, client.Apply, opt...)).To(Succeed())

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
	g.Expect(k.Create(ctx, &storageSecret)).To(Succeed())

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
	ctx context.Context,
	appName, nsName string,
	k client.Client,
	g *GomegaWithT,
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

	g.Expect(k.Create(ctx, &release)).To(Succeed())

	return release
}
