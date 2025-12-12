package handlers

import (
	"encoding/json"
	"fmt"
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/models"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
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

//// Get first page (20 items)
//result, err := ctx.Storage.Certificates().GetRequestsPaginated(ctx, models.PaginationParams{
//	Limit:  20,
//	Offset: 0,
//})
//
//// Get second page
//result, err := ctx.Storage.Certificates().GetRequestsPaginated(ctx, models.PaginationParams{
//	Limit:  20,
//	Offset: 20,
//})
//
//// Check if there are more pages
//if result.HasMore {
//	fmt.Printf("Showing %d-%d of %d total\n",
//		result.Offset+1,
//		result.Offset+len(result.Requests),
//		result.Total)
//}

func GETCertificateRequests(ctx *middlewares.AppContext) {
	user, ok := ctx.SessionManager.GetAuthenticatedUser(ctx)
	if !ok || user == nil {
		ctx.SetJSONError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	requests, err := ctx.Storage.Certificates().GetRequests(ctx)
	if err != nil {
		ctx.Logger.Error("failed to get certificate requests",
			"error", err)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to get certificate requests")
		return
	}

	if requests == nil {
		ctx.SetJSONStatus(http.StatusOK, "No certificate requests found")
		return
	}

	ctx.WriteJSON(http.StatusOK, requests)
}

func GETCertificateRequest(ctx *middlewares.AppContext) {
	requestIdParam := chi.URLParam(ctx.Request, "id")
	if requestIdParam == "" {
		ctx.SetJSONError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	requestId, err := strconv.Atoi(strings.TrimSpace(requestIdParam))
	if err != nil {
		ctx.SetJSONError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	user, ok := ctx.SessionManager.GetAuthenticatedUser(ctx)
	if !ok || user == nil {
		ctx.SetJSONError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	requests, err := ctx.Storage.Certificates().GetRequestByID(ctx, requestId)
	if err != nil {
		ctx.Logger.Error("failed to get certificate requests",
			"error", err)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to get certificate requests")
		return
	}

	if requests != nil && ((requests.OwnerSub != user.Username || requests.OwnerIss != user.Iss) && !slices.Contains(user.Groups, ctx.Config.Features.MTLSManagement.AdminGroup)) {
		ctx.SetJSONError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	if requests == nil {
		ctx.SetJSONStatus(http.StatusOK, "No certificate requests found")
		return
	}

	ctx.WriteJSON(http.StatusOK, requests)
}

func POSTCertificateRequestReview(ctx *middlewares.AppContext) {
	panic("implement me")
}

func GETUserCertificateRequests(ctx *middlewares.AppContext) {
	user, ok := ctx.SessionManager.GetAuthenticatedUser(ctx)
	if !ok || user == nil {
		ctx.SetJSONError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	requests, err := ctx.Storage.Certificates().GetRequestsByUser(ctx, user.Sub, user.Iss)
	if err != nil {
		ctx.Logger.Error("failed to get certificate requests",
			"error", err)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to get certificate requests")
		return
	}

	if requests == nil {
		ctx.SetJSONStatus(http.StatusOK, "No certificate requests found")
		return
	}

	ctx.WriteJSON(http.StatusOK, requests)
}
