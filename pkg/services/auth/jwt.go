package auth

import (
	"fmt"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/dgrijalva/jwt-go"
)

var SecretKey string

const ExpirationTime = time.Minute * 15

// Claims is a custom JWT claims that contains some token information
type Claims struct {
	jwt.StandardClaims
	Provider      gitproviders.GitProviderName `json:"provider"`
	ProviderToken string `json:"provider_token"`
}

// Generate generates and signs a new token
func Generate(secretKey string, expirationTime time.Duration, providerName gitproviders.GitProviderName, providerToken string) (string, error) {
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

// Verify verifies the access token string and return a user claim if the token is valid
func Verify(secretKey string, accessToken string) (*Claims, error) {
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
		return nil, fmt.Errorf("invalid token: %w", err)
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
