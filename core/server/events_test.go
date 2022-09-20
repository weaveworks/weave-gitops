package server_test

import (
	"context"
	"fmt"

	"testing"

	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: scheme,
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
			Kind:      pb.Kind_Kustomization,
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
			Kind:      pb.Kind_HelmRelease,
		},
	})
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(res.Events).To(HaveLen(1))

	g.Expect(res.Events[0].Component).To(Equal(helmEvent.Source.Component))
}

func newNamespace(ctx context.Context, k client.Client, g *GomegaWithT) *corev1.Namespace {
	ns := &corev1.Namespace{}
	ns.Name = "kube-test-" + rand.String(5)

	g.Expect(k.Create(ctx, ns)).To(Succeed())

	return ns
}
