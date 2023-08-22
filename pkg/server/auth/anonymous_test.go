package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
)

func TestAnonymousPrincipalGetter(t *testing.T) {
	getter := NewAnonymousPrincipalGetter(logr.Discard(), "test-user")

	p, err := getter.Principal(httptest.NewRequest(http.MethodGet, "https://example.com", nil))
	if err != nil {
		t.Fatal(err)
	}

	if p.ID != "test-user" {
		t.Fatalf("got ID %s, want %s", p.ID, "test-user")
	}
}
