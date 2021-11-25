package charts

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

const (
	testNamespace  = "testing"
	testSecretName = "https-credentials"
)

var _ ChartClient = (*HelmChartClient)(nil)

func TestUpdateCache_with_bad_url(t *testing.T) {
	hr := makeTestHelmRepository("http://[::1]:namedport/index.yaml")
	cc := NewHelmChartClient(makeTestClient(t), testNamespace, hr)

	err := cc.UpdateCache(context.TODO())
	//TODO fix
	// test.AssertErrorMatch(t, "invalid chart URL format", err)
	assert.NotNil(t, err)
}

func TestUpdateCache_with_missing_missing_secret_for_auth(t *testing.T) {
	fc := makeTestClient(t)
	ts := httptest.NewServer(basicAuthHandler(makeServeMux(t), "test", "password"))
	hr := makeTestHelmRepository(ts.URL, func(hr *sourcev1beta1.HelmRepository) {
		hr.Spec.SecretRef = &fluxmeta.LocalObjectReference{
			Name: testSecretName,
		}
	})
	tempDir, err := ioutil.TempDir("", "prefix")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatal(err)
		}
	})
	cc := NewHelmChartClient(fc, testNamespace, hr, WithCacheDir(tempDir))

	err = cc.UpdateCache(context.TODO())
	//test.AssertErrorMatch(t, `repository authentication: secrets "https-credentials" not found`, err)
	assert.NotNil(t, err)

}

func TestFileFromChart(t *testing.T) {
	ts := httptest.NewServer(makeServeMux(t))
	hr := makeTestHelmRepository(ts.URL)
	c := &ChartReference{Chart: "demo-profile", Version: "0.0.1", SourceRef: referenceForRepository(hr)}
	cc := makeChartClient(t, makeTestClient(t), hr)

	values, err := cc.FileFromChart(context.TODO(), c, "values.yaml")
	if err != nil {
		t.Fatal(err)
	}

	want := []byte("favoriteDrink: coffee\n")
	if diff := cmp.Diff(want, values); diff != "" {
		t.Fatalf("failed to get values:\n%s", diff)
	}
}

func TestFileFromChart_with_unknown_name(t *testing.T) {
	ts := httptest.NewServer(makeServeMux(t))
	hr := makeTestHelmRepository(ts.URL)
	c := &ChartReference{Chart: "demo-profile", Version: "0.0.1", SourceRef: referenceForRepository(hr)}
	cc := makeChartClient(t, makeTestClient(t), hr)

	_, err := cc.FileFromChart(context.TODO(), c, "unknown.yaml")
	//test.AssertErrorMatch(t, `failed to find file: unknown.yaml`, err)
	assert.NotNil(t, err)

}

func TestFileFromChart_missing_version(t *testing.T) {
	ts := httptest.NewServer(makeServeMux(t))
	hr := makeTestHelmRepository(ts.URL)
	c := &ChartReference{Chart: "demo-profile", Version: "0.0.2", SourceRef: referenceForRepository(hr)}
	cc := makeChartClient(t, makeTestClient(t), hr)

	_, err := cc.FileFromChart(context.TODO(), c, "values.yaml")
	//test.AssertErrorMatch(t, `chart "demo-profile" version "0.0.2" not found`, err)
	assert.NotNil(t, err)
}

func TestFileFromChart_missing_chart(t *testing.T) {
	ts := httptest.NewServer(makeServeMux(t, func(ri *repo.IndexFile) {
		ri.Entries["demo-profile"][0].Metadata.Version = "0.0.2"
		ri.Entries["demo-profile"][0].URLs = nil
	}))
	hr := makeTestHelmRepository(ts.URL)
	c := &ChartReference{Chart: "demo-profile", Version: "0.0.2", SourceRef: referenceForRepository(hr)}
	cc := makeChartClient(t, makeTestClient(t), hr)

	_, err := cc.FileFromChart(context.TODO(), c, "values.yaml")
	//test.AssertErrorMatch(t, `chart "demo-profile" version "0.0.2" has no downloadable URLs`, err)
	assert.NotNil(t, err)
}

