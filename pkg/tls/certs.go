package tls

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

type Certificate struct {
	Cert []byte
	Key  []byte
}

func GenerateSelfSignedCertificate(sans ...string) (Certificate, error) {
	notBefore := time.Now().Add(-1 * time.Minute)
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(42),
		Subject: pkix.Name{
			CommonName: "Weave GitOps CLI",
		},
		NotBefore:             notBefore,
		NotAfter:              notBefore.Add(3 * 24 * time.Hour),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		DNSNames:              sans,
	}

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return Certificate{}, fmt.Errorf("failed to create private key: %w", err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privKey.PublicKey, privKey)
	if err != nil {
		return Certificate{}, fmt.Errorf("failed to create certificate: %w", err)
	}

	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		return Certificate{}, fmt.Errorf("failed to encode certificate: %w", err)
	}

	keyBytes, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return Certificate{}, fmt.Errorf("unable to marshal private key: %w", err)
	}

	privKeyPEM := new(bytes.Buffer)
	if err := pem.Encode(privKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}); err != nil {
		return Certificate{}, fmt.Errorf("failed to encode private key: %w", err)
	}

	return Certificate{
		Cert: certPEM.Bytes(),
		Key:  privKeyPEM.Bytes(),
	}, nil
}
