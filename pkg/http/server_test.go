package http_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"testing"

	. "github.com/onsi/gomega"

	wegohttp "github.com/weaveworks/weave-gitops/pkg/http"
)

func portInUse(port int) bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func TestMultiServerStartReturnsImmediatelyWithClosedContext(t *testing.T) {
	g := NewGomegaWithT(t)
	srv := wegohttp.MultiServer{
		CertFile: "testdata/localhost.crt",
		KeyFile:  "testdata/localhost.key",
		Logger:   log.Default(),
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	g.Expect(srv.Start(ctx, nil)).To(Succeed())
}

func TestMultiServerWithoutTLSConfigFailsToStart(t *testing.T) {
	g := NewGomegaWithT(t)
	srv := wegohttp.MultiServer{}

	err := srv.Start(t.Context(), nil)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(HavePrefix("failed to create TLS listener"))
}

func TestMultiServerServesOverBothProtocols(t *testing.T) {
	g := NewGomegaWithT(t)

	httpPort := rand.N(49151-1024) + 1024  // #nosec G404
	httpsPort := rand.N(49151-1024) + 1024 // #nosec G404

	for httpPort == httpsPort || portInUse(httpPort) || portInUse(httpsPort) {
		httpPort = rand.N(49151-1024) + 1024  // #nosec G404
		httpsPort = rand.N(49151-1024) + 1024 // #nosec G404
	}

	srv := wegohttp.MultiServer{
		HTTPPort:  httpPort,
		HTTPSPort: httpsPort,
		CertFile:  "testdata/localhost.crt",
		KeyFile:   "testdata/localhost.key",
		Logger:    log.Default(),
	}
	ctx, cancel := context.WithCancel(t.Context())

	exitChan := make(chan struct{})
	go func(exitChan chan<- struct{}) {
		hndlr := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(rw, "success")
		})
		g.Expect(srv.Start(ctx, hndlr)).To(Succeed())
		close(exitChan)
	}(exitChan)

	// test HTTP

	var resp *http.Response

	g.Eventually(func() error {
		var err error
		resp, err = http.Get(fmt.Sprintf("http://localhost:%d/", httpPort))
		return err
	}).Should(Succeed())
	g.Expect(resp).NotTo(BeNil(), "response is nil even though no error has been returned")
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	body, err := io.ReadAll(resp.Body)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(body)).To(Equal("success"))

	// test HTTPS

	certBytes, err := os.ReadFile("testdata/localhost.crt")
	g.Expect(err).NotTo(HaveOccurred())

	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(certBytes)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    rootCAs,
			MinVersion: tls.VersionTLS12,
		},
	}
	c := http.Client{
		Transport: tr,
	}
	resp, err = c.Get(fmt.Sprintf("https://localhost:%d/", httpsPort))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	body, err = io.ReadAll(resp.Body)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(body)).To(Equal("success"))

	cancel()
	g.Eventually(exitChan, "3s").Should(BeClosed())

	// ensure both ports are freed up

	_, err = c.Get(fmt.Sprintf("https://localhost:%d/", httpsPort))
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("connection refused"))

	_, err = http.Get(fmt.Sprintf("http://localhost:%d/", httpPort))
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("connection refused"))
}
