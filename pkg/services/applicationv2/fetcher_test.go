package applicationv2

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Fetcher", func() {
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

			Expect(result.Name).To(Equal(app.Name))
		})
		It("it returns a not-found error when an app doesn't exist", func() {
			ctx := context.Background()
			k8s := fake.NewClientBuilder().WithScheme(kube.CreateScheme()).Build()

			gs := NewFetcher(k8s)

			_, err := gs.Get(ctx, "foo", "ns")
			Expect(err).To(Equal(ErrNotFound))
		})
	})

	Describe(".List()", func() {
		It("lists multiple applications", func() {
			ctx := context.Background()
			k8s := fake.NewClientBuilder().WithScheme(kube.CreateScheme()).Build()

			app := &wego.Application{}
			app.Name = "my-app"
			app.Namespace = "some-ns"

			Expect(k8s.Create(ctx, app)).To(Succeed())

			app2 := &wego.Application{}
			app2.Name = "my-app2"
			app2.Namespace = "some-ns"

			Expect(k8s.Create(ctx, app2)).To(Succeed())

			gs := NewFetcher(k8s)

			result, err := gs.List(ctx, app.Namespace)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(result)).To(Equal(2))
		})
		It("lists does not list an application in a different namespace", func() {
			ctx := context.Background()
			k8s := fake.NewClientBuilder().WithScheme(kube.CreateScheme()).Build()

			app := &wego.Application{}
			app.Name = "my-app"
			app.Namespace = "some-ns"

			Expect(k8s.Create(ctx, app)).To(Succeed())

			app2 := &wego.Application{}
			app2.Name = "my-app2"
			app2.Namespace = "a-diff-ns"

			Expect(k8s.Create(ctx, app2)).To(Succeed())

			gs := NewFetcher(k8s)

			result, err := gs.List(ctx, app.Namespace)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(result)).To(Equal(1))
		})
		It("lists an empty application list", func() {
			ctx := context.Background()
			k8s := fake.NewClientBuilder().WithScheme(kube.CreateScheme()).Build()

			gs := NewFetcher(k8s)

			result, err := gs.List(ctx, "ns")
			Expect(err).NotTo(HaveOccurred())
			Expect(len(result)).To(Equal(0))
		})
	})
})
