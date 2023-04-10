package auth

import (
	"fmt"
)

// ClaimsConfig provides the keys to extract the details for a Principal
// from set of JWT claims.
type ClaimsConfig struct {
	Username string
	Groups   string
}

type claimsToken interface {
	Claims(v interface{}) error
}

// PrincipalFromClaims takes a token and parses the claims using the
// configuration and returns a configured UserPrincipal with the details in the
// claims.
func (c *ClaimsConfig) PrincipalFromClaims(token claimsToken) (*UserPrincipal, error) {
	claims := map[string]interface{}{}
	if err := token.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims from the JWT token: %w", err)
	}

	var (
		idKey     = ScopeEmail
		groupsKey = ScopeGroups
	)

	if c != nil && c.Username != "" {
		idKey = c.Username
	}

	if c != nil && c.Groups != "" {
		groupsKey = c.Groups
	}

	id, ok := claims[idKey].(string)
	if !ok {
		return nil, fmt.Errorf("missing %q claim in response", idKey)
	}

	groups := []string{}

	if v, ok := claims[groupsKey]; ok {
		
		gv, ok := v.([]interface{})

		if ok {
			for _, v := range gv {
				if s, ok := v.(string); !ok {
					return nil, fmt.Errorf("invalid groups claim %q in response %v", groupsKey, v)
				} else {
					groups = append(groups, s)
				}
			}
		} else {
			if s, ok := v.(string); ok && len(s) > 0{
				groups = append(groups, s)
			} else {
				return nil, fmt.Errorf("the groups claim %q is an empty value", groupsKey)
			}
		}
	}

	return &UserPrincipal{ID: id, Groups: groups}, nil
}
