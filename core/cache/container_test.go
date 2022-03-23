package cache_test

import (
	"context"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newNamespace(ctx context.Context, prefix string, k client.Client, g *GomegaWithT) *corev1.Namespace {
	ns := &corev1.Namespace{}
	ns.Name = prefix + "kube-test-" + rand.String(5)

	g.Expect(k.Create(ctx, ns)).To(Succeed())

	return ns
}
