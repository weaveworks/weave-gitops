package cache_test

import (
	"context"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func newNamespace(ctx context.Context, prefix string, g *GomegaWithT) *corev1.Namespace {
	ns := &corev1.Namespace{}
	ns.Name = prefix + "kube-test-" + rand.String(5)

	g.Expect(k8sEnv.Client.Create(ctx, ns)).To(Succeed())

	return ns
}
