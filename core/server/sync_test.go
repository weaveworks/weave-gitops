package server_test

import (
	"context"
	"errors"
	"testing"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta2"
	imgautomationv1 "github.com/fluxcd/image-automation-controller/api/v1beta1"
	reflectorv1 "github.com/fluxcd/image-reflector-controller/api/v1beta2"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/metadata"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/weaveworks/weave-gitops/core/fluxsync"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestSync(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c := makeGRPCServer(k8sEnv.Rest, t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	name := "myapp"
	ns := newNamespace(ctx, k, g)

	gitRepo := makeGitRepo(name, *ns)

	kust := makeKustomization(name, *ns, gitRepo)

	chart := makeHelmChart(name, *ns)
	helmRepo := makeHelmRepo(name, *ns)
	hr := makeHelmRelease(name, *ns, helmRepo, chart)

	bucket := makeBucket(name, *ns)
	ociRepo := makeOCIRepo(name, *ns)

	ir := makeImageRepository(name, *ns)
	iua := makeImageUpdateAutomation(name, *ns)

	g.Expect(k.Create(ctx, kust)).Should(Succeed())
	g.Expect(k.Create(ctx, hr)).Should(Succeed())
	g.Expect(k.Create(ctx, bucket)).Should(Succeed())
	g.Expect(k.Create(ctx, chart)).Should(Succeed())
	g.Expect(k.Create(ctx, helmRepo)).Should(Succeed())
	g.Expect(k.Create(ctx, gitRepo)).Should(Succeed())
	g.Expect(k.Create(ctx, ociRepo)).Should(Succeed())
	g.Expect(k.Create(ctx, ir)).Should(Succeed())
	g.Expect(k.Create(ctx, iua)).Should(Succeed())

	tests := []struct {
		name         string
		msg          *pb.SyncFluxObjectRequest
		reconcilable fluxsync.Reconcilable
		source       fluxsync.Reconcilable
	}{{
		name: "helm release no source",
		msg: &pb.SyncFluxObjectRequest{
			Objects: []*pb.ObjectRef{{ClusterName: "Default",
				Kind: helmv2.HelmReleaseKind}},
			WithSource: false,
		},
		reconcilable: fluxsync.HelmReleaseAdapter{HelmRelease: hr},
	}, {
		name: "helm release with source",
		msg: &pb.SyncFluxObjectRequest{
			Objects: []*pb.ObjectRef{{ClusterName: "Default",
				Kind: helmv2.HelmReleaseKind}},
			WithSource: true,
		},
		reconcilable: fluxsync.HelmReleaseAdapter{HelmRelease: hr},
		source:       fluxsync.HelmRepositoryAdapter{HelmRepository: helmRepo},
	}, {
		name: "kustomization no source",
		msg: &pb.SyncFluxObjectRequest{
			Objects: []*pb.ObjectRef{{ClusterName: "Default",
				Kind: kustomizev1.KustomizationKind}},
			WithSource: false,
		},
		reconcilable: fluxsync.KustomizationAdapter{Kustomization: kust},
	}, {
		name: "kustomization with source",
		msg: &pb.SyncFluxObjectRequest{
			Objects: []*pb.ObjectRef{{ClusterName: "Default",
				Kind: kustomizev1.KustomizationKind}},
			WithSource: true,
		},
		reconcilable: fluxsync.KustomizationAdapter{Kustomization: kust},
		source:       fluxsync.GitRepositoryAdapter{GitRepository: gitRepo},
	}, {
		name: "gitrepository",
		msg: &pb.SyncFluxObjectRequest{
			Objects: []*pb.ObjectRef{{ClusterName: "Default",
				Kind: sourcev1.GitRepositoryKind}},
			WithSource: false,
		},
		reconcilable: fluxsync.GitRepositoryAdapter{GitRepository: gitRepo},
	}, {
		name: "bucket",
		msg: &pb.SyncFluxObjectRequest{
			Objects: []*pb.ObjectRef{{ClusterName: "Default",
				Kind: sourcev1b2.BucketKind}},
			WithSource: false,
		},
		reconcilable: fluxsync.BucketAdapter{Bucket: bucket},
	}, {
		name: "helmchart",
		msg: &pb.SyncFluxObjectRequest{
			Objects: []*pb.ObjectRef{{ClusterName: "Default",
				Kind: sourcev1b2.HelmChartKind}},
			WithSource: false,
		},
		reconcilable: fluxsync.HelmChartAdapter{HelmChart: chart},
	}, {
		name: "helmrepository",
		msg: &pb.SyncFluxObjectRequest{
			Objects: []*pb.ObjectRef{{ClusterName: "Default",
				Kind: sourcev1b2.HelmRepositoryKind}},
			WithSource: false,
		},
		reconcilable: fluxsync.HelmRepositoryAdapter{HelmRepository: helmRepo},
	}, {
		name: "ocirepository",
		msg: &pb.SyncFluxObjectRequest{
			Objects: []*pb.ObjectRef{{ClusterName: "Default",
				Kind: sourcev1b2.OCIRepositoryKind}},
			WithSource: false,
		},
		reconcilable: fluxsync.OCIRepositoryAdapter{OCIRepository: ociRepo},
	}, {
		name: "image repository",
		msg: &pb.SyncFluxObjectRequest{
			Objects: []*pb.ObjectRef{{ClusterName: "Default",
				Kind: reflectorv1.ImageRepositoryKind}},
			WithSource: false,
		},
		reconcilable: fluxsync.ImageRepositoryAdapter{ImageRepository: ir},
	}, {
		name: "image update automation",
		msg: &pb.SyncFluxObjectRequest{
			Objects: []*pb.ObjectRef{{ClusterName: "Default",
				Kind: imgautomationv1.ImageUpdateAutomationKind}},
			WithSource: false,
		},
		reconcilable: fluxsync.ImageUpdateAutomationAdapter{ImageUpdateAutomation: iua},
	}, {
		name: "multiple objects",
		msg: &pb.SyncFluxObjectRequest{
			Objects: []*pb.ObjectRef{{ClusterName: "Default",
				Kind: helmv2.HelmReleaseKind}, {ClusterName: "Default",
				Kind: helmv2.HelmReleaseKind}},
			WithSource: true,
		},
		reconcilable: fluxsync.HelmReleaseAdapter{HelmRelease: hr},
		source:       fluxsync.HelmRepositoryAdapter{HelmRepository: helmRepo},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.msg
			for _, msg := range msg.Objects {
				msg.Name = tt.reconcilable.GetName()
				msg.Namespace = tt.reconcilable.GetNamespace()
			}

			done := make(chan error)
			defer close(done)

			go func() {
				md := metadata.Pairs(MetadataUserKey, "anne", MetadataGroupsKey, "system:masters")
				outgoingCtx := metadata.NewOutgoingContext(ctx, md)
				_, err := c.SyncFluxObject(outgoingCtx, msg)
				select {
				case <-done:
					return
				default:
					done <- err
				}
			}()

			ticker := time.NewTicker(500 * time.Millisecond)
			for {
				select {
				case <-ticker.C:
					if tt.msg.WithSource {
						sn := types.NamespacedName{
							Name:      tt.source.GetName(),
							Namespace: tt.source.GetNamespace(),
						}
						if err := simulateReconcile(ctx, k, sn, tt.source.AsClientObject()); err != nil {
							t.Fatal(err)
						}
					}

					an := types.NamespacedName{Name: name, Namespace: ns.Name}
					if err := simulateReconcile(ctx, k, an, tt.reconcilable.AsClientObject()); err != nil {
						t.Fatal(err)
					}

				case err := <-done:
					if err != nil {
						t.Error(err)
					}
					return
				}
			}
		})
	}
}

func simulateReconcile(ctx context.Context, k client.Client, name types.NamespacedName, o client.Object) error {
	switch obj := o.(type) {
	case *helmv2.HelmRelease:
		if err := k.Get(ctx, name, obj); err != nil {
			return err
		}

		obj.Status.SetLastHandledReconcileRequest(time.Now().Format(time.RFC3339Nano))

		return k.Status().Update(ctx, obj)

	case *kustomizev1.Kustomization:
		if err := k.Get(ctx, name, obj); err != nil {
			return err
		}

		obj.Status.SetLastHandledReconcileRequest(time.Now().Format(time.RFC3339Nano))

		return k.Status().Update(ctx, obj)

	case *sourcev1b2.Bucket:
		if err := k.Get(ctx, name, obj); err != nil {
			return err
		}

		obj.Status.SetLastHandledReconcileRequest(time.Now().Format(time.RFC3339Nano))

		return k.Status().Update(ctx, obj)

	case *sourcev1.GitRepository:
		if err := k.Get(ctx, name, obj); err != nil {
			return err
		}

		obj.Status.SetLastHandledReconcileRequest(time.Now().Format(time.RFC3339Nano))

		return k.Status().Update(ctx, obj)

	case *sourcev1b2.HelmChart:
		if err := k.Get(ctx, name, obj); err != nil {
			return err
		}

		obj.Status.SetLastHandledReconcileRequest(time.Now().Format(time.RFC3339Nano))

		return k.Status().Update(ctx, obj)

	case *sourcev1b2.HelmRepository:
		if err := k.Get(ctx, name, obj); err != nil {
			return err
		}

		obj.Status.SetLastHandledReconcileRequest(time.Now().Format(time.RFC3339Nano))

		return k.Status().Update(ctx, obj)

	case *sourcev1b2.OCIRepository:
		if err := k.Get(ctx, name, obj); err != nil {
			return err
		}

		obj.Status.SetLastHandledReconcileRequest(time.Now().Format(time.RFC3339Nano))

		return k.Status().Update(ctx, obj)

	case *reflectorv1.ImageRepository:
		if err := k.Get(ctx, name, obj); err != nil {
			return err
		}

		obj.Status.SetLastHandledReconcileRequest(time.Now().Format(time.RFC3339Nano))

		return k.Status().Update(ctx, obj)

	case *imgautomationv1.ImageUpdateAutomation:
		if err := k.Get(ctx, name, obj); err != nil {
			return err
		}

		obj.Status.SetLastHandledReconcileRequest(time.Now().Format(time.RFC3339Nano))

		return k.Status().Update(ctx, obj)
	}

	return errors.New("simulating reconcile: unsupported type")
}

func makeGitRepo(name string, ns corev1.Namespace) *sourcev1.GitRepository {
	return &sourcev1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns.Name,
		},
		Spec: sourcev1.GitRepositorySpec{
			URL: "https://example.com/owner/repo",
		},
		Status: sourcev1.GitRepositoryStatus{
			ReconcileRequestStatus: meta.ReconcileRequestStatus{
				LastHandledReconcileAt: time.Now().Format(time.RFC3339Nano),
			},
		},
	}
}

