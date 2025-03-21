package testutils

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/weaveworks/weave-gitops/pkg/kube"
)

var k8sEnv *K8sTestEnv

type K8sTestEnv struct {
	Env        *envtest.Environment
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
		CRDDirectoryPaths:     crdPaths,
		ErrorIfCRDPathMissing: true,
		CRDInstallOptions: envtest.CRDInstallOptions{
			CleanUpAfterUse: false,
		},
	}

	var err error
	cfg, err := testEnv.Start()
	if err != nil {
		return nil, fmt.Errorf("could not start testEnv: %w", err)
	}

	scheme, err := kube.CreateScheme()
	if err != nil {
		return nil, fmt.Errorf("could not create scheme: %w", err)
	}

	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize discovery client: %w", err)
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize dynamic client: %w", err)
	}

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	k8sEnv = &K8sTestEnv{
		Env:        testEnv,
		Client:     k8sClient,
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
func MakeJWToken(t *testing.T, key *rsa.PrivateKey, email string, opts ...func(map[string]any)) string {
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
		Expiry:    jwt.NewNumericDate(notBefore.Add(maxAgeSecondsAuthCookie)),
	}
	extraClaims := map[string]any{
		"groups":             []string{"testing"},
		"email":              email,
		"preferred_username": "testing",
	}

	for _, opt := range opts {
		opt(extraClaims)
	}

	signed, err := jwt.Signed(signer).Claims(claims).Claims(extraClaims).Serialize()
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

// DeleteAllOf loops through all namespaces and deletes all resources from the given type
func DeleteAllOf(g *gomega.GomegaWithT, obj client.Object) {
	ctx := context.Background()

	nss := &corev1.NamespaceList{}
	g.Expect(k8sEnv.Client.List(ctx, nss)).To(gomega.Succeed())

	for _, ns := range nss.Items {
		g.Expect(k8sEnv.Client.DeleteAllOf(ctx, obj, client.InNamespace(ns.Name))).To(gomega.Succeed())
	}
}

// DeleteNamespace deletes a namespace.
// Note: deleting a namespace using this function on tests wont delete the underlying resources
// like in a real environment would.
func DeleteNamespace(g *gomega.GomegaWithT, ns *corev1.Namespace) {
	// Code borrowed from controller-runtime: https://github.com/kubernetes-sigs/controller-runtime/blob/eb39b8eb28cfe920fa2450eb38f814fc9e8003e8/pkg/client/client_test.go#L51
	clientset, err := kubernetes.NewForConfig(k8sEnv.Rest)
	g.Expect(err).To(gomega.BeNil())

	ctx := context.Background()

	ns, err = clientset.CoreV1().Namespaces().Get(ctx, ns.Name, metav1.GetOptions{})
	if err != nil {
		return
	}

	err = clientset.CoreV1().Namespaces().Delete(ctx, ns.Name, metav1.DeleteOptions{})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// finalize if necessary
	pos := -1
	finalizers := ns.Spec.Finalizers

	for i, fin := range finalizers {
		if fin == "kubernetes" {
			pos = i
			break
		}
	}

	if pos == -1 {
		// no need to finalize
		return
	}

	// re-get in order to finalize
	ns, err = clientset.CoreV1().Namespaces().Get(ctx, ns.Name, metav1.GetOptions{})
	if err != nil {
		return
	}

	ns.Spec.Finalizers = append(finalizers[:pos], finalizers[pos+1:]...)
	_, err = clientset.CoreV1().Namespaces().Finalize(ctx, ns, metav1.UpdateOptions{})
	g.Expect(err).NotTo(gomega.HaveOccurred())

WAIT_LOOP:
	for i := 0; i < 10; i++ {
		ns, err = clientset.CoreV1().Namespaces().Get(ctx, ns.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			// success!
			return
		}
		select {
		case <-ctx.Done():
			break WAIT_LOOP
			// failed to delete in time, see failure below
		case <-time.After(100 * time.Millisecond):
			// do nothing, try again
		}
	}
	g.Fail(fmt.Sprintf("timed out waiting for namespace %q to be deleted", ns.Name))
}
