package profiles_test

import (
	"context"
	"fmt"
	"io"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/testing"

	"github.com/weaveworks/weave-gitops/pkg/services/profiles"
)

const getProfilesResp = `{
  "profiles": [
    {
      "name": "podinfo",
      "home": "https://github.com/stefanprodan/podinfo",
      "sources": [
        "https://github.com/stefanprodan/podinfo"
      ],
      "description": "Podinfo Helm chart for Kubernetes",
      "keywords": [],
      "maintainers": [
        {
          "name": "stefanprodan",
          "email": "stefanprodan@users.noreply.github.com",
          "url": ""
        }
      ],
      "icon": "",
      "annotations": {},
      "kubeVersion": ">=1.19.0-0",
      "helmRepository": {
		  "name": "podinfo",
		  "namespace": "weave-system"
	  },
      "availableVersions": [
        "6.0.0",
        "6.0.1"
      ]
    }
  ]
}
`

var _ = Describe("GetProfiles", func() {
	var (
		buffer      *gbytes.Buffer
		clientSet   *fake.Clientset
		profilesSvc *profiles.ProfilesSvc
	)

	BeforeEach(func() {
		buffer = gbytes.NewBuffer()
		clientSet = fake.NewSimpleClientset()
		profilesSvc = profiles.NewService(clientSet)
	})

	It("prints the available profiles", func() {
		clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
			return true, newFakeResponseWrapper(getProfilesResp), nil
		})

		Expect(profilesSvc.Get(context.TODO(), profiles.GetOptions{
			Namespace: "test-namespace",
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

			err := profilesSvc.Get(context.TODO(), profiles.GetOptions{
				Namespace: "test-namespace",
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

			err := profilesSvc.Get(context.TODO(), profiles.GetOptions{
				Namespace: "test-namespace",
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

			err := profilesSvc.Get(context.TODO(), profiles.GetOptions{
				Namespace: "test-namespace",
				Writer:    buffer,
				Port:      "9001",
			})
			Expect(err).To(MatchError("failed to make GET request to service test-namespace/wego-app path \"/v1/profiles\" status code: 404"))
		})
	})
})

func (w fakeResponseWrapper) DoRaw(context.Context) ([]byte, error) {
	return w.raw, w.err
}

func (w fakeResponseWrapper) Stream(context.Context) (io.ReadCloser, error) {
	fmt.Println("stream called")
	return nil, nil
}

func newFakeResponseWrapper(raw string) fakeResponseWrapper {
	return fakeResponseWrapper{raw: []byte(raw)}
}

func newFakeResponseWrapperWithErr(err string) fakeResponseWrapper {
	return fakeResponseWrapper{err: fmt.Errorf(err)}
}

func newFakeResponseWrapperWithStatusCode(code int32) fakeResponseWrapper {
	return fakeResponseWrapper{err: &errors.StatusError{ErrStatus: metav1.Status{Code: code}}}
}

type fakeResponseWrapper struct {
	raw []byte
	err error
}
