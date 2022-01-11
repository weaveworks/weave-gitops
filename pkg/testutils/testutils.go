package testutils

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/osys/osysfakes"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	fakelogr "github.com/weaveworks/weave-gitops/pkg/vendorfakes/logr"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/fluxcd/go-git-providers/gitprovider"
	kustomizev2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

var k8sEnv *K8sTestEnv

type K8sTestEnv struct {
	Client     client.Client
	DynClient  dynamic.Interface
	RestMapper *restmapper.DeferredDiscoveryRESTMapper
	Rest       *rest.Config
	Stop       func()
}

// Starts a local k8s test environment for testing Kubernetes operations such as Create, Get, Delete, etc.

// Note that crdPaths are relative to the path of the test file,
// NOT the current working directory or path that the tests were started from.
func StartK8sTestEnvironment(crdPaths []string) (*K8sTestEnv, error) {
	if k8sEnv != nil {
		return k8sEnv, nil
	}

	testEnv := &envtest.Environment{
		CRDDirectoryPaths: crdPaths,
		CRDInstallOptions: envtest.CRDInstallOptions{
			CleanUpAfterUse: false,
		},
	}

	var err error
	cfg, err := testEnv.Start()

	if err != nil {
		return nil, fmt.Errorf("could not start testEnv: %w", err)
	}

	scheme := kube.CreateScheme()

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		ClientDisableCacheFor: []client.Object{
			&wego.Application{},
			&corev1.Namespace{},
			&corev1.Secret{},
			&appsv1.Deployment{},
			&corev1.ConfigMap{},
			&kustomizev2.Kustomization{},
			&sourcev1.GitRepository{},
			&v1.CustomResourceDefinition{},
		},
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create controller manager: %w", err)
	}

	go func() {
		err := k8sManager.Start(ctrl.SetupSignalHandler())
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize discovery client: %s", err)
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize dynamic client: %s", err)
	}

	k8sEnv = &K8sTestEnv{
		Client:     k8sManager.GetClient(),
		DynClient:  dyn,
		RestMapper: mapper,
		Rest:       cfg,
		Stop: func() {
			err := testEnv.Stop()
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}

	return k8sEnv, nil
}

// MakeFakeLogr returns an API compliant logr object that can be used for unit testing.
// Without these stubs filled in, a nil pointer exception will be thrown on log.V().
func MakeFakeLogr() *fakelogr.FakeLogger {
	log := &fakelogr.FakeLogger{}
	log.WithValuesStub = func(i ...interface{}) logr.Logger {
		return log
	}
	log.VStub = func(i int) logr.Logger {
		return log
	}

	return log
}

type LocalFluxRunner struct {
	runner.Runner
}

func (r *LocalFluxRunner) Run(command string, args ...string) ([]byte, error) {
	cmd := "../flux/bin/flux"

	return r.Runner.Run(cmd, args...)
}

type DummyPullRequest struct {
}

func (d DummyPullRequest) Get() gitprovider.PullRequestInfo {
	return gitprovider.PullRequestInfo{WebURL: ""}
}

func (d DummyPullRequest) APIObject() interface{} {
	return nil
}

// Set up a flux binary in a temp dir that will be used to generate flux manifests
func SetupFlux() (flux.Flux, string, error) {
	dir, err := ioutil.TempDir("", "a-home-dir")
	if err != nil {
		return nil, "", err
	}

	cliRunner := &runner.CLIRunner{}
	osysClient := &osysfakes.FakeOsys{}
	realFlux := flux.New(osysClient, cliRunner)
	osysClient.UserHomeDirStub = func() (string, error) {
		return dir, nil
	}

	fluxBin, err := ioutil.ReadFile(filepath.Join("..", "..", "flux", "bin", "flux"))
	if err != nil {
		os.RemoveAll(dir)
		return nil, "", err
	}

	binPath, err := realFlux.GetBinPath()
	if err != nil {
		os.RemoveAll(dir)
		return nil, "", err
	}

	err = os.MkdirAll(binPath, 0777)
	if err != nil {
		os.RemoveAll(dir)
		return nil, "", err
	}

	exePath, err := realFlux.GetExePath()
	if err != nil {
		os.RemoveAll(dir)
		return nil, "", err
	}

	err = ioutil.WriteFile(exePath, fluxBin, 0777)
	if err != nil {
		os.RemoveAll(dir)
		return nil, "", err
	}

	return realFlux, dir, nil
}

func Setenv(k, v string) func() {
	prev := os.Environ()
	os.Setenv(k, v)

	return func() {
		os.Unsetenv(k)

		for _, kv := range prev {
			parts := strings.SplitN(kv, "=", 2)
			os.Setenv(parts[0], parts[1])
		}
	}
}

// MakeRSAPrivateKey generates and returns an RSA Private Key.
func MakeRSAPrivateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	k, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatal(err)
	}

	return k
}

// MakeJWToken creates and signs a token with the provided key.
func MakeJWToken(t *testing.T, key *rsa.PrivateKey, email string) string {
	t.Helper()

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, nil)
	if err != nil {
		t.Fatal(err)
	}

	maxAgeSecondsAuthCookie := time.Second * 600
	notBefore := time.Now().Add(-time.Second * 60)
	claims := jwt.Claims{
		Issuer:    "http://127.0.0.1:5556/dex",
		Subject:   "testing",
		Audience:  jwt.Audience{"test-service"},
		NotBefore: jwt.NewNumericDate(notBefore),
		IssuedAt:  jwt.NewNumericDate(notBefore),
		Expiry:    jwt.NewNumericDate(notBefore.Add(time.Duration(maxAgeSecondsAuthCookie))),
	}
	githubClaims := struct {
		Groups            []string `json:"groups"`
		Email             string   `json:"email"`
		PreferredUsername string   `json:"preferred_username"`
	}{
		[]string{"testing"},
		email,
		"example",
	}

	signed, err := jwt.Signed(signer).Claims(claims).Claims(githubClaims).CompactSerialize()
	if err != nil {
		t.Fatal(err)
	}

	return signed
}

// MakeKeysetServer starts an HTTP server that can serve JSONWebKey sets.
func MakeKeysetServer(t *testing.T, key *rsa.PrivateKey) *httptest.Server {
	t.Helper()

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var keys jose.JSONWebKeySet
		keys.Keys = []jose.JSONWebKey{
			{
				Key:       key.Public(),
				Use:       "sig",
				Algorithm: "RS256",
			},
		}
		_ = json.NewEncoder(w).Encode(keys)
	}))
	t.Cleanup(ts.Close)

	return ts
}
