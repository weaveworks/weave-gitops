package applicationv2

import (
	"context"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Getter", func() {
	Describe(".Get()", func() {
		It("gets an application that exists", func() {
			ctx := context.Background()
			k8s := fake.NewClientBuilder().WithScheme(kube.CreateScheme()).Build()

			app := &wego.Application{}
			app.Name = "my-app"
			app.Namespace = "some-ns"

			Expect(k8s.Create(ctx, app)).To(Succeed())

			gs := NewFetcher(k8s)

			result, err := gs.Get(ctx, app.Name, app.Namespace)
			Expect(err).NotTo(HaveOccurred())

			expected := models.Application{
				Name:      app.Name,
				Namespace: app.Namespace,
			}

			diff := cmp.Diff(expected, result)

			if diff != "" {
				GinkgoT().Errorf("mismatch (-actual, +expected):\n%s", diff)
			}
		})
		It("it returns a not-found error when an app doesn't exist", func() {
			ctx := context.Background()
			k8s := fake.NewClientBuilder().WithScheme(kube.CreateScheme()).Build()

			gs := NewFetcher(k8s)

			result, err := gs.Get(ctx, "foo", "ns")
			Expect(err).To(Equal(ErrNotFound))
			Expect(result).To(BeNil())
		})
	})
})
