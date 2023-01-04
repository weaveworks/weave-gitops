package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type AdminClaims struct {
	jwt.RegisteredClaims
}

type TokenSigner interface {
	Sign(subject string) (string, error)
}

type TokenVerifier interface {
	Verify(token string) (*AdminClaims, error)
}

type TokenSignerVerifier interface {
	TokenSigner
	TokenVerifier
}

type HMACTokenSignerVerifier struct {
	expireAfter time.Duration
	hmacSecret  []byte

	devMode bool
}

func NewHMACTokenSignerVerifier(expireAfter time.Duration) (*HMACTokenSignerVerifier, error) {
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

func (sv *HMACTokenSignerVerifier) Sign(subject string) (string, error) {
	claims := AdminClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(sv.expireAfter).UTC()),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
			Subject:   subject,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(sv.hmacSecret)
}

func (sv *HMACTokenSignerVerifier) Verify(tokenString string) (*AdminClaims, error) {
	if sv.devMode {
		parser := jwt.NewParser()

		token, _, err := parser.ParseUnverified(tokenString, &AdminClaims{})
		if err != nil {
			return nil, fmt.Errorf("failed to parse unverified token: %w", err)
		}

		return token.Claims.(*AdminClaims), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return sv.hmacSecret, nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	if claims, ok := token.Claims.(*AdminClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("invalid token")
	}
}

func (sv *HMACTokenSignerVerifier) SetDevMode(enabled bool) {
	sv.devMode = enabled
}
