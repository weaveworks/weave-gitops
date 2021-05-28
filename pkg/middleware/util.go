package middleware

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
)

func CreateTestAuthenticatedClient(t *testing.T, u *url.URL, userID string) *http.Client {
	jar, _ := cookiejar.New(&cookiejar.Options{})
	claims := jwt.MapClaims{}
	claims[UserIDKey] = userID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(AuthSecret))

	if err != nil {
		t.Fatal(err)
	}

	cookies := []*http.Cookie{{
		Name:  SessionIDCookieName,
		Value: tokenString,
	}}

	jar.SetCookies(u, cookies)

	return &http.Client{
		Jar: jar,
	}
}