func makeKustomization(name string, ns corev1.Namespace, source *sourcev1.GitRepository) *kustomizev1.Kustomization {
	k := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{},
		Status: kustomizev1.KustomizationStatus{
			ReconcileRequestStatus: meta.ReconcileRequestStatus{
				LastHandledReconcileAt: time.Now().Format(time.RFC3339Nano),
			},
		},
	}

	if source != nil {
		k.Spec.SourceRef = kustomizev1.CrossNamespaceSourceReference{
			Kind:      sourcev1.GitRepositoryKind,
			Name:      source.GetName(),
			Namespace: source.GetNamespace(),
		}
	}

	return k
}

func makeHelmChart(name string, ns corev1.Namespace) *sourcev1b2.HelmChart {
	return &sourcev1b2.HelmChart{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns.Name,
		},
		Spec: sourcev1b2.HelmChartSpec{
			Chart:   "somechart",
			Version: "v0.0.1",
			SourceRef: sourcev1b2.LocalHelmChartSourceReference{
				Kind: sourcev1b2.HelmRepositoryKind,
				Name: name,
			},
		},
		Status: sourcev1b2.HelmChartStatus{
			ReconcileRequestStatus: meta.ReconcileRequestStatus{
				LastHandledReconcileAt: time.Now().Format(time.RFC3339Nano),
			},
		},
	}
}

