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
	Sub          string     `json:"sub"`
	Iss          string     `json:"iss"`
	Name         string     `json:"name"`
	Token        string     `json:"token,omitempty"`
	ExpiresAt    time.Time  `json:"expires_at"`
	IsDisabled   bool       `json:"is_disabled"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	Scopes       []string   `json:"scopes,omitempty"`
	CreatedByIss string     `json:"created_by_iss"`
	CreatedBySub string     `json:"created_by_sub"`
	CreatedAt    time.Time  `json:"created_at"`
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

	userScopes := user.GetScopes(ctx.Config)

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
		DeletedAt:    sa.DeletedAt,
		CreatedByIss: sa.CreatedByIss,
		CreatedBySub: sa.CreatedBySub,
		CreatedAt:    sa.CreatedAt,
	}

	ctx.WriteJSON(http.StatusCreated, response)
}

func GETServiceAccounts(ctx *middlewares.AppContext) {
	user := ctx.GetPrincipal().(*models.User)

	serviceAccounts, err := ctx.Storage.GetServiceAccountsByCreator(ctx, user.Iss, user.Sub)
	if err != nil {
		ctx.Logger.Error("failed to get service accounts", "error", err)
		ctx.SetJSONError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	response := make([]ServiceAccount, 0, len(serviceAccounts))
	for _, sa := range serviceAccounts {
		response = append(response, ServiceAccount{
			Sub:          sa.Sub,
			Iss:          sa.Iss,
			Name:         sa.Name,
			Scopes:       sa.Scopes,
			ExpiresAt:    sa.TokenExpiresAt,
			IsDisabled:   sa.IsDisabled,
			DeletedAt:    sa.DeletedAt,
			CreatedByIss: sa.CreatedByIss,
			CreatedBySub: sa.CreatedBySub,
			CreatedAt:    sa.CreatedAt,
		})
	}

	ctx.WriteJSON(http.StatusOK, response)
}

func GETServiceAccountWhoami(ctx *middlewares.AppContext) {
	sa := ctx.GetPrincipal().(*models.ServiceAccount)

	response := ServiceAccount{
		Sub:          sa.Sub,
		Iss:          sa.Iss,
		Name:         sa.Name,
		Scopes:       sa.Scopes,
		IsDisabled:   sa.IsDisabled,
		DeletedAt:    sa.DeletedAt,
		ExpiresAt:    sa.TokenExpiresAt,
		CreatedByIss: sa.CreatedByIss,
		CreatedBySub: sa.CreatedBySub,
		CreatedAt:    sa.CreatedAt,
	}

	ctx.WriteJSON(http.StatusOK, response)
}

func DELETEServiceAccount(ctx *middlewares.AppContext) {
	user := ctx.GetPrincipal().(*models.User)

	targetIss := ctx.Request.URL.Query().Get("iss")
	targetSub := ctx.Request.URL.Query().Get("sub")

	if targetIss == "" || targetSub == "" {
		ctx.SetJSONError(http.StatusBadRequest, "Missing iss or sub query parameters")
		return
	}

	// Get the service account to verify ownership
	sa, err := ctx.Storage.GetServiceAccountByID(ctx, targetIss, targetSub)
	if err != nil {
		ctx.Logger.Error("failed to get service account", "error", err)
		ctx.SetJSONError(http.StatusNotFound, "Service account not found")
		return
	}

	// Verify the user owns this service account
	if sa.CreatedByIss != user.Iss || sa.CreatedBySub != user.Sub {
		ctx.SetJSONError(http.StatusForbidden, "You can only delete service accounts you created")
		return
	}

	// Check if already deleted
	if sa.DeletedAt != nil {
		ctx.SetJSONError(http.StatusGone, "Service account already deleted")
		return
	}

	// Delete (soft delete with audit trail)
	if err := ctx.Storage.DeleteServiceAccount(ctx, targetIss, targetSub); err != nil {
		ctx.Logger.Error("failed to delete service account", "error", err)
		ctx.SetJSONError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	ctx.WriteJSON(http.StatusOK, map[string]string{"message": "Service account deleted"})
}

func PATCHServiceAccountPause(ctx *middlewares.AppContext) {
	user := ctx.GetPrincipal().(*models.User)

	targetIss := ctx.Request.URL.Query().Get("iss")
	targetSub := ctx.Request.URL.Query().Get("sub")

	if targetIss == "" || targetSub == "" {
		ctx.SetJSONError(http.StatusBadRequest, "Missing iss or sub query parameters")
		return
	}

	// Get the service account to verify ownership
	sa, err := ctx.Storage.GetServiceAccountByID(ctx, targetIss, targetSub)
	if err != nil {
		ctx.Logger.Error("failed to get service account", "error", err)
		ctx.SetJSONError(http.StatusNotFound, "Service account not found")
		return
	}

	// Verify the user owns this service account
	if sa.CreatedByIss != user.Iss || sa.CreatedBySub != user.Sub {
		ctx.SetJSONError(http.StatusForbidden, "You can only pause service accounts you created")
		return
	}

	// Check if deleted
	if sa.DeletedAt != nil {
		ctx.SetJSONError(http.StatusGone, "Cannot pause a deleted service account")
		return
	}

	// Pause
	if err := ctx.Storage.PauseServiceAccount(ctx, targetIss, targetSub); err != nil {
		ctx.Logger.Error("failed to pause service account", "error", err)
		ctx.SetJSONError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	ctx.WriteJSON(http.StatusOK, map[string]string{"message": "Service account paused"})
}

func PATCHServiceAccountUnpause(ctx *middlewares.AppContext) {
	user := ctx.GetPrincipal().(*models.User)

	targetIss := ctx.Request.URL.Query().Get("iss")
	targetSub := ctx.Request.URL.Query().Get("sub")

	if targetIss == "" || targetSub == "" {
		ctx.SetJSONError(http.StatusBadRequest, "Missing iss or sub query parameters")
		return
	}

	// Get the service account to verify ownership
	sa, err := ctx.Storage.GetServiceAccountByID(ctx, targetIss, targetSub)
	if err != nil {
		ctx.Logger.Error("failed to get service account", "error", err)
		ctx.SetJSONError(http.StatusNotFound, "Service account not found")
		return
	}

	// Verify the user owns this service account
	if sa.CreatedByIss != user.Iss || sa.CreatedBySub != user.Sub {
		ctx.SetJSONError(http.StatusForbidden, "You can only unpause service accounts you created")
		return
	}

	// Check if deleted
	if sa.DeletedAt != nil {
		ctx.SetJSONError(http.StatusGone, "Cannot unpause a deleted service account")
		return
	}

	// Unpause
	if err := ctx.Storage.UnpauseServiceAccount(ctx, targetIss, targetSub); err != nil {
		ctx.Logger.Error("failed to unpause service account", "error", err)
		ctx.SetJSONError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	ctx.WriteJSON(http.StatusOK, map[string]string{"message": "Service account unpaused"})
}

func GETUserScopes(ctx *middlewares.AppContext) {
	user := ctx.GetPrincipal().(*models.User)

	scopes := user.GetScopes(ctx.Config)
	if scopes == nil {
		scopes = []string{}
	}

	ctx.WriteJSON(http.StatusOK, map[string][]string{"scopes": scopes})
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
