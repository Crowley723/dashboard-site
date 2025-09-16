package handlers

import (
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/models"
	"net/http"
)

func GETCallbackHandler(ctx *middlewares.AppContext) {
	if errorParam := ctx.Request.URL.Query().Get("error"); errorParam != "" {
		errorDesc := ctx.Request.URL.Query().Get("error_description")

		ctx.Logger.Warn("OIDC callback error", "error", errorParam, "description", errorDesc)
		ctx.Redirect("/callback?error="+errorParam, http.StatusFound)
		return
	}

	user := &models.User{}
	user, err := ctx.OIDCProvider.HandleCallback(ctx)
	if err != nil {
		ctx.Logger.Error("Failed to handle OIDC callback", "error", err)
		ctx.Redirect("/callback?error=auth_failed", http.StatusFound)
		return
	}

	ctx.Logger.Info("User successfully authenticated",
		"user_id", user.Sub,
		"username", user.Username,
		"email", user.Email,
	)

	redirectTo := ctx.SessionManager.GetRedirectAfterLogin(ctx)
	if redirectTo != "" {
		ctx.Redirect(redirectTo, http.StatusFound)
		return
	}

	ctx.Redirect("/auth/complete?status=success", http.StatusFound)
}
