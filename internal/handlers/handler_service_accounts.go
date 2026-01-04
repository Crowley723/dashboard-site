package handlers

import (
	"encoding/json"
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/models"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ServiceAccount struct {
	Sub          string    `json:"sub"`
	Iss          string    `json:"iss"`
	Name         string    `json:"name"`
	Token        string    `json:"token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsDisabled   bool      `json:"is_disabled"`
	Scopes       []string  `json:"scopes,omitempty"`
	CreatedByIss string    `json:"created_by_iss"`
	CreatedBySub string    `json:"created_by_sub"`
	CreatedAt    time.Time `json:"created_at"`
}

func POSTServiceAccount(ctx *middlewares.AppContext) {
	user := ctx.GetPrincipal().(*models.User)

	type request struct {
		Name           string   `json:"name"`
		TokenExpiresAt string   `json:"token_expires_at"`
		Scopes         []string `json:"scopes"`
	}

	var req request
	if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
		ctx.Logger.Error("failed to decode request body", "error", err)
		ctx.SetJSONError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		ctx.SetJSONError(http.StatusBadRequest, "Service account name is required")
		return
	}

	if len(req.Name) > 255 {
		ctx.SetJSONError(http.StatusBadRequest, "Service account name too long (max 255 characters)")
		return
	}

	if len(req.Scopes) == 0 {
		ctx.SetJSONError(http.StatusBadRequest, "At least one scope is required")
		return
	}

	expiryTime, err := time.Parse(time.RFC3339, req.TokenExpiresAt)
	if err != nil {
		ctx.Logger.Error("failed to parse token expiry time", "error", err)
		ctx.SetJSONError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if expiryTime.Before(time.Now()) {
		ctx.SetJSONError(http.StatusBadRequest, "Token expiry time must be in the future")
		return
	}

	//TODO: allow configuring max service token lifetime.
	//maxExpiry := time.Now().Add(365 * 24 * time.Hour)
	//if expiryTime.After(maxExpiry) {
	//	ctx.SetJSONError(http.StatusBadRequest, "Token expiry cannot exceed 1 year")
	//	return
	//}

	userScopes := user.GetScopes()

	if !HasAllScopes(userScopes, req.Scopes) {
		ctx.SetJSONError(http.StatusForbidden, "Cannot grant scopes you don't have")
		return
	}

	token, lookupId, hashedSecret, err := middlewares.GenerateAPIToken()
	if err != nil {
		ctx.Logger.Error("failed to generate API token", "error", err)
		ctx.SetJSONError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	serviceAccount := &models.ServiceAccount{
		Sub:            uuid.New().String(),
		Iss:            ctx.Config.Server.ExternalURL,
		Name:           req.Name,
		LookupId:       lookupId,
		SecretHash:     hashedSecret,
		TokenExpiresAt: expiryTime,
		Scopes:         req.Scopes,
		CreatedByIss:   user.Iss,
		CreatedBySub:   user.Sub,
		CreatedAt:      time.Now(),
	}

	sa, err := ctx.Storage.CreateServiceAccount(ctx, serviceAccount)
	if err != nil {
		ctx.SetJSONError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		ctx.Logger.Error("failed to create service account", "error", err)
		return
	}

	response := ServiceAccount{
		Sub:          sa.Sub,
		Iss:          sa.Iss,
		Name:         sa.Name,
		Token:        token,
		Scopes:       sa.Scopes,
		ExpiresAt:    sa.TokenExpiresAt,
		IsDisabled:   sa.IsDisabled,
		CreatedByIss: sa.CreatedByIss,
		CreatedBySub: sa.CreatedBySub,
		CreatedAt:    sa.CreatedAt,
	}

	ctx.WriteJSON(http.StatusCreated, response)
}

func GETServiceAccounts(ctx *middlewares.AppContext) {

}

func GETServiceAccountWhoami(ctx *middlewares.AppContext) {
	sa := ctx.GetPrincipal().(*models.ServiceAccount)

	response := ServiceAccount{
		Sub:          sa.Sub,
		Iss:          sa.Iss,
		Name:         sa.Name,
		Scopes:       sa.Scopes,
		IsDisabled:   sa.IsDisabled,
		ExpiresAt:    sa.TokenExpiresAt,
		CreatedByIss: sa.CreatedByIss,
		CreatedBySub: sa.CreatedBySub,
		CreatedAt:    sa.CreatedAt,
	}

	ctx.WriteJSON(http.StatusOK, response)
}

func HasAllScopes(userScopes, requestedScopes []string) bool {
	scopeSet := make(map[string]bool, len(userScopes))
	for _, scope := range userScopes {
		scopeSet[scope] = true
	}

	for _, scope := range requestedScopes {
		if !scopeSet[scope] {
			return false
		}
	}

	return true
}
