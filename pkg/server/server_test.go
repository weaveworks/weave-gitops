package server_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"golang.org/x/oauth2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("ApplicationsServer", func() {
	It("ListApplications", func() {
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
		kubeClient.GetApplicationStub = func(ctx context.Context, name types.NamespacedName) (*wego.Application, error) {
			return &wego.Application{
				ObjectMeta: v1.ObjectMeta{Name: "my-app"},
				Spec:       wego.ApplicationSpec{Path: "bar"},
			}, nil
		}

		res, err := client.GetApplication(context.Background(), &applications.GetApplicationRequest{Name: "my-app"})
		Expect(err).NotTo(HaveOccurred())

		Expect(res.Application.Name).To(Equal("my-app"))
	})
	It("GetAuthenticationProviders", func() {
		res, err := client.GetAuthenticationProviders(context.Background(), &applications.GetAuthenticationProvidersRequest{})
		Expect(err).NotTo(HaveOccurred())

		expected := &applications.OauthProvider{
			Name: string(gitproviders.ProviderNameGithub),
		}

		Expect(res.Providers).Should(ContainElement(expected))
	})
	It("Authenticate", func() {
		token := "def456xyz"
		fakeOauthProvider.ExchangeStub = func(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
			return &oauth2.Token{AccessToken: token}, nil
		}
		body := &applications.AuthenticateRequest{
			ProviderName: string(gitproviders.ProviderNameGithub),
			Code:         "abc123supersecretcode",
		}

		res, err := client.Authenticate(context.Background(), body)
		Expect(err).NotTo(HaveOccurred())

		Expect(res.Token).To(Equal(token))
	})
})
