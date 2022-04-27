package internal

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/weaveworks/weave-gitops/gitops/pkg/utils"
	"strings"
)

// CodeVerifier is for PKCE OAuth workflows. This will generate a random string with a random length.
type CodeVerifier struct {
	min   int
	max   int
	value string
}

func NewCodeVerifier(min, max int) (CodeVerifier, error) {
	value, err := utils.GenerateRandomString(min, max)
	if err != nil {
		return CodeVerifier{}, fmt.Errorf("new code verifier: %w", err)
	}

	return CodeVerifier{
		min:   min,
		max:   max,
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
	encodedValue = strings.Replace(encodedValue, "+", "-", -1)
	encodedValue = strings.Replace(encodedValue, "/", "_", -1)
	encodedValue = strings.Replace(encodedValue, "=", "", -1)

	return encodedValue, nil
}

// RawValue is the original string to be used to fetch a refresh token during the OAuth callback flow
func (c CodeVerifier) RawValue() string {
	return c.value
}
