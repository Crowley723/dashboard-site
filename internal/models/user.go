package models

import (
	"slices"
	"time"
)

type User struct {
	Sub          string    `json:"sub"`
	Iss          string    `json:"iss"`
	Username     string    `json:"username"`
	DisplayName  string    `json:"display_name"`
	Email        string    `json:"email"`
	Groups       []string  `json:"groups"`
	IsSystem     bool      `json:"is_system"`
	LastLoggedIn time.Time `json:"last_logged_in"`
	CreatedAt    time.Time `json:"created_at"`
}

func (u *User) GetIss() string {
	return u.Iss
}

func (u *User) GetSub() string {
	return u.Sub
}

func (u *User) GetScopes() []string {
	var scopes []string
	for _, group := range u.Groups {
		scopes = append(scopes, mapGroupToScopes(group)...)
	}
	return dedupe(scopes)
}

func (u *User) HasScope(scope string) bool {
	scopes := u.GetScopes()

	return slices.Contains(scopes, scope)
}

func (u *User) MatchesOwner(iss, sub string) bool {
	return u.Iss == iss && u.Sub == sub
}

// TODO: refactor to use config-based mappings. Will require passing appContext to this method.
// mapGroupToScopes is a temporary method to map hardcoded groups to specific scopes
func mapGroupToScopes(group string) []string {
	switch group {
	case "conduit:mtls:admin":
		return []string{
			"cert:request",
			"cert:read",
			"cert:approve",
			"cert:renew",
			"cert:revoke",
		}
	case "conduit:mtls:user":
		return []string{
			"cert:request",
			"cert:read",
			"cert:renew",
		}
	default:
		return []string{}
	}
}

func dedupe(slice []string) []string {
	keys := make(map[string]bool)
	var list []string

	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
