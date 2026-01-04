package models

import (
	"slices"
	"time"
)

type ServiceAccount struct {
	Sub            string    `json:"sub"`
	Iss            string    `json:"iss"`
	Name           string    `json:"name"`
	LookupId       string    `json:"lookup_id"`
	SecretHash     string    `json:"secret_hash"`
	TokenExpiresAt time.Time `json:"token_expires_at"`
	Scopes         []string  `json:"scopes"`
	IsDisabled     bool      `json:"is_disabled"`
	CreatedBySub   string    `json:"created_by_sub"`
	CreatedByIss   string    `json:"created_by_iss"`
	CreatedAt      time.Time `json:"created_at"`
}

func (s ServiceAccount) GetIss() string {
	return s.Iss
}

func (s ServiceAccount) GetSub() string {
	return s.Sub
}

func (s ServiceAccount) GetScopes() []string {
	return s.Scopes
}

func (s ServiceAccount) HasScope(scope string) bool {
	return slices.Contains(s.GetScopes(), scope)
}

func (s ServiceAccount) MatchesOwner(iss, sub string) bool {
	return s.Iss == iss && s.Sub == sub
}
