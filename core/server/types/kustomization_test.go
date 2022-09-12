package types

import (
	"testing"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKustomization(t *testing.T) {
	g := NewGomegaWithT(t)

	format.UseStringerRepresentation = true // Makes the representation more compact

	d, err := time.ParseDuration("1h2m3s")
	g.Expect(err).NotTo(HaveOccurred())

	tests := []struct {
		name        string
		clusterName string
		state       kustomizev1.Kustomization
		result      *pb.Kustomization
	}{
		{
			"empty",
			"Default",
			kustomizev1.Kustomization{},
			&pb.Kustomization{
				SourceRef:   &pb.FluxObjectRef{},
				Interval:    &pb.Interval{},
				Conditions:  []*pb.Condition{},
				ClusterName: "Default",
				DependsOn:   []*pb.NamespacedObjectReference{},
			},
		},
		{
			"same-ns",
			"Default",
			kustomizev1.Kustomization{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-object",
					Namespace: "namespace-of-all-objects",
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "some-version",
				},
				Spec: kustomizev1.KustomizationSpec{
					Interval: metav1.Duration{Duration: d},
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Kind:      "GitRepository",
						Name:      "some-git-repository",
						Namespace: "", // Flux won't put the namespace here if it's the same as the object
					},
				},
			},
			&pb.Kustomization{
				Name:      "some-object",
				Namespace: "namespace-of-all-objects",
				SourceRef: &pb.FluxObjectRef{
					Kind:      pb.FluxObjectKind_KindGitRepository,
					Name:      "some-git-repository",
					Namespace: "namespace-of-all-objects",
				},
				Interval:    &pb.Interval{Hours: 1, Minutes: 2, Seconds: 3},
				Conditions:  []*pb.Condition{},
				ClusterName: "Default",
				ApiVersion:  "some-version",
				DependsOn:   []*pb.NamespacedObjectReference{},
			},
		},
		{
			"cross-ns",
			"Default",
			kustomizev1.Kustomization{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-object",
					Namespace: "namespace-of-object",
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "some-version",
				},
				Spec: kustomizev1.KustomizationSpec{
					Interval: metav1.Duration{Duration: d},
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Kind:      "HelmRepository",
						Name:      "some-helm-repository",
						Namespace: "namespace-of-source",
					},
				},
			},
			&pb.Kustomization{
				Name:      "some-object",
				Namespace: "namespace-of-object",
				SourceRef: &pb.FluxObjectRef{
					Kind:      pb.FluxObjectKind_KindHelmRepository,
					Name:      "some-helm-repository",
					Namespace: "namespace-of-source",
				},
				Interval:    &pb.Interval{Hours: 1, Minutes: 2, Seconds: 3},
				Conditions:  []*pb.Condition{},
				ClusterName: "Default",
				ApiVersion:  "some-version",
				DependsOn:   []*pb.NamespacedObjectReference{},
			},
		},
		{
			"all-fields",
			"Default",
			kustomizev1.Kustomization{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-object",
					Namespace: "namespace-of-object",
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "some-version",
				},
				Spec: kustomizev1.KustomizationSpec{
					Path:     "./my-cluster",
					Interval: metav1.Duration{Duration: d},
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Kind:      "Bucket",
						Name:      "some-bucket",
						Namespace: "namespace-of-source",
					},
					Suspend: true,
				},
				Status: kustomizev1.KustomizationStatus{
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionTrue,
							Reason:             "InstallSucceeded",
							Message:            "Release reconciliation succeeded",
							LastTransitionTime: metav1.Time{Time: time.Date(2022, time.January, 1, 1, 1, 1, 0, time.UTC)},
						},
					},
					LastAppliedRevision:   "what even is a bucket revision?",
					LastAttemptedRevision: "the one before",
					Inventory: &kustomizev1.ResourceInventory{
						Entries: []kustomizev1.ResourceRef{
							{
								ID:      "flux-system_helm-controller_apps_Deployment",
								Version: "v1",
							},
							{
								ID:      "flux-system_kustomize-controller_apps_Deployment",
								Version: "v1",
							},
						},
					},
				},
			},
			&pb.Kustomization{
				Namespace: "namespace-of-object",
				Name:      "some-object",
				Path:      "./my-cluster",
				SourceRef: &pb.FluxObjectRef{
					Kind:      pb.FluxObjectKind_KindBucket,
					Name:      "some-bucket",
					Namespace: "namespace-of-source",
				},
				Interval: &pb.Interval{Hours: 1, Minutes: 2, Seconds: 3},
				Conditions: []*pb.Condition{
					{
						Type:      "Ready",
						Status:    "True",
						Reason:    "InstallSucceeded",
						Message:   "Release reconciliation succeeded",
						Timestamp: "2022-01-01T01:01:01Z",
					},
				},
				LastAppliedRevision:   "what even is a bucket revision?",
				LastAttemptedRevision: "the one before",
				Inventory: []*pb.GroupVersionKind{
					{
						Group:   "apps",
						Kind:    "Deployment",
						Version: "v1",
					},
				},
				Suspended:   true,
				ClusterName: "Default",
				ApiVersion:  "some-version",
				DependsOn:   []*pb.NamespacedObjectReference{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := KustomizationToProto(&tt.state, tt.clusterName, "")
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(res).To(Equal(tt.result))
		})
	}
}
