package fluxinstall

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/fluxinstall/internal/httpclient"
	isrc "github.com/weaveworks/weave-gitops/pkg/fluxinstall/internal/src"
)

type HTTPGetter interface {
	Get(url string) (resp *http.Response, err error)
}

type Product struct {
	Version string
	cli     HTTPGetter
}

func NewProduct(version string) *Product {
	return &Product{
		Version: version,
		cli:     httpclient.NewHTTPClient(),
	}
}

func NewProductWithHTTPClient(version string, cli HTTPGetter) *Product {
	return &Product{
		Version: version,
		cli:     cli,
	}
}

func (p *Product) IsSourceImpl() isrc.InstallSrcSigil {
	return isrc.InstallSrcSigil{}
}

func (p *Product) Install(ctx context.Context) (string, error) {
	gitopsCacheFluxDir, err := getFluxBinaryDir(p.Version)
	if err != nil {
		return "", err
	}

	// check if the dir not found
	if _, err := os.Stat(gitopsCacheFluxDir); os.IsNotExist(err) {
		if err := os.MkdirAll(gitopsCacheFluxDir, 0755); err != nil {
			return "", err
		}
	}

	// TODO enable windows support
	extension := "tar.gz"
	binaryURL := fmt.Sprintf("https://github.com/fluxcd/flux2/releases/download/v%s/flux_%s_%s_%s.%s", p.Version, p.Version, runtime.GOOS, runtime.GOARCH, extension)
	client := p.cli
	resp, err := client.Get(binaryURL)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(body)
	sha256sum := fmt.Sprintf("%x", h.Sum(nil))
	filename := fmt.Sprintf("flux_%s_%s_%s.%s", p.Version, runtime.GOOS, runtime.GOARCH, extension)

	if err := p.verifyChecksum(filename, sha256sum); err != nil {
		return "", err
	}

	binary, err := extractTarGz(body)
	if err != nil {
		return "", err
	}

	binaryPath := filepath.Join(gitopsCacheFluxDir, "flux")
	if err := os.WriteFile(binaryPath, binary, 0744); err != nil {
		return "", err
	}

	return binaryPath, nil
}

func (p *Product) Remove(ctx context.Context) error {
	dir, err := getFluxBinaryDir(p.Version)
	if err != nil {
		return err
	}

	return os.RemoveAll(dir)
}

func (p *Product) Find(ctx context.Context) (string, error) {
	dir, err := getFluxBinaryDir(p.Version)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "flux"), nil
}

func (p *Product) verifyChecksum(filename string, sum string) error {
	checkSumURL := fmt.Sprintf("https://github.com/fluxcd/flux2/releases/download/v%s/flux_%s_checksums.txt", p.Version, p.Version)
	client := p.cli
	resp, err := client.Get(checkSumURL)

	if err != nil {
		return err
	}

	// read string from response body
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	checksum := map[string]string{}
	lines := strings.Split(string(body), "\n")

	for _, line := range lines {
		parts := strings.SplitN(line, "  ", 2)

		if len(parts) != 2 {
			continue
		}

		checksum[parts[1]] = parts[0]
	}

	if checksum[filename] != sum {
		return fmt.Errorf("checksum mismatch for %s", filename)
	}

	return nil
}

// extractTarGz accepts an io.Reader of Flux's .tar.gz and extract it to a byte array
func extractTarGz(b []byte) ([]byte, error) {
	// ungzip
	gr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	// untar the "flux" entry
	tr := tar.NewReader(gr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		if hdr.Name != "flux" {
			continue
		}

		var bytesBuffer bytes.Buffer
		if _, err := io.Copy(&bytesBuffer, tr); err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}

		return bytesBuffer.Bytes(), nil
	}

	return nil, fmt.Errorf("no flux binary found in archive")
}

func getFluxBinaryDir(version string) (string, error) {
	// get cache dir
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	gitopsCacheFluxDir := filepath.Join(cacheDir, ".gitops", "flux", version)

	return gitopsCacheFluxDir, nil
}
