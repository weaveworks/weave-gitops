package app_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8sApps "github.com/weaveworks/weave-gitops/api/v1alpha"
)

var _ = Describe("Get", func() {
	It("gets an app", func() {
		kubeClient.GetApplicationStub = func(name string) (*k8sApps.Application, error) {
			return &k8sApps.Application{
				Spec: k8sApps.ApplicationSpec{Foo: "bar"},
			}, nil
		}

		a, err := appSrv.Get(defaultParams.Name)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(a.Spec.Foo).To(Equal("bar"))
	})
})
