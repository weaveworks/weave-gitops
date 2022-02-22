package app

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Get", func() {
	It("gets an app", func() {
		kubeClient.GetApplicationStub = func(ctx context.Context, name types.NamespacedName) (*wego.Application, error) {
			return &wego.Application{
				Spec: wego.ApplicationSpec{Path: "bar"},
			}, nil
		}

		a, err := appSrv.Get(types.NamespacedName{Name: "my-app", Namespace: "my-namespace"})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(a.Spec.Path).To(Equal("bar"))
	})
})
