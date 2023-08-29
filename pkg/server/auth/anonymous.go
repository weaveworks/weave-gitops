package auth

import (
	"net/http"

	"github.com/go-logr/logr"
)

// AnonymousPrincipalGetter will always succeed.
//
// The principal it returns will have the configured name.
type AnonymousPrincipalGetter struct {
	log  logr.Logger
	name string
}

func NewAnonymousPrincipalGetter(log logr.Logger, name string) PrincipalGetter {
	return &AnonymousPrincipalGetter{
		log:  log,
		name: name,
	}
}

func (pg *AnonymousPrincipalGetter) Principal(r *http.Request) (*UserPrincipal, error) {
	return &UserPrincipal{ID: pg.name, Groups: []string{}}, nil
}
