package tls_test

import (
	"crypto/tls"
	"crypto/x509"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	wegotls "github.com/weaveworks/weave-gitops/pkg/tls"
)

func TestSelfSignedCertificate(t *testing.T) {
	g := NewGomegaWithT(t)

	cert, err := wegotls.GenerateSelfSignedCertificate("foo", "bar")
	g.Expect(err).NotTo(HaveOccurred(), "error generating certificate")

	parsedCert, err := tls.X509KeyPair(cert.Cert, cert.Key)
	g.Expect(err).NotTo(HaveOccurred(), "error loading key pair")

	g.Expect(parsedCert.Certificate).To(HaveLen(1), "there should only be one certificate")

	x509Cert, err := x509.ParseCertificate(parsedCert.Certificate[0])
	g.Expect(err).NotTo(HaveOccurred(), "error parsing certificate")

	g.Expect(x509Cert.DNSNames).To(HaveLen(2), "unexpected number of SANs found in certificate")
	g.Expect(x509Cert.DNSNames).To(ConsistOf("foo", "bar"), "unexpected SANs found in certificate")
	g.Expect(x509Cert.NotAfter.Sub(x509Cert.NotBefore)).To(Equal(3*24*time.Hour), "unexpected lifetime of certificate")
}
