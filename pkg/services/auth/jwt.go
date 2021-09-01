package auth

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"

	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/golang-jwt/jwt/v4"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

var SecretKey string

const ExpirationTime = time.Minute * 15

var ErrUnauthorizedToken = errors.New("unauthorized token")

// Claims is a custom JWT claims that contains some token information
type Claims struct {
	jwt.StandardClaims
	Provider      gitproviders.GitProviderName `json:"provider"`
	ProviderToken string                       `json:"provider_token"`
}

//counterfeiter:generate . JWTClient
type JWTClient interface {
	GenerateJWT(secretKey string, expirationTime time.Duration, providerName gitproviders.GitProviderName, providerToken string) (string, error)
	VerifyJWT(secretKey string, accessToken string) (*Claims, error)
}

func NewJwtClient() JWTClient {
	return &internalJwtClient{}
}

type internalJwtClient struct {
}

// GenerateJWT generates and signs a new token
func GenerateJWT(secretKey string, expirationTime time.Duration, providerName gitproviders.GitProviderName, providerToken string) (string, error) {
	claims := Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expirationTime).Unix(),
		},
		Provider:      providerName,
		ProviderToken: providerToken,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(secretKey))
}

// VerifyJWT verifies the access token string and return a user claim if the token is valid
func (i *internalJwtClient) VerifyJWT(secretKey string, accessToken string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}

			return []byte(secretKey), nil
		},
	)
	if err != nil {
		return nil, ErrUnauthorizedToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
	SecretKey = rand.String(20)
}
