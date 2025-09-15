package handlers

import (
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/models"
	"net/http"
)

func LogoutHandler(ctx *middlewares.AppContext) {
	logger := ctx.Logger

	user := &models.User{}
	user, ok := ctx.SessionManager.GetUser(ctx)
	if !ok {
		logger.Error("Failed to retrieve user session")
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to logout")
		return
	}

	err := ctx.SessionManager.Logout(ctx)
	if err != nil {
		logger.Error("Failed to logout user", "error", err)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to logout")
		return
	}

	if user != nil {
		logger.Info("User logged out", "username", user.Username)
	}

	ctx.SetJSONStatus(http.StatusOK, "OK")
}
