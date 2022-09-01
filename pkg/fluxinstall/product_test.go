package fluxinstall

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"net/http"
)

func createMockFluxArchive(data []byte) ([]byte, error) {
	buf := &bytes.Buffer{}

	// Create new Writers for gzip and tar
	// These writers are chained. Writing to the tar writer will
	// write to the gzip writer which in turn will write to
	// the "buf" writer
	gw := gzip.NewWriter(buf)
	tw := tar.NewWriter(gw)

	if err := addToArchive(tw, "flux", data); err != nil {
		return nil, err
	}

	if err := tw.Flush(); err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	if err := gw.Flush(); err != nil {
		return nil, err
	}

	if err := gw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func addToArchive(tw *tar.Writer, filename string, data []byte) error {
	// Create a tar Header from the FileInfo data
	header := &tar.Header{
		Typeflag:   tar.TypeReg,
		Name:       filename,
		Linkname:   filename,
		Size:       int64(len(data)),
		Mode:       644,
		Devmajor:   0,
		Devminor:   0,
		PAXRecords: nil,
		Format:     0,
	}

	// Use full path as name (FileInfoHeader only takes the basename)
	// If we don't do this the directory strucuture would
	// not be preserved
	// https://golang.org/src/archive/tar/common.go?#L626
	header.Name = filename

	// Write file header to the tar archive
	err := tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to tar archive
	_, err = io.Copy(tw, bytes.NewReader(data))
	if err != nil {
		return err
	}

	return nil
}

type MockProductHttpClient struct {
}

func (m *MockProductHttpClient) Get(url string) (resp *http.Response, err error) {
	if url == "https://github.com/fluxcd/flux2/releases/download/v0.32.0/flux_0.32.0_linux_amd64.tar.gz" {
		body, err := createMockFluxArchive([]byte("flux"))

		if err != nil {
			return nil, err
		}

		return &http.Response{
			Body: ioutil.NopCloser(bytes.NewReader(body)),
		}, nil
	} else if url == "https://github.com/fluxcd/flux2/releases/download/v0.32.0/flux_0.32.0_checksums.txt" {
		body := []byte(`77622fd02dd5ad9377e17ecb59fa4f9598016bf0bf9761d09c9ed633840d7c7d  flux_0.32.0_linux_amd64.tar.gz
`)

		return &http.Response{
			Body: ioutil.NopCloser(bytes.NewReader(body)),
		}, nil
	}

	return nil, nil
}

var _ = Describe("Product", func() {

	var product *Product

	It("should verify checksum of flux product", func() {
		By("creating product, and verify checksum", func() {
			product = &Product{
				Version: "0.32.0",
				cli:     &MockProductHttpClient{},
			}
			_, err := product.Install(context.Background())
			Expect(err).To(BeNil())

			err = product.verifyChecksum("flux_0.32.0_linux_amd64.tar.gz", "77622fd02dd5ad9377e17ecb59fa4f9598016bf0bf9761d09c9ed633840d7c7d")
			Expect(err).To(BeNil())
		})
	})

})
