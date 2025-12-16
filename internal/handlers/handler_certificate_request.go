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

	requestStatus := string(models.StatusAwaitingReview)

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

	if ctx.Config.Features.MTLSManagement.AutoApproveAdminRequests && slices.Contains(user.Groups, ctx.Config.Features.MTLSManagement.AdminGroup) {
		err = ctx.Storage.Certificates().UpdateCertificateStatus(ctx, certRequest.ID, models.StatusApproved, user.Iss, user.Sub, "Auto Approved")
		if err != nil {
			ctx.Logger.Error("failed to auto approve certificate request", "error", err)
			ctx.SetJSONError(http.StatusInternalServerError, "Failed to auto approve certificate request")
			return
		}
		ctx.Logger.Debug("request is auto-approved", "iss", user.Iss, "sub", user.Sub, "request_status", requestStatus)
	}

	updatedRequest, _ := ctx.Storage.Certificates().GetRequestByID(ctx, certRequest.ID)
	ctx.WriteJSON(http.StatusCreated, updatedRequest)
}

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

	if requests != nil &&
		!user.MatchesUser(requests.OwnerIss, requests.OwnerSub) &&
		!slices.Contains(user.Groups, ctx.Config.Features.MTLSManagement.AdminGroup) {
		ctx.SetJSONError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	if requests == nil {
		ctx.SetJSONStatus(http.StatusOK, "No certificate requests found")
		return
	}

	ctx.WriteJSON(http.StatusOK, requests)
}

func POSTCertificateReview(ctx *middlewares.AppContext) {
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

	var review struct {
		NewStatus   models.CertificateRequestStatus `json:"new_status"`
		ReviewNotes string                          `json:"review_notes"`
	}

	if err := json.NewDecoder(ctx.Request.Body).Decode(&review); err != nil {
		ctx.SetJSONError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	review.NewStatus = models.CertificateRequestStatus(strings.TrimSpace(string(review.NewStatus)))

	if review.NewStatus != models.StatusApproved && review.NewStatus != models.StatusRejected {
		ctx.SetJSONError(http.StatusBadRequest,
			"Invalid status. Must be 'approved' or 'rejected'")
		return
	}

	request, err := ctx.Storage.Certificates().GetRequestByID(ctx, requestId)
	if err != nil {
		ctx.Logger.Error("failed to get certificate requests",
			"error", err)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to fetch certificate request")
		return
	}

	if ctx.Config.Features.MTLSManagement.AllowAdminsToApproveOwnRequests && user.MatchesUser(request.OwnerIss, request.OwnerSub) {
		ctx.SetJSONError(http.StatusForbidden, "You are not allowed to approve your own requests")
		return
	}

	if request == nil {
		ctx.SetJSONStatus(http.StatusOK, "No certificate requests found")
		return
	}

	if request.Status != models.StatusAwaitingReview {
		ctx.SetJSONError(http.StatusBadRequest,
			fmt.Sprintf("Cannot review request with status '%s'. Only requests with status 'awaiting_review' can be reviewed.",
				request.Status))
		return
	}

	err = ctx.Storage.Certificates().UpdateCertificateStatus(ctx, request.ID, review.NewStatus, user.Iss, user.Sub, review.ReviewNotes)
	if err != nil {
		ctx.Logger.Error("failed to update certificate request status", "error", err)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to update certificate request status")
		return
	}

	ctx.Logger.Info("certificate request reviewed",
		"request_id", requestId,
		"reviewer", user.Username,
		"new_status", review.NewStatus,
	)

	updatedRequest, _ := ctx.Storage.Certificates().GetRequestByID(ctx, requestId)
	ctx.WriteJSON(http.StatusOK, updatedRequest)
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
