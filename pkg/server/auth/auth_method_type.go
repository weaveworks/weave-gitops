// This is mostly cribbed from https://github.com/uber-go/zap/blob/master/zapcore/level.go
// as that works nicely with cobra

package auth

import (
	"bytes"
	"fmt"
)

type AuthMethod uint8

const (
	// User & password read from a secret
	UserAccount AuthMethod = iota
	// OIDC authentication (recommended)
	OIDC
	// EE CLI tokens
	TokenPassthrough

	// Anonymous
	Anonymous
)

// This is a function to mimic a const slice
func DefaultAuthMethods() []AuthMethod {
	return []AuthMethod{UserAccount, OIDC}
}

// Returns all the auth methods that can be configured via the auth-methods flag
// Anonymous is not included as it is configured via another --insecure-no-auth flag
func AllUserAuthMethods() []string {
	allUserAuthMethods := []AuthMethod{UserAccount, OIDC, TokenPassthrough}
	res := []string{}
	for _, method := range allUserAuthMethods {
		res = append(res, method.String())
	}

	return res

}

func DefaultAuthMethodStrings() []string {
	res := []string{}
	for _, method := range DefaultAuthMethods() {
		res = append(res, method.String())
	}

	return res
}

func ParseAuthMethodArray(authStrings []string) (map[AuthMethod]bool, error) {
	res := map[AuthMethod]bool{}

	for _, methodString := range authStrings {
		method, err := ParseAuthMethod(methodString)
		if err != nil {
			return nil, err
		}

		res[method] = true
	}

	return res, nil
}

func (am *AuthMethod) String() string {
	switch *am {
	case UserAccount:
		return "user-account"
	case OIDC:
		return "oidc"
	case TokenPassthrough:
		return "token-passthrough"
	case Anonymous:
		return "anonymous"
	default:
		return fmt.Sprintf("AuthMethod(%d)", am)
	}
}

func (am *AuthMethod) UnmarshalText(text []byte) error {
	text = bytes.ToLower(text)
	switch string(text) {
	case "user-account":
		*am = UserAccount
	case "oidc":
		*am = OIDC
	case "token-passthrough":
		*am = TokenPassthrough
	case "anonymous":
		*am = Anonymous
	default:
		return fmt.Errorf("unknown auth method '%q'", text)
	}

	return nil
}

func ParseAuthMethod(text string) (AuthMethod, error) {
	var method AuthMethod
	err := method.UnmarshalText([]byte(text))

	return method, err
}
