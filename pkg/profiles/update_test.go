package profiles_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/weaveworks/weave-gitops/pkg/profiles"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

var helmReleaseTypeMeta = metav1.TypeMeta{
	Kind:       "HelmRelease",
	APIVersion: "helm.toolkit.fluxcd.io/v2beta1",
}

var _ = Describe("UpdateProfile", func() {
	var (
		buffer          *gbytes.Buffer
		clientSet       *fake.Clientset
		tmpDir          string
		profileFilepath string
	)

	BeforeEach(func() {
		buffer = gbytes.NewBuffer()
		clientSet = fake.NewSimpleClientset()
		var err error
		tmpDir, err = ioutil.TempDir("", "update-profile")
		Expect(err).NotTo(HaveOccurred())
		profileFilepath = filepath.Join(tmpDir, "podinfo.yaml")
		writeResource(&helmv2.HelmRelease{
			TypeMeta: helmReleaseTypeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:      "podinfo",
				Namespace: "default",
			},
			Spec: helmv2.HelmReleaseSpec{
				ReleaseName: "podinfo",
				Chart: helmv2.HelmChartTemplate{
					Spec: helmv2.HelmChartTemplateSpec{
						Version: "v0.1.0",
					},
				},
			},
		}, profileFilepath)
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	It("updates the profile to the desired version", func() {
		clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
			return true, newFakeResponseWrapper(getProfilesResp), nil
		})

		err := profiles.UpdateProfile(context.TODO(), profiles.UpdateOptions{
			Namespace:       "test-namespace",
			Port:            "9001",
			ProfileFilepath: profileFilepath,
			ProfileName:     "podinfo",
			ProfileVersion:  "6.0.0",
			ClientSet:       clientSet,
			Writer:          buffer,
		})

		Expect(err).NotTo(HaveOccurred())
		helmRes := helmv2.HelmRelease{}
		decodeFile(filepath.Join(tmpDir, "podinfo.yaml"), &helmRes)
		Expect(helmRes).To(Equal(helmv2.HelmRelease{
			TypeMeta: helmReleaseTypeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:      "podinfo",
				Namespace: "default",
			},
			Spec: helmv2.HelmReleaseSpec{
				ReleaseName: "podinfo",
				Chart: helmv2.HelmChartTemplate{
					Spec: helmv2.HelmChartTemplateSpec{
						Version: "6.0.0",
					},
				},
			},
		}))

		lines := strings.Split(string(buffer.Contents()), "\n")
		Expect(lines).To(ConsistOf(
			"Updating profile podinfo to version 6.0.0",
		))
	})

	When("the desired version doesn't exist", func() {
		It("errors", func() {
			clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
				return true, newFakeResponseWrapper(getProfilesResp), nil
			})

			err := profiles.UpdateProfile(context.TODO(), profiles.UpdateOptions{
				Namespace:       "test-namespace",
				Port:            "9001",
				ProfileFilepath: profileFilepath,
				ProfileName:     "podinfo",
				ProfileVersion:  "6.0.2",
				ClientSet:       clientSet,
				Writer:          buffer,
			})

			Expect(err).To(MatchError(`version "6.0.2" is not available for profile "podinfo". Available versions: 6.0.0,6.0.1`))
		})
	})

	When("getting the profiles fails", func() {
		It("errors", func() {
			clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
				return true, newFakeResponseWrapperWithErr("nope"), nil
			})

			err := profiles.UpdateProfile(context.TODO(), profiles.UpdateOptions{
				Namespace:       "test-namespace",
				Port:            "9001",
				ProfileFilepath: profileFilepath,
				ProfileName:     "dontexist",
				ProfileVersion:  "6.0.2",
				ClientSet:       clientSet,
				Writer:          buffer,
			})

			Expect(err).To(MatchError(ContainSubstring("failed to get available profiles: ")))
		})
	})

	When("the profile doesn't exist", func() {
		It("errors", func() {
			clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
				return true, newFakeResponseWrapper(getProfilesResp), nil
			})

			err := profiles.UpdateProfile(context.TODO(), profiles.UpdateOptions{
				Namespace:       "test-namespace",
				Port:            "9001",
				ProfileFilepath: profileFilepath,
				ProfileName:     "dontexist",
				ProfileVersion:  "6.0.2",
				ClientSet:       clientSet,
				Writer:          buffer,
			})

			Expect(err).To(MatchError(`profile "dontexist" does not exist. Run "gitops get profiles" to see available profiles`))
		})
	})

	When("the profile file doesn't exist", func() {
		It("errors", func() {
			err := profiles.UpdateProfile(context.TODO(), profiles.UpdateOptions{
				Namespace:       "test-namespace",
				Port:            "9001",
				ProfileFilepath: "/tmp/foo/bar/baz/lol",
				ProfileName:     "podinfo",
				ProfileVersion:  "6.0.0",
				ClientSet:       clientSet,
				Writer:          buffer,
			})

			Expect(err).To(MatchError(ContainSubstring(`error reading profile "podinfo" at path "/tmp/foo/bar/baz/lol": `)))
		})
	})

	When("the profile file isn't a helm release", func() {
		BeforeEach(func() {
			Expect(ioutil.WriteFile(profileFilepath, []byte("not-{yaml"), 0750)).To(Succeed())
		})

		It("errors", func() {
			err := profiles.UpdateProfile(context.TODO(), profiles.UpdateOptions{
				Namespace:       "test-namespace",
				Port:            "9001",
				ProfileFilepath: profileFilepath,
				ProfileName:     "podinfo",
				ProfileVersion:  "6.0.0",
				ClientSet:       clientSet,
				Writer:          buffer,
			})

			Expect(err).To(MatchError(ContainSubstring(`error unmarshaling`)))
		})
	})
})

func decodeFile(filepath string, obj interface{}) {
	content, err := ioutil.ReadFile(filepath)
	Expect(err).NotTo(HaveOccurred())

	err = yaml.NewYAMLOrJSONDecoder(bytes.NewReader(content), 4096).Decode(obj)
	Expect(err).NotTo(HaveOccurred())
}

func writeResource(obj runtime.Object, filename string) {
	e := kjson.NewSerializerWithOptions(kjson.DefaultMetaFactory, nil, nil, kjson.SerializerOptions{Yaml: true, Strict: true})
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	Expect(err).NotTo(HaveOccurred())

	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	Expect(e.Encode(obj, f)).To(Succeed())
}
