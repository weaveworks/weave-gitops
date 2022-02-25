package tls_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	wegotls "github.com/weaveworks/weave-gitops/pkg/server/tls"
)

func TestGenerateKeyPair(t *testing.T) {
	tlsConfig, err := wegotls.TLSConfig([]string{"foo"})
	assert.NoError(t, err)

	require.Len(t, tlsConfig.Certificates, 1)
	cert, err := x509.ParseCertificate(tlsConfig.Certificates[0].Certificate[0])
	require.NoError(t, err)

	// Make sure DNS name is included
	assert.Equal(t, []string{"foo"}, cert.DNSNames)
	// Important for Chrome that we mark it for server auth
	assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, cert.ExtKeyUsage)
	// Serial number should be unique, (sorry if you randomly get 1!)
	assert.NotEqual(t, 1, cert.SerialNumber, "Maybe you randomly did get a 1?")
}

func TestTLSConfigCanBeServed(t *testing.T) {
	tlsConfig, err := wegotls.TLSConfig([]string{"127.0.0.1"})
	require.NoError(t, err)

	ts, client := getTestClient(t, tlsConfig)
	ts.StartTLS()

	defer ts.Close()

	res, err := client.Get(ts.URL)
	require.NoError(t, err)
	greeting, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	res.Body.Close()
	assert.Equal(t, "hello", string(greeting))
}

func TestTLSHostnameIsChecked(t *testing.T) {
	tlsConfig, err := wegotls.TLSConfig([]string{"123.123.123.123"})
	require.NoError(t, err)

	ts, client := getTestClient(t, tlsConfig)
	ts.StartTLS()

	defer ts.Close()

	_, err = client.Get(ts.URL)
	require.Regexp(t, "x509: certificate is valid for 123\\.123\\.123\\.123, not 127\\.0\\.0\\.1", err)
}

func getTestClient(t *testing.T, tlsConfig *tls.Config) (*httptest.Server, http.Client) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello")
	}))

	ts.TLS = tlsConfig

	certs := x509.NewCertPool()

	for _, c := range tlsConfig.Certificates {
		roots, err := x509.ParseCertificates(c.Certificate[len(c.Certificate)-1])
		require.NoError(t, err)

		for _, root := range roots {
			certs.AddCert(root)
		}
	}

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certs,
			},
		},
	}

	return ts, client
}
