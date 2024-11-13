package internal

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/utils"
)

// CodeVerifier is for PKCE OAuth workflows. This will generate a random string with a random length.
type CodeVerifier struct {
	value string
}

func NewCodeVerifier(minLen, maxLen int) (CodeVerifier, error) {
	value, err := utils.GenerateRandomString(minLen, maxLen)
	if err != nil {
		return CodeVerifier{}, fmt.Errorf("new code verifier: %w", err)
	}

	return CodeVerifier{
		value: value,
	}, nil
}

// CodeChallenge is an encoded hash value to be used in the authorization flow
func (c CodeVerifier) CodeChallenge() (string, error) {
	h := sha256.New()

	_, err := h.Write([]byte(c.value))
	if err != nil {
		return "", fmt.Errorf("code verifier issue hashing challenge: %w", err)
	}

	encodedValue := base64.StdEncoding.EncodeToString(h.Sum(nil))
	encodedValue = strings.ReplaceAll(encodedValue, "+", "-")
	encodedValue = strings.ReplaceAll(encodedValue, "/", "_")
	encodedValue = strings.ReplaceAll(encodedValue, "=", "")

	return encodedValue, nil
}

// RawValue is the original string to be used to fetch a refresh token during the OAuth callback flow
func (c CodeVerifier) RawValue() string {
	return c.value
}
