package handlers

import (
	"encoding/json"
	"fmt"
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/models"
	"net/http"
)

func POSTCertificateRequest(ctx *middlewares.AppContext) {
	user, ok := ctx.SessionManager.GetAuthenticatedUser(ctx)
	if !ok {
		ctx.Logger.Warn("session not found")
		ctx.SetJSONError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	var req struct {
		Message      string `json:"message"`
		ValidityDays int    `json:"validity_days"`
	}

	if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
		ctx.Logger.Error("failed to decode request body", "error", err)
		ctx.SetJSONError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	if req.ValidityDays == 0 {
		req.ValidityDays = 90
	}

	if req.ValidityDays < ctx.Config.Features.MTLSManagement.MinCertificateValidityDays || req.ValidityDays > ctx.Config.Features.MTLSManagement.MaxCertificateValidityDays {
		ctx.SetJSONError(http.StatusBadRequest, fmt.Sprintf("validity_days must be between %d and %d days", ctx.Config.Features.MTLSManagement.MinCertificateValidityDays, ctx.Config.Features.MTLSManagement.MaxCertificateValidityDays))
		return
	}

	commonName := deriveCommonName(user)

	organizationalUnits := deriveOrganizationalUnits(user)

	certRequest, err := ctx.Storage.Certificates().CreateRequest(
		ctx,
		user.Sub,
		user.Iss,
		commonName,
		string(models.StatusAwaitingReview),
		req.Message,
		[]string{}, //empty dns name for client certs
		organizationalUnits,
		req.ValidityDays,
	)

	if err != nil {
		ctx.Logger.Error("failed to create certificate request",
			"error", err,
			"user", user.Username,
			"common_name", commonName,
		)

		ctx.SetJSONError(http.StatusInternalServerError, "failed to create certificate request")
		return
	}

	ctx.Logger.Debug("certificate request created",
		"request_id", certRequest.ID,
		"user", user.Username,
		"common_name", commonName,
		"ous", organizationalUnits,
	)

	ctx.WriteJSON(http.StatusCreated, certRequest)
}

func GETCertificateRequests(ctx *middlewares.AppContext) {
	panic("implement me")
}

func GETCertificateRequest(ctx *middlewares.AppContext) {
	panic("implement me")
}

func POSTCertificateRequestReview(ctx *middlewares.AppContext) {
	panic("implement me")
}

func GETUserCertificateRequests(ctx *middlewares.AppContext) {
	panic("implement me")
}
