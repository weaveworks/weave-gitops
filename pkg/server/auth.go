package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

var UserIDContextKey = new(int)

var AuthContextKey = new(int)

const SessionIDCookieName = "token"

const UserIDKey = "userID"

const AuthSecret = "somesecret"

func doError(w http.ResponseWriter, msg string) {
	bytes, err := json.Marshal(struct {
		Message string `json:"msg"`
		Code    int    `json:"code"`
	}{Message: msg, Code: http.StatusUnauthorized})

	log.Debug(msg)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write(bytes)
}

func extractUserID(ctx context.Context) (string, error) {
	s, ok := ctx.Value(UserIDContextKey).(string)

	if !ok {
		return "", errors.New("could not read userID from context")
	}

	return s, nil
}

func withAuth(base http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cookie, err := r.Cookie(SessionIDCookieName)

		if err != nil {
			doError(w, "no cookie")
			return
		}

		sessionInfo, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			return []byte(AuthSecret), nil
		})

		if err != nil {
			doError(w, "could not parse session")
			return
		}

		ctx = context.WithValue(ctx, AuthContextKey, sessionInfo)

		r = r.WithContext(ctx)

		base.ServeHTTP(w, r)
	})
}
