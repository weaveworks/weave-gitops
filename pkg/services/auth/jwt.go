package auth

import (
	"time"

	"github.com/pkg/errors"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"

	"github.com/golang-jwt/jwt/v4"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// ExpirationTime jwt token expiration time
const ExpirationTime = time.Minute * 15

// ErrUnauthorizedToken unauthorized token error
var ErrUnauthorizedToken = errors.New("unauthorized token")

// Claims is a custom JWT claims that contains some token information
type Claims struct {
	jwt.RegisteredClaims
	Provider      gitproviders.GitProviderName `json:"provider"`
	ProviderToken string                       `json:"provider_token"`
}

// JWTClient represents a type that has methods to generate and verify JWT tokens.
//counterfeiter:generate . JWTClient
type JWTClient interface {
	GenerateJWT(expirationTime time.Duration, providerName gitproviders.GitProviderName, providerToken string) (string, error)
	VerifyJWT(accessToken string) (*Claims, error)
}

// NewJwtClient initialize JWTClient instance
func NewJwtClient(secretKey string) JWTClient {
	return &internalJWTClient{secretKey: secretKey}
}

type internalJWTClient struct {
	secretKey string
}

// GenerateJWT generates and signs a new token
func (i *internalJWTClient) GenerateJWT(expirationTime time.Duration, providerName gitproviders.GitProviderName, providerToken string) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationTime)),
		},
		Provider:      providerName,
		ProviderToken: providerToken,
	}

	if expirationTime == 0 {
		// It is possible for the GitLab backend to specify an `expires_in` of 0.
		// Edit the `Expire access tokens` setting to enable/disable expiring tokens.
		// Gitlab defaults to 2 hour expiration, so replicate it here I guess?
		claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(2 * time.Hour))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	return token.SignedString([]byte(i.secretKey))
}

// VerifyJWT verifies the access token string and return a user claim if the token is valid
func (i *internalJWTClient) VerifyJWT(accessToken string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, errors.New("unexpected token signing method")
			}

			return []byte(i.secretKey), nil
		},
	)

	if err != nil {
		return nil, errors.WithMessage(ErrUnauthorizedToken, err.Error())
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
