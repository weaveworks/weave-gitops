package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithRoutePrefix(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	routePrefix := "test"
	routePrefixMux := WithRoutePrefix(mux, routePrefix)

	testCases := []struct {
		path     string
		expected int
		content  string
	}{
		// redirect from root to route prefix for nice UX in the browser
		{path: "/", expected: http.StatusFound, content: "<a href=\"/test/\">Found</a>.\n\n"},
		// Redirects to trailing slash, nice for the browser
		{path: "/test", expected: http.StatusMovedPermanently, content: "<a href=\"/test/\">Moved Permanently</a>.\n\n"},
		// We didn't add a root handler
		{path: "/test/", expected: http.StatusNotFound},
		// access the mounted handler
		{path: "/test/foo", expected: http.StatusOK},
		// trailing slash results in 404
		{path: "/test/foo/", expected: http.StatusNotFound},
		// non-existent paths
		{path: "/test/foo/bar", expected: http.StatusNotFound},
		{path: "/test/foo/bar/", expected: http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()
			routePrefixMux.ServeHTTP(w, req)
			if w.Code != tc.expected {
				t.Errorf("expected %d, got %d, %v", tc.expected, w.Code, w.Body.String())
			}
			if tc.content != "" && w.Body.String() != tc.content {
				t.Errorf("expected %s, got %s", tc.content, w.Body.String())
			}
		})
	}
}

func TestInjectHTMLBaseTag(t *testing.T) {
	baseHTML := []byte("<head><title>Test</title></head>")
	// should always have leading and trailing slash
	expectedWithValue := []byte(`<head><base href="/test/"><title>Test</title></head>`)

	testCases := []struct{ routePrefix string }{
		{routePrefix: "test"},
		{routePrefix: "/test"},
		{routePrefix: "test/"},
		{routePrefix: "/test/"},
	}

	for _, tc := range testCases {
		t.Run(tc.routePrefix, func(t *testing.T) {
			actual := InjectHTMLBaseTag(baseHTML, tc.routePrefix)
			if !bytes.Equal(actual, expectedWithValue) {
				t.Errorf("expected %s, got %s", expectedWithValue, actual)
			}
		})
	}
}
