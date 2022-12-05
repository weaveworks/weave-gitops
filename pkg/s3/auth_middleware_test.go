package s3

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/minio/minio-go/v7/pkg/signer"
)

func generateRandomBody(method string) io.Reader {
	if method == "GET" || method == "DELETE" {
		return nil
	}

	rand.Seed(time.Now().UnixNano())
	size := rand.Intn(2000) + 2000
	buf := make([]byte, size)
	rand.Read(buf)

	return bytes.NewReader(buf)
}

func TestVerifySignature(t *testing.T) {
	var testcases []struct {
		name     string
		req      *http.Request
		expected error
	}

	for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
		for _, region := range []string{"us-east-1", "us-west-1"} {
			for _, host := range []string{"", "localhost", "localhost:8080", "localhost:9000"} {
				for _, bucket := range []string{"", "testbucket"} {
					for _, object := range []string{"", "testobject"} {
						for _, query := range []string{"", "query", "a=b&c=d"} {
							// good case
							testcases = append(testcases, struct {
								name     string
								req      *http.Request
								expected error
							}{
								name: fmt.Sprintf("good - %s %s/%s/%s/%s?%s", method, region, host, bucket, object, query),
								req: func() *http.Request {
									req, err := http.NewRequest(method, fmt.Sprintf("http://%s/%s/%s?%s", host, bucket, object, query), generateRandomBody(method))
									if err != nil {
										t.Fatal(err)
									}
									signed := signer.SignV4(*req, "gitopsrun", "gitopsrun123", "", region)
									return signed
								}(),
								expected: nil,
							})

							// bad signature cases
							testcases = append(testcases, struct {
								name     string
								req      *http.Request
								expected error
							}{
								name: fmt.Sprintf("bad signature - %s %s/%s/%s/%s?%s", method, region, host, bucket, object, query),
								req: func() *http.Request {
									req, err := http.NewRequest(method, fmt.Sprintf("http://%s/%s/%s?%s", host, bucket, object, query), generateRandomBody(method))
									if err != nil {
										t.Fatal(err)
									}
									signed := signer.SignV4(*req, "gitopsrun", "invalid", "", region)
									return signed
								}(),
								expected: fmt.Errorf("access denied: signature does not match"),
							})

							// bad credential cases
							testcases = append(testcases, struct {
								name     string
								req      *http.Request
								expected error
							}{
								name: fmt.Sprintf("bad credential - %s %s/%s/%s/%s?%s", method, region, host, bucket, object, query),
								req: func() *http.Request {
									req, err := http.NewRequest(method, fmt.Sprintf("http://%s/%s/%s?%s", host, bucket, object, query), generateRandomBody(method))
									if err != nil {
										t.Fatal(err)
									}
									signed := signer.SignV4(*req, "invalid", "gitopsrun123", "", region)
									return signed
								}(),
								expected: fmt.Errorf("access denied: credential does not match"),
							})
						}
					}
				}
			}
		}
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			if tc.expected == nil {
				g.Expect(verifySignature(*tc.req, "gitopsrun", "gitopsrun123")).To(Succeed())
			} else {
				g.Expect(verifySignature(*tc.req, "gitopsrun", "gitopsrun123").Error()).To(Equal(tc.expected.Error()))
			}
		})
	}
}
