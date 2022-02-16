package auth

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type TokenSigner interface {
	Sign() (string, error)
}

type TokenVerifier interface {
	Verify(token string) error
}

type TokenSignerVerifier interface {
	TokenSigner
	TokenVerifier
}

type HMACTokenSignerVerifier struct {
	expireAfter time.Duration
	hmacSecret  []byte
}

func NewHMACTokenSignerVerifier(expireAfter time.Duration) (TokenSignerVerifier, error) {
	hmacSecret := make([]byte, 64)

	_, err := rand.Read(hmacSecret)
	if err != nil {
		return nil, fmt.Errorf("could not generate random HMAC secret: %w", err)
	}

	return &HMACTokenSignerVerifier{
		expireAfter: expireAfter,
		hmacSecret:  hmacSecret,
	}, nil
}

func (sv *HMACTokenSignerVerifier) Sign() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, &jwt.StandardClaims{
		IssuedAt:  time.Now().UTC().Unix(),
		ExpiresAt: time.Now().Add(sv.expireAfter).UTC().Unix(),
		NotBefore: time.Now().UTC().Unix(),
		Subject:   "admin",
	})

	return token.SignedString(sv.hmacSecret)
}

func (sv *HMACTokenSignerVerifier) Verify(token string) error {
	return nil
}
