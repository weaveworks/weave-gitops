package auth_test

import (
	"strings"
	"testing"

	"github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func TestInvariant(t *testing.T) {
	authMethods := []auth.AuthMethod{auth.UserAccount, auth.OIDC, auth.TokenPassthrough}

	for _, method := range authMethods {
		authstring := method.String()

		parsedMethod, err := auth.ParseAuthMethod(authstring)

		if err != nil {
			t.Fatalf("Auth methods should parse without error, got %s", err)
		}

		if parsedMethod != method {
			t.Fatalf("Parsing a stringified method should get the original method, expected %d, got %d", method, parsedMethod)
		}
	}
}

func TestBadAuthMethod(t *testing.T) {
	method, err := auth.ParseAuthMethod("badMethod")
	if !strings.HasPrefix(err.Error(), "Unknown auth method") {
		t.Fatalf("Expected ParseAuthMethod to produce 'Unknown auth method' error, instead got (method='%d', err='%s')", method, err)
	}
}

func TestParseAuthMethodArray(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	parseTests := []struct {
		name        string
		methodArray []string
		expectedMap map[auth.AuthMethod]bool
		expectedErr bool
	}{
		{
			name:        "Empty array",
			methodArray: []string{},
			expectedMap: map[auth.AuthMethod]bool{},
			expectedErr: false,
		},
		{
			name:        "Array of all",
			methodArray: []string{"oidc", "user-account", "token-passthrough"},
			expectedMap: map[auth.AuthMethod]bool{auth.OIDC: true, auth.UserAccount: true, auth.TokenPassthrough: true},
			expectedErr: false,
		},
		{
			name:        "Bad element",
			methodArray: []string{"unknown-method"},
			expectedMap: nil,
			expectedErr: true,
		},
	}

	for _, tt := range parseTests {
		t.Run(tt.name, func(t *testing.T) {
			methods, err := auth.ParseAuthMethodArray(tt.methodArray)

			if tt.expectedErr {
				g.Expect(err).To(gomega.HaveOccurred())
				g.Expect(methods).To(gomega.BeNil())
			} else {
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(methods).To(gomega.Equal(tt.expectedMap))
			}
		})
	}
}
