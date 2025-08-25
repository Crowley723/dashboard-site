package handlers

import (
	"homelab-dashboard/auth"
	"homelab-dashboard/middlewares"
	"net/http"
)

func LoginHandler(ctx *middlewares.AppContext) {
	if auth.IsAuthenticated(ctx) {
		ctx.Logger.Info("User already authenticated, redirecting to home")
		ctx.SetJSONStatus(http.StatusOK, "ok")
		return
	}

	currentURL := ctx.Request.Header.Get("Referer")
	if currentURL == "" {
		currentURL = "/"
	}

	ctx.SessionManager.Put(ctx.Request.Context(), "redirect_after_login", currentURL)

	authURL, err := auth.StartLogin(ctx)
	if err != nil {
		ctx.Logger.Error("Failed to start login", "error", err)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to initiate login")
		return
	}

	ctx.Logger.Info("Redirecting to OIDC Provider", "url", authURL)

	ctx.WriteJSON(http.StatusOK, map[string]string{
		"status":       "redirect_required",
		"redirect_url": authURL,
	})
}
