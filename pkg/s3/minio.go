package s3

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinioClient(endpoint string, accessKey, secretKey, caCert []byte) (*minio.Client, error) {
	tr, err := NewTLSRoundTripper(caCert)
	if err != nil {
		return nil, fmt.Errorf("failed creating transport: %w", err)
	}

	return minio.New(
		endpoint,
		&minio.Options{
			Creds:        credentials.NewStaticV4(string(accessKey), string(secretKey), ""),
			Secure:       true,
			BucketLookup: minio.BucketLookupPath,
			Transport:    tr,
		},
	)
}

func NewTLSRoundTripper(caCert []byte) (http.RoundTripper, error) {
	tr, err := minio.DefaultTransport(true)
	if err != nil {
		return nil, fmt.Errorf("failed creating default transport: %w", err)
	}

	tr.TLSClientConfig = &tls.Config{
		// Can't use SSLv3 because of POODLE and BEAST
		// Can't use TLSv1.0 because of POODLE and BEAST using CBC cipher
		// Can't use TLSv1.1 because of RC4 cipher usage
		MinVersion: tls.VersionTLS12,
	}
	rootCAs := mustGetSystemCertPool()

	rootCAs.AppendCertsFromPEM(caCert)
	tr.TLSClientConfig.RootCAs = rootCAs

	return tr, nil
}

// mustGetSystemCertPool - return system CAs or empty pool in case of error (or windows)
func mustGetSystemCertPool() *x509.CertPool {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return x509.NewCertPool()
	}

	return pool
}
