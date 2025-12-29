package handlers

import (
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/models"
	"net/http"
)

type ConfigResponse struct {
	MTLS MTLSConfigResponse `json:"mtls,omnitempty"`
}

type MTLSConfigResponse struct {
	AdminGroup string `json:"admin_group"`
	UserGroup  string `json:"user_group"`
}

type AuthStatusResponse struct {
	Authenticated bool            `json:"authenticated"`
	User          *models.User    `json:"user,omitempty"`
	Config        *ConfigResponse `json:"config,omitempty"`
}

func GETAuthStatusHandler(ctx *middlewares.AppContext) {
	response := AuthStatusResponse{
		Authenticated: false,
	}

	if !ctx.SessionManager.IsUserAuthenticated(ctx) {
		ctx.WriteJSON(http.StatusUnauthorized, response)
		return
	}

	config := &ConfigResponse{}

	if ctx.Config.Features != nil && ctx.Config.Features.MTLSManagement.Enabled {
		config.MTLS = MTLSConfigResponse{
			AdminGroup: ctx.Config.Features.MTLSManagement.AdminGroup,
			UserGroup:  ctx.Config.Features.MTLSManagement.UserGroup,
		}
	}

	response.Config = config

	user, ok := ctx.SessionManager.GetAuthenticatedUser(ctx)
	if user == nil {
		ctx.WriteJSON(http.StatusUnauthorized, response)
		return
	}

	if ok {
		response.Authenticated = true
		response.User = user
		ctx.WriteJSON(http.StatusOK, response)
		return
	}

	ctx.WriteJSON(http.StatusUnauthorized, response)
}