func makeBucket(name string, ns corev1.Namespace) *sourcev1b2.Bucket {
	return &sourcev1b2.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns.Name,
		},
		Spec: sourcev1b2.BucketSpec{},
		Status: sourcev1b2.BucketStatus{
			ReconcileRequestStatus: meta.ReconcileRequestStatus{
				LastHandledReconcileAt: time.Now().Format(time.RFC3339Nano),
			},
		},
	}
}

func makeHelmRepo(name string, ns corev1.Namespace) *sourcev1b2.HelmRepository {
	return &sourcev1b2.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns.Name,
		},
		Spec: sourcev1b2.HelmRepositorySpec{
			URL: "http://example.com",
		},
		Status: sourcev1b2.HelmRepositoryStatus{
			ReconcileRequestStatus: meta.ReconcileRequestStatus{
				LastHandledReconcileAt: time.Now().Format(time.RFC3339Nano),
			},
		},
	}
}

func makeHelmRelease(name string, ns corev1.Namespace, repo *sourcev1b2.HelmRepository, chart *sourcev1b2.HelmChart) *helmv2.HelmRelease {
	return &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns.Name,
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:   chart.Spec.Chart,
					Version: chart.Spec.Version,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Name:      repo.GetName(),
						Namespace: repo.GetNamespace(),
						Kind:      sourcev1b2.HelmRepositoryKind,
					},
				},
			},
		},
		Status: helmv2.HelmReleaseStatus{
			ReconcileRequestStatus: meta.ReconcileRequestStatus{
				LastHandledReconcileAt: time.Now().Format(time.RFC3339Nano),
			},
		},
	}
}

