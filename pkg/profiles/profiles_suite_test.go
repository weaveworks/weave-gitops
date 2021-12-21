package profiles_test

import (
	"context"
	"fmt"
	"io"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
      "helmRepository": null,
      "availableVersions": [
        "6.0.0",
        "6.0.1"
      ]
    }
  ]
}
`

func TestProfiles(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Profiles Suite")
}

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
