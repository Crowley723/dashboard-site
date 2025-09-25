package handlers

import (
	"fmt"
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/models"
	"net/http"
	"net/url"
)

func GETCallbackHandler(ctx *middlewares.AppContext) {
	if errorParam := ctx.Request.URL.Query().Get("error"); errorParam != "" {
		errorDesc := ctx.Request.URL.Query().Get("error_description")
		errorURI := ctx.Request.URL.Query().Get("error_uri")
		state := ctx.Request.URL.Query().Get("state")

		ctx.Logger.Warn("OIDC callback error", "error", errorParam, "description", errorDesc)

		errorURL := "/error?error=" + url.QueryEscape(errorParam)
		if errorDesc != "" {
			errorURL += "&error_description=" + url.QueryEscape(errorDesc)
		}
		if errorURI != "" {
			errorURL += "&error_uri=" + url.QueryEscape(errorURI)
		}
		if state != "" {
			errorURL += "&state=" + url.QueryEscape(state)
		}

		ctx.Redirect(errorURL, http.StatusFound)
		return
	}

	user := &models.User{}
	user, err := ctx.OIDCProvider.HandleCallback(ctx)
	if err != nil {
		ctx.Logger.Error("Failed to handle OIDC callback", "error", err)
		errorURL := fmt.Sprintf("/error?error=%s&error_description=%s", url.QueryEscape("server error"), url.QueryEscape("authentication failed"))
		ctx.Redirect(errorURL, http.StatusFound)
		return
	}

	ctx.Logger.Info("User successfully authenticated",
		"user_id", user.Sub,
		"username", user.Username,
		"email", RedactEmail(user.Email),
	)

	redirectTo := ctx.SessionManager.GetRedirectAfterLogin(ctx)
	if redirectTo != "" {
		ctx.Redirect(redirectTo, http.StatusFound)
		return
	}

	ctx.Redirect("/", http.StatusFound)
}
