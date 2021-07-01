package app_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
)

var _ = Describe("Get", func() {
	It("gets an app", func() {
		kubeClient.GetApplicationStub = func(ctx context.Context, name string) (*wego.Application, error) {
			return &wego.Application{
				Spec: wego.ApplicationSpec{Path: "bar"},
			}, nil
		}

		a, err := appSrv.Get(defaultParams.Name)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(a.Spec.Path).To(Equal("bar"))
	})
})
