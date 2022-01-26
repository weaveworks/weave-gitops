package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
)

const (
	fluxSystem            = "flux-system"
	fluxNamespaceManifest = `
apiVersion: v1
kind: Namespace
metadata:
  name: flux-system
`
)

type appServerFixture struct {
	*GomegaWithT
	env        *testutils.K8sTestEnv
	testServer *httptest.Server
}

func (f appServerFixture) arrangeAppPath(namespace string) string {
	return fmt.Sprintf("%s/v1/namespace/%s/app", f.testServer.URL, namespace)
}

func (f appServerFixture) arrangeFluxSystemNamespace(t *testing.T) {
	obj := &unstructured.Unstructured{}

	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	_, gvk, err := dec.Decode([]byte(fluxNamespaceManifest), nil, obj)
	if err != nil {
		t.Fatalf("could not decode manifest: %s", err.Error())
	}

	mapper := f.env.RestMapper

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		t.Fatalf("could not get rest mapping: %s", err.Error())
	}

	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = f.env.DynClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		dr = f.env.DynClient.Resource(mapping.Resource)
	}

	_, err = dr.Create(context.Background(), obj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("could not create resource: %s", err.Error())
	}
}

func (f appServerFixture) cleanUpFixture(t *testing.T) {
	f.env.Stop()
}

func setUpAppServerTest(t *testing.T) appServerFixture {
	os.Setenv("KUBEBUILDER_ASSETS", "../../tools/bin/envtest")

	env, err := testutils.StartK8sTestEnvironment([]string{
		"../../manifests/crds",
		"../../tools/testcrds",
	})
	if err != nil {
		t.Errorf("could not start testEnv: %s", err)
	}

	s := runtime.NewServeMux()

	_ = Hydrate(context.Background(), s, env.Rest)
	testServer := httptest.NewServer(s)

	return appServerFixture{
		testServer:  testServer,
		env:         env,
		GomegaWithT: NewGomegaWithT(t),
	}
}

func TestAppServer(t *testing.T) {
	f := setUpAppServerTest(t)

	res, _ := http.Get(f.testServer.URL + "/v1/namespace/flux-system/app")
	out, _ := ioutil.ReadAll(res.Body)
	f.Expect(string(out)).To(Equal(`{"apps":[]}`))
	f.Expect(res.StatusCode).To(Equal(http.StatusOK))

	// App Requests
	appOne := &pb.AddAppRequest{
		Name:        "app-1",
		Namespace:   "flux-system",
		DisplayName: "App 1",
		Description: "The first of many",
	}
	data, _ := json.Marshal(appOne)

	const appOneJson = `{"namespace":"flux-system","name":"app-1","description":"The first of many","displayName":"App 1"}`

	const expectedAppOneJson = `{"success":true, "app":` + appOneJson + `}`

	// The namespace does not exist
	res, _ = http.Post(f.arrangeAppPath(fluxSystem), "application/json", bytes.NewReader(data))
	out, _ = ioutil.ReadAll(res.Body)
	f.Expect(string(out)).To(MatchJSON(`{"code":5, "message":"namespaces \"flux-system\" not found", "details":[]}`))
	f.Expect(res.StatusCode).To(Equal(http.StatusNotFound))

	f.arrangeFluxSystemNamespace(t)

	// Create app 1 now that the namespace exists
	res, _ = http.Post(f.arrangeAppPath(fluxSystem), "application/json", bytes.NewReader(data))
	out, _ = ioutil.ReadAll(res.Body)
	f.Expect(string(out)).To(MatchJSON(expectedAppOneJson))
	f.Expect(res.StatusCode).To(Equal(http.StatusOK))

	// Get list of apps again
	res, _ = http.Get(f.testServer.URL + "/v1/namespace/flux-system/app")
	out, _ = ioutil.ReadAll(res.Body)
	f.Expect(string(out)).To(MatchJSON(`{"apps":[` + appOneJson + `]}`))
	f.Expect(res.StatusCode).To(Equal(http.StatusOK))

	// Create app 2 now that the namespace exists
	appTwo := &pb.AddAppRequest{
		Name:        "app-2",
		Namespace:   "flux-system",
		DisplayName: "App 2",
		Description: "The second of many",
	}
	data, _ = json.Marshal(appTwo)

	const appTwoJson = `{"namespace":"flux-system","name":"app-2","description":"The second of many","displayName":"App 2"}`

	const expectedAppTwoJson = `{"success":true, "app":` + appTwoJson + `}`

	res, _ = http.Post(f.arrangeAppPath(fluxSystem), "application/json", bytes.NewReader(data))
	out, _ = ioutil.ReadAll(res.Body)
	f.Expect(string(out)).To(MatchJSON(expectedAppTwoJson))
	f.Expect(res.StatusCode).To(Equal(http.StatusOK))

	// Get list of apps again
	res, _ = http.Get(f.testServer.URL + "/v1/namespace/flux-system/app")
	out, _ = ioutil.ReadAll(res.Body)
	f.Expect(string(out)).To(MatchJSON(`{"apps":[` + appOneJson + `,` + appTwoJson + `]}`))
	f.Expect(res.StatusCode).To(Equal(http.StatusOK))

	// Get list of apps, wrong namespace
	res, _ = http.Get(f.testServer.URL + "/v1/namespace/gitops-system/app")
	out, _ = ioutil.ReadAll(res.Body)
	f.Expect(string(out)).To(MatchJSON(`{"apps":[]}`))
	f.Expect(res.StatusCode).To(Equal(http.StatusOK))

	// App Kustomizations Requests
	kustOne := &pb.AddKustomizationReq{Name: "app-1"}
	data, _ = json.Marshal(kustOne)

	res, _ = http.Post(f.testServer.URL+"/v1/namespace/flux-system/app/app-1/kustomization", "application/json", bytes.NewReader(data))
	f.Expect(res.StatusCode).To(Equal(http.StatusBadRequest))

	f.Expect(true).To((BeTrue()))
	f.cleanUpFixture(t)
}
