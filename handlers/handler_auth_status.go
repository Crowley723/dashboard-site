package handlers

import (
	"homelab-dashboard/auth"
	"homelab-dashboard/middlewares"
	"net/http"
)

type AuthStatusResponse struct {
	Authenticated bool       `json:"authenticated"`
	User          *auth.User `json:"user,omitempty"`
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
