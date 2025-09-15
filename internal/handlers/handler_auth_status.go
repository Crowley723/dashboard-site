package handlers

import (
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/models"
	"net/http"
)

type AuthStatusResponse struct {
	Authenticated bool         `json:"authenticated"`
	User          *models.User `json:"user,omitempty"`
}

func AuthStatusHandler(ctx *middlewares.AppContext) {
	response := AuthStatusResponse{
		Authenticated: false,
	}

	if !ctx.SessionManager.IsUserAuthenticated(ctx) {
		ctx.WriteJSON(http.StatusUnauthorized, response)
		return
	}

	if user, ok := ctx.SessionManager.GetCurrentUser(ctx); ok {
		response.Authenticated = true
		response.User = user
		ctx.WriteJSON(http.StatusOK, response)
		return
	}

	ctx.WriteJSON(http.StatusUnauthorized, response)
}
