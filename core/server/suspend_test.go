package server_test

import (
	"context"
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	imgautomationv1 "github.com/fluxcd/image-automation-controller/api/v1beta1"
	reflectorv1 "github.com/fluxcd/image-reflector-controller/api/v1beta2"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	. "github.com/onsi/gomega"
	api "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"google.golang.org/grpc/metadata"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestSuspend_Suspend(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	c := makeGRPCServer(k8sEnv.Rest, t)

	ns := newNamespace(ctx, k, g)

	gr := makeGitRepo("git-repo-1", *ns)

	hr := makeHelmRepo("repo-1", *ns)

	tests := []struct {
		kind string
		obj  client.Object
	}{
		{
			kind: sourcev1.GitRepositoryKind,
			obj:  gr,
		},
		{
			kind: sourcev1b2.HelmRepositoryKind,
			obj:  hr,
		},
		{
			kind: sourcev1b2.BucketKind,
			obj:  makeBucket("bucket-1", *ns),
		},
		{
			kind: kustomizev1.KustomizationKind,
			obj:  makeKustomization("kust-1", *ns, gr),
		},
		{
			kind: helmv2.HelmReleaseKind,
			obj:  makeHelmRelease("hr-1", *ns, hr, makeHelmChart("somechart", *ns)),
		},
		{
			kind: reflectorv1.ImageRepositoryKind,
			obj:  makeImageRepository("ir-1", *ns),
		},
		{
			kind: imgautomationv1.ImageUpdateAutomationKind,
			obj:  makeImageUpdateAutomation("iua-1", *ns),
		},
	}

	requestObjects := []*api.ObjectRef{}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			g := NewGomegaWithT(t)
			g.Expect(k.Create(ctx, tt.obj)).Should(Succeed())
			object := &api.ObjectRef{
				Kind:        tt.kind,
				Name:        tt.obj.GetName(),
				Namespace:   tt.obj.GetNamespace(),
				ClusterName: "Default",
			}
			req := &api.ToggleSuspendResourceRequest{
				Objects: []*api.ObjectRef{object},
				Suspend: true,
			}
			principalID := "anne"
			md := metadata.Pairs(MetadataUserKey, principalID, MetadataGroupsKey, "system:masters")
			outgoingCtx := metadata.NewOutgoingContext(ctx, md)
			_, err = c.ToggleSuspendResource(outgoingCtx, req)
			g.Expect(err).NotTo(HaveOccurred())
			name := types.NamespacedName{Name: tt.obj.GetName(), Namespace: ns.Name}
			g.Expect(checkSpec(t, k, principalID, name, tt.obj, true)).To(BeTrue())
			requestObjects = append(requestObjects, object)
		})
	}

	t.Run("resume multiple", func(t *testing.T) {
		req := &api.ToggleSuspendResourceRequest{
			Objects: requestObjects,
			Suspend: false,
		}

		principalID := "anne"
		md := metadata.Pairs(MetadataUserKey, principalID, MetadataGroupsKey, "system:masters")
		outgoingCtx := metadata.NewOutgoingContext(ctx, md)
		_, err = c.ToggleSuspendResource(outgoingCtx, req)
		g.Expect(err).NotTo(HaveOccurred())

		for _, tt := range tests {
			name := types.NamespacedName{Name: tt.obj.GetName(), Namespace: ns.Name}
			g.Expect(checkSpec(t, k, principalID, name, tt.obj, false)).To(BeFalse())
		}
	})

	t.Run("will error", func(t *testing.T) {
		md := metadata.Pairs(MetadataUserKey, "anne", MetadataGroupsKey, "system:masters")
		outgoingCtx := metadata.NewOutgoingContext(ctx, md)
		_, err = c.ToggleSuspendResource(outgoingCtx, &api.ToggleSuspendResourceRequest{

			Objects: []*api.ObjectRef{{
				Kind:        sourcev1.GitRepositoryKind,
				Name:        "fakeName",
				Namespace:   "fakeNamespace",
				ClusterName: "Default",
			}, {Kind: sourcev1.GitRepositoryKind,
				Name:        "fakeName2",
				Namespace:   "fakeNamespace2",
				ClusterName: "Default2"}},
			Suspend: true,
		})

		g.Expect(err.Error()).To(ContainSubstring("2 errors occurred"))
	})
}