func makeOCIRepo(name string, ns corev1.Namespace) *sourcev1b2.OCIRepository {
	return &sourcev1b2.OCIRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns.Name,
		},
		Spec: sourcev1b2.OCIRepositorySpec{
			URL: "oci://ghcr.io/some/chart",
		},
		Status: sourcev1b2.OCIRepositoryStatus{
			ReconcileRequestStatus: meta.ReconcileRequestStatus{
				LastHandledReconcileAt: time.Now().Format(time.RFC3339Nano),
			},
		},
	}
}

func makeImageRepository(name string, ns corev1.Namespace) *reflectorv1.ImageRepository {
	return &reflectorv1.ImageRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns.Name,
		},
		Spec: reflectorv1.ImageRepositorySpec{},
		Status: reflectorv1.ImageRepositoryStatus{
			ReconcileRequestStatus: meta.ReconcileRequestStatus{
				LastHandledReconcileAt: time.Now().Format(time.RFC3339Nano),
			},
		},
	}
}

func makeImageUpdateAutomation(name string, ns corev1.Namespace) *imgautomationv1.ImageUpdateAutomation {
	iua := &imgautomationv1.ImageUpdateAutomation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns.Name,
		},
		Spec: imgautomationv1.ImageUpdateAutomationSpec{SourceRef: imgautomationv1.CrossNamespaceSourceReference{Kind: sourcev1.GitRepositoryKind}},
		Status: imgautomationv1.ImageUpdateAutomationStatus{
			ReconcileRequestStatus: meta.ReconcileRequestStatus{
				LastHandledReconcileAt: time.Now().Format(time.RFC3339Nano),
			},
		},
	}

	return iua
}
