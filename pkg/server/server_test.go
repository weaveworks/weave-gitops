package server_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/api/applications"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ApplicationsServer", func() {
	It("AddApplication", func() {
		kubeClient.GetApplicationsStub = func(ctx context.Context, ns string) ([]wego.Application, error) {
			return []wego.Application{
				{
					ObjectMeta: v1.ObjectMeta{Name: "my-app"},
					Spec:       wego.ApplicationSpec{Path: "bar"},
				},
				{
					ObjectMeta: v1.ObjectMeta{Name: "my-app1"},
					Spec:       wego.ApplicationSpec{Path: "bar2"},
				},
			}, nil
		}

		res, err := client.ListApplications(context.Background(), &applications.ListApplicationsRequest{})

		Expect(err).NotTo(HaveOccurred())

		Expect(len(res.Applications)).To(Equal(2))
	})
	It("GetApplication", func() {
		kubeClient.GetApplicationStub = func(ctx context.Context, name string) (*wego.Application, error) {
			return &wego.Application{
				ObjectMeta: v1.ObjectMeta{Name: "my-app"},
				Spec:       wego.ApplicationSpec{Path: "bar"},
			}, nil
		}

		res, err := client.GetApplication(context.Background(), &applications.GetApplicationRequest{ApplicationName: "my-app"})
		Expect(err).NotTo(HaveOccurred())

		Expect(res.Application.Name).To(Equal("my-app"))
	})
})
