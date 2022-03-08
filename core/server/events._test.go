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
)

func TestListFluxEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	eventObjectName := "my-kustomization"
	ns := newNamespace(ctx, k, g)

	event := &corev1.Event{
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprintf("%s.16da7d2e2c5b0930", eventObjectName),
			Namespace: ns.Name,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "kustomization",
			Namespace: ns.Name,
			Name:      eventObjectName,
		},
		Source: corev1.EventSource{
			Component: "kustomize-controller",
		},
	}

	// An event we don't care about. Shoud not show up in our response.
	otherEvent := &corev1.Event{
		ObjectMeta: v1.ObjectMeta{
			Name:      "someotherobject.16da7d2e2c5b0930",
			Namespace: ns.Name,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "kustomization",
			Namespace: ns.Name,
			Name:      "someotherobject",
		},
		Source: corev1.EventSource{
			Component: "kustomize-controller",
		},
	}

	g.Expect(k.Create(ctx, event)).To(Succeed())
	g.Expect(k.Create(ctx, otherEvent)).To(Succeed())

	res, err := c.ListFluxEvents(ctx, &pb.ListFluxEventsRequest{
		Namespace: ns.Name,
		InvolvedObject: &pb.ObjectReference{
			Name:      eventObjectName,
			Namespace: ns.Name,
			Kind:      "kustomization",
		},
	})
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(res.Events).To(HaveLen(1))

	g.Expect(res.Events[0].Component).To(Equal(event.Source.Component))
}
