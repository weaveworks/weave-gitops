package server

import (
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-logr/logr"
)

func TestCreateRedirector(t *testing.T) {
	log := logr.Discard()
	// Easiest way to create a filesystem..
	fsys, err := fs.Sub(os.DirFS("testdata"), "public")
	if err != nil {
		t.Fatalf("failed to create fs: %v", err)
	}

	t.Run("We read the index.html and inject base", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/foo", nil)
		w := httptest.NewRecorder()
		handler := IndexHTMLHandler(fsys, log, "/prefix")

		handler.ServeHTTP(w, req)

		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)

		// Check the status code
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status OK; got %v", resp.StatusCode)
		}

		// Check that the base tag was injected
		if !strings.Contains(string(body), `<base href="/prefix/">`) {
			t.Errorf("base tag not injected correctly: %v", string(body))
		}
	})

	t.Run("file not found", func(t *testing.T) {
		brokenFS, err := fs.Sub(os.DirFS("testdata"), "nonexistent")
		if err != nil {
			t.Fatalf("failed to create fs: %v", err)
		}
		req := httptest.NewRequest("GET", "http://example.com/foo", nil)
		w := httptest.NewRecorder()
		handler := IndexHTMLHandler(brokenFS, log, "/prefix")

		handler.ServeHTTP(w, req)

		resp := w.Result()

		// Check the status code
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status InternalServerError; got %v", resp.StatusCode)
		}
	})
}

func TestAssetHandlerFunc(t *testing.T) {
	// Mock assetHandler to just record that it was called and with what request
	assetHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("assetHandler called"))
	})

	// Mock redirector to just record that it was called and with what request
	redirector := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("redirector called"))
	})

	handler := AssetHandler(assetHandler, redirector)

	tests := []struct {
		name       string
		requestURI string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Asset request with extension",
			requestURI: "/static/somefile.js",
			wantStatus: http.StatusOK,
			wantBody:   "assetHandler called",
		},
		{
			name:       "Non-asset request",
			requestURI: "/some-page",
			wantStatus: http.StatusOK,
			wantBody:   "redirector called",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.requestURI, nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, tt.wantStatus)
			}

			if rr.Body.String() != tt.wantBody {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), tt.wantBody)
			}
		})
	}
}
