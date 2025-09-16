package handlers

import (
	"homelab-dashboard/internal/middlewares"
	"net/http"
)

func GETLoginHandler(ctx *middlewares.AppContext) {
	if ctx.SessionManager.IsAuthenticated(ctx) {
		ctx.Logger.Info("User already authenticated")
		ctx.SetJSONStatus(http.StatusOK, "ok")
		return
	}

	currentURL := ctx.Request.Header.Get("Referer")
	if currentURL == "" {
		currentURL = "/"
	}

	ctx.SessionManager.SetRedirectAfterLogin(ctx, currentURL)

	authURL, err := ctx.OIDCProvider.StartLogin(ctx)
	if err != nil {
		ctx.Logger.Error("Failed to start login", "error", err)
		ctx.SetJSONError(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	ctx.Logger.Info("Redirecting to OIDC Provider", "url", authURL)

	ctx.WriteJSON(http.StatusOK, map[string]string{
		"status":       "redirect_required",
		"redirect_url": authURL,
	})
}
