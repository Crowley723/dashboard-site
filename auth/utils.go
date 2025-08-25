package auth

import (
	"github.com/coreos/go-oidc/v3/oidc"
)

// Extract user information from ID token claims
func extractUserClaimsFromToken(idToken *oidc.IDToken) (*User, error) {
	var claims struct {
		Sub               string   `json:"sub"`
		Iss               string   `json:"iss"`
		PreferredUsername string   `json:"preferred_username"`
		Name              string   `json:"name"`
		Email             string   `json:"email"`
		Groups            []string `json:"groups"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	user := &User{
		Sub:         claims.Sub,
		Iss:         claims.Iss,
		Username:    claims.PreferredUsername,
		DisplayName: claims.Name,
		Email:       claims.Email,
		Groups:      claims.Groups,
	}

	return user, nil
}
