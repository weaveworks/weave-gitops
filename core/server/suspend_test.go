package server_test

import (
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	imgautomationv1 "github.com/fluxcd/image-automation-controller/api/v1"
	reflectorv1 "github.com/fluxcd/image-reflector-controller/api/v1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/metadata"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

func TestSuspend_Suspend(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := t.Context()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	c := makeGRPCServer(ctx, t, k8sEnv.Rest)

	ns := newNamespace(ctx, k, g)

	gr := makeGitRepo("git-repo-1", *ns)

	hr := makeHelmRepo("repo-1", *ns)

	tests := []struct {
		kind       string
		apiVersion string
		obj        client.Object
	}{
		{
			kind:       sourcev1.GitRepositoryKind,
			apiVersion: sourcev1.GroupVersion.String(),
			obj:        gr,
		},
		{
			kind:       sourcev1.HelmRepositoryKind,
			apiVersion: sourcev1.GroupVersion.String(),
			obj:        hr,
		},
		{
			kind:       sourcev1.BucketKind,
			apiVersion: sourcev1.GroupVersion.String(),
			obj:        makeBucket("bucket-1", *ns),
		},
		{
			kind:       kustomizev1.KustomizationKind,
			apiVersion: kustomizev1.GroupVersion.String(),
			obj:        makeKustomization("kust-1", *ns, gr),
		},
		{
			kind:       helmv2.HelmReleaseKind,
			apiVersion: helmv2.GroupVersion.String(),
			obj:        makeHelmRelease("hr-1", *ns, hr, makeHelmChart("somechart", *ns)),
		},
		{
			kind:       reflectorv1.ImageRepositoryKind,
			apiVersion: reflectorv1.GroupVersion.String(),
			obj:        makeImageRepository("ir-1", *ns),
		},
		{
			kind:       imgautomationv1.ImageUpdateAutomationKind,
			apiVersion: imgautomationv1.GroupVersion.String(),
			obj:        makeImageUpdateAutomation("iua-1", *ns),
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
				Comment: "testing some things out",
			}
			principalID := "anne"
			md := metadata.Pairs(MetadataUserKey, principalID, MetadataGroupsKey, "system:masters")
			outgoingCtx := metadata.NewOutgoingContext(ctx, md)
			_, err = c.ToggleSuspendResource(outgoingCtx, req)
			g.Expect(err).NotTo(HaveOccurred())
			name := types.NamespacedName{Name: tt.obj.GetName(), Namespace: ns.Name}

			unstructuredObj := getUnstructuredObj(t, k, name, tt.kind, tt.apiVersion)
			suspendVal, _, err := unstructured.NestedBool(unstructuredObj.Object, "spec", "suspend")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(suspendVal).To(BeTrue())

			checkSuspendAnnotations(t, principalID, unstructuredObj.GetAnnotations(), name, suspendVal)

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
			unstructuredObj := getUnstructuredObj(t, k, name, tt.kind, tt.apiVersion)
			suspendVal, _, err := unstructured.NestedBool(unstructuredObj.Object, "spec", "suspend")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(suspendVal).To(BeFalse())

			checkSuspendAnnotations(t, principalID, unstructuredObj.GetAnnotations(), name, suspendVal)
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
			}, {
				Kind:        sourcev1.GitRepositoryKind,
				Name:        "fakeName2",
				Namespace:   "fakeNamespace2",
				ClusterName: "Default2",
			}},
			Suspend: true,
		})

		g.Expect(err.Error()).To(ContainSubstring("2 errors occurred"))
	})
}

func getUnstructuredObj(t *testing.T, k client.Client, name types.NamespacedName, kind, apiVersion string) *unstructured.Unstructured {
	t.Helper()
	unstructuredObj := &unstructured.Unstructured{}
	unstructuredObj.SetKind(kind)
	unstructuredObj.SetAPIVersion(apiVersion)
	if err := k.Get(t.Context(), name, unstructuredObj); err != nil {
		t.Error(err)
	}

	return unstructuredObj
}

// checkSuspendAnnotations checks for the existence of suspend annotations
// passes if suspended and annotations exist, or not suspended and annotations don't exist
// if annotations exist, the principal is checked in the annotation for suspended-by
func checkSuspendAnnotations(t *testing.T, principalID string, annotations map[string]string, name types.NamespacedName, suspend bool) {
	t.Helper()
	if suspend {
		// suspended and annotations exist check
		if suspendedBy, ok := annotations["metadata.weave.works/suspended-by"]; ok {
			// principal check if suspended and annotations exist
			if suspendedBy != principalID {
				t.Errorf("expected annotation metadata.weave.works/suspended-by to be set to the principal %s", principalID)
			}
		} else {
			t.Errorf("expected annotation metadata.weave.works/suspended-by not found for %s", name)
		}
		if _, ok := annotations["metadata.weave.works/suspended-comment"]; !ok {
			t.Errorf("expected annotation metadata.weave.works/suspended-comment not found for %s", name)
		}
	} else {
		// not suspended and annotations don't exist check
		if _, ok := annotations["metadata.weave.works/suspended-by"]; ok {
			t.Errorf("expected annotation metadata.weave.works/suspended-by not found for %s", name)
		}
		if _, ok := annotations["metadata.weave.works/suspended-comment"]; ok {
			t.Errorf("expected annotation metadata.weave.works/suspended-comment not found for %s", name)
		}
	}
}