func TestParseValues(t *testing.T) {
	ts := httptest.NewServer(makeServeMux(t, func(ri *repo.IndexFile) {
		ri.Entries["demo-profile"][0].Metadata.Version = "0.0.2"
		ri.Entries["demo-profile"][0].URLs = nil
	}))
	hr := makeTestHelmRepository(ts.URL)
	f, err := os.ReadFile("testdata/parsing/values.yaml")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	values := base64.StdEncoding.EncodeToString(f)
	res, err := ParseValues("podinfo", "0.0.2", values, "dev", hr)
	if err != nil {
		t.Fatalf("failed to parse profile values:\n%s", err)
	}
	actual, _ := yaml.Marshal(res)
	expected, _ := os.ReadFile("testdata/parsing/profile.yaml")
	if diff := cmp.Diff(expected, actual, protocmp.Transform()); diff != "" {
		t.Fatalf("Helm release didn't match expected:\n%s", diff)
	}
}

func makeServeMux(t *testing.T, opts ...func(*repo.IndexFile)) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/charts/index.yaml", func(w http.ResponseWriter, req *http.Request) {
		b, err := yaml.Marshal(makeTestChartIndex(opts...))
		if err != nil {
			t.Fatal(err)
		}
		w.Write(b)
	})
	mux.Handle("/", http.FileServer(http.Dir("testdata")))
	return mux
}

func referenceForRepository(s *sourcev1beta1.HelmRepository) helmv2beta1.CrossNamespaceObjectReference {
	return helmv2beta1.CrossNamespaceObjectReference{
		APIVersion: s.TypeMeta.APIVersion,
		Kind:       s.TypeMeta.Kind,
		Name:       s.ObjectMeta.Name,
		Namespace:  s.ObjectMeta.Name,
	}
}

func makeTestChartIndex(opts ...func(*repo.IndexFile)) *repo.IndexFile {
	ri := &repo.IndexFile{
		APIVersion: "v1",
		Entries: map[string]repo.ChartVersions{
			"demo-profile": repo.ChartVersions{
				{
					Metadata: &chart.Metadata{
						Annotations: map[string]string{
							ProfileAnnotation: "demo-profile",
						},
						Description: "Simple demo profile",
						Home:        "https://example.com/testing",
						Name:        "demo-profile",
						Sources: []string{
							"https://example.com/testing",
						},
						Version: "0.0.1",
					},
					Created: time.Now(),
					Digest:  "aaff4545f79d8b2913a10cb400ebb6fa9c77fe813287afbacf1a0b897cdffffff",
					URLs: []string{
						"/charts/demo-profile-0.1.0.tgz",
					},
				},
			},
		},
	}
	for _, o := range opts {
		o(ri)
	}
	return ri
}

func basicAuthHandler(next http.Handler, user, pass string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if ok && (u == user && p == pass) {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="test"`))
		w.WriteHeader(401)
		w.Write([]byte("401 Unauthorized\n"))
	})
}

func makeTestClient(t *testing.T, objs ...runtime.Object) client.Client {
	t.Helper()
	s := runtime.NewScheme()
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	return fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(objs...).Build()
}

// Based on https://fluxcd.io/docs/components/source/helmrepositories/
func makeTestSecret(user, pass string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		Type: corev1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      testSecretName,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			"username": []byte(user),
			"password": []byte(pass),
		},
	}
}

func makeChartClient(t *testing.T, cl client.Client, hr *sourcev1beta1.HelmRepository) *HelmChartClient {
	t.Helper()
	tempDir, err := ioutil.TempDir("", "prefix")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatal(err)
		}
	})
	cc := NewHelmChartClient(cl, testNamespace, hr, WithCacheDir(tempDir))
	if err := cc.UpdateCache(context.TODO()); err != nil {
		t.Fatal(err)
	}
	return cc
}