func checkSpec(t *testing.T, k client.Client, principalID string, name types.NamespacedName, obj client.Object, suspend bool) bool {
	switch v := obj.(type) {
	case *sourcev1.GitRepository:
		if err := k.Get(context.Background(), name, v); err != nil {
			t.Error(err)
		}
		checkSuspendAnnotations(t, principalID, v.ObjectMeta.Annotations, name, suspend)
		return v.Spec.Suspend
	case *kustomizev1.Kustomization:
		if err := k.Get(context.Background(), name, v); err != nil {
			t.Error(err)
		}
		checkSuspendAnnotations(t, principalID, v.ObjectMeta.Annotations, name, suspend)
		return v.Spec.Suspend

	case *helmv2.HelmRelease:
		if err := k.Get(context.Background(), name, v); err != nil {
			t.Error(err)
		}
		checkSuspendAnnotations(t, principalID, v.ObjectMeta.Annotations, name, suspend)
		return v.Spec.Suspend

	case *sourcev1b2.Bucket:
		if err := k.Get(context.Background(), name, v); err != nil {
			t.Error(err)
		}
		checkSuspendAnnotations(t, principalID, v.ObjectMeta.Annotations, name, suspend)
		return v.Spec.Suspend

	case *sourcev1b2.HelmRepository:
		if err := k.Get(context.Background(), name, v); err != nil {
			t.Error(err)
		}
		checkSuspendAnnotations(t, principalID, v.ObjectMeta.Annotations, name, suspend)
		return v.Spec.Suspend

	case *reflectorv1.ImageRepository:
		if err := k.Get(context.Background(), name, v); err != nil {
			t.Error(err)
		}
		checkSuspendAnnotations(t, principalID, v.ObjectMeta.Annotations, name, suspend)
		return v.Spec.Suspend

	case *imgautomationv1.ImageUpdateAutomation:
		if err := k.Get(context.Background(), name, v); err != nil {
			t.Error(err)
		}
		checkSuspendAnnotations(t, principalID, v.ObjectMeta.Annotations, name, suspend)
		return v.Spec.Suspend
	}

	t.Errorf("unsupported object %T", obj)

	return false
}

// checkSuspendAnnotations checks for the existance of suspend annotations
// passes if suspended and annotations exist, or not suspended and annotations don't exist
// if annotations exist, the principal is checked in the annotation for suspended-by
func checkSuspendAnnotations(t *testing.T, principalID string, annotations map[string]string, name types.NamespacedName, suspend bool) {
	if suspend {
		// suspended and annotations exist check
		if suspendedBy, ok := annotations["weave.works/suspended-by"]; ok {
			// principal check if suspended and annotations exist
			if suspendedBy != principalID {
				t.Errorf("expected annotation weave.works/suspended-by to be set to the principal %s", principalID)
			}
		} else {
			t.Errorf("expected annotation weave.works/suspended-by not found for %s", name)
		}
		if _, ok := annotations["weave.works/suspended-comment"]; !ok {
			t.Errorf("expected annotation weave.works/suspended-comment not found for %s", name)
		}
	} else {
		// not suspended and annotations don't exist check
		if _, ok := annotations["weave.works/suspended-by"]; ok {
			t.Errorf("expected annotation weave.works/suspended-by not found for %s", name)
		}
		if _, ok := annotations["weave.works/suspended-comment"]; ok {
			t.Errorf("expected annotation weave.works/suspended-comment not found for %s", name)
		}
	}
}
