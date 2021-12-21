package profiles_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/testing"

	"github.com/weaveworks/weave-gitops/pkg/profiles"
)

var _ = Describe("GetProfiles", func() {
	var (
		buffer    *gbytes.Buffer
		clientSet *fake.Clientset
	)

	BeforeEach(func() {
		buffer = gbytes.NewBuffer()
		clientSet = fake.NewSimpleClientset()
	})

	It("prints the available profiles", func() {
		clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
			return true, newFakeResponseWrapper(getProfilesResp), nil
		})

		Expect(profiles.GetProfiles(context.TODO(), profiles.GetOptions{
			Namespace: "test-namespace",
			ClientSet: clientSet,
			Writer:    buffer,
			Port:      "9001",
		})).To(Succeed())

		Expect(string(buffer.Contents())).To(Equal(`NAME	DESCRIPTION	AVAILABLE_VERSIONS
podinfo	Podinfo Helm chart for Kubernetes	6.0.0,6.0.1
`))
	})

	When("the response isn't valid", func() {
		It("errors", func() {
			clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
				return true, newFakeResponseWrapper("not=json"), nil
			})

			err := profiles.GetProfiles(context.TODO(), profiles.GetOptions{
				Namespace: "test-namespace",
				ClientSet: clientSet,
				Writer:    buffer,
				Port:      "9001",
			})
			Expect(err).To(MatchError(ContainSubstring("failed to unmarshal response")))
		})
	})

	When("making the request fails", func() {
		It("errors", func() {
			clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
				return true, newFakeResponseWrapperWithErr("nope"), nil
			})

			err := profiles.GetProfiles(context.TODO(), profiles.GetOptions{
				Namespace: "test-namespace",
				ClientSet: clientSet,
				Writer:    buffer,
				Port:      "9001",
			})
			Expect(err).To(MatchError("failed to make GET request to service test-namespace/wego-app path \"/v1/profiles\": nope"))
		})
	})

	When("the request returns non-200", func() {
		It("errors", func() {
			clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
				return true, newFakeResponseWrapperWithStatusCode(http.StatusNotFound), nil
			})

			err := profiles.GetProfiles(context.TODO(), profiles.GetOptions{
				Namespace: "test-namespace",
				ClientSet: clientSet,
				Writer:    buffer,
				Port:      "9001",
			})
			Expect(err).To(MatchError("failed to make GET request to service test-namespace/wego-app path \"/v1/profiles\" status code: 404"))
		})
	})
})
