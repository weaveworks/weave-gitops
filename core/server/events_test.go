package server_test

import (
	"context"
	"fmt"
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	kustomizationObjectName := "my-kustomization"
	helmObjectName := "my-helmrelease"
	ns := newNamespace(ctx, k, g)

	// Kustomization
	kustomizeEvent := &corev1.Event{
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprintf("%s.16da7d2e2c5b0930", kustomizationObjectName),
			Namespace: ns.Name,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Kustomization",
			Namespace: ns.Name,
			Name:      kustomizationObjectName,
		},
		Source: corev1.EventSource{
			Component: "kustomize-controller",
		},
	}

	// HelmRelease
	helmEvent := &corev1.Event{
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprintf("%s.16da7d2e2c5b0930", helmObjectName),
			Namespace: ns.Name,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "HelmRelease",
			Namespace: ns.Name,
			Name:      helmObjectName,
		},
		Source: corev1.EventSource{
			Component: "helm-controller",
		},
	}

	// An event we don't care about. Shoud not show up in our response.
	otherEvent := &corev1.Event{
		ObjectMeta: v1.ObjectMeta{
			Name:      "someotherobject.16da7d2e2c5b0930",
			Namespace: ns.Name,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Kustomization",
			Namespace: ns.Name,
			Name:      "someotherobject",
		},
		Source: corev1.EventSource{
			Component: "kustomize-controller",
		},
	}

	g.Expect(k.Create(ctx, kustomizeEvent)).To(Succeed())
	g.Expect(k.Create(ctx, helmEvent)).To(Succeed())
	g.Expect(k.Create(ctx, otherEvent)).To(Succeed())

	// Get kustomization events
	res, err := c.ListEvents(ctx, &pb.ListEventsRequest{
		InvolvedObject: &pb.ObjectRef{
			Name:      kustomizationObjectName,
			Namespace: ns.Name,
			Kind:      kustomizev1.KustomizationKind,
		},
	})
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(res.Events).To(HaveLen(1))

	g.Expect(res.Events[0].Component).To(Equal(kustomizeEvent.Source.Component))

	// Get helmrelease events
	res, err = c.ListEvents(ctx, &pb.ListEventsRequest{
		InvolvedObject: &pb.ObjectRef{
			Name:      helmObjectName,
			Namespace: ns.Name,
			Kind:      helmv2.HelmReleaseKind,
		},
	})
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(res.Events).To(HaveLen(1))

	g.Expect(res.Events[0].Component).To(Equal(helmEvent.Source.Component))
}
