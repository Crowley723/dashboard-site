package handlers

import (
	"encoding/json"
	"fmt"
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/models"
	"homelab-dashboard/internal/storage"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type CertificateRequestResponse struct {
	ID                  int                             `json:"id"`
	OwnerIss            string                          `json:"owner_iss"`
	OwnerSub            string                          `json:"owner_sub"`
	OwnerUsername       string                          `json:"owner_username"`
	OwnerDisplayName    string                          `json:"owner_display_name"`
	Message             string                          `json:"message,omitempty"`
	Events              []models.CertificateEvent       `json:"events,omitempty"`
	CommonName          string                          `json:"common_name"`
	DNSNames            []string                        `json:"dns_names,omitempty"`
	OrganizationalUnits []string                        `json:"organizational_units,omitempty"`
	ValidityDays        int                             `json:"validity_days"`
	Status              models.CertificateRequestStatus `json:"status,omitempty"`
	RequestedAt         time.Time                       `json:"requested_at"`
	IssuedAt            *time.Time                      `json:"issued_at,omitempty"`
	ExpiresAt           *time.Time                      `json:"expires_at,omitempty"`
	SerialNumber        *string                         `json:"serial_number,omitempty"`
}

// POSTCertificateRequest is used by any authenticated user to create a certificate request
func POSTCertificateRequest(ctx *middlewares.AppContext) {
	user, ok := ctx.SessionManager.GetAuthenticatedUser(ctx)
	if !ok {
		ctx.SetJSONError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	if !slices.Contains(user.Groups, ctx.Config.Features.MTLSManagement.UserGroup) && !slices.Contains(user.Groups, ctx.Config.Features.MTLSManagement.AdminGroup) {
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

	certRequest, err := ctx.Storage.CreateCertificateRequest(
		ctx,
		user.Sub,
		user.Iss,
		commonName,
		string(models.StatusAwaitingReview),
		req.Message,
		[]string{}, //empty dns name for client certs
		nil,        // not implemented
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
		err = ctx.Storage.UpdateCertificateRequestStatus(ctx, certRequest.ID, models.StatusApproved, ctx.Config.Server.ExternalURL, storage.SystemSub, "Auto Approved")
		if err != nil {
			ctx.Logger.Error("failed to auto approve certificate request", "error", err)
			ctx.SetJSONError(http.StatusInternalServerError, "Failed to auto approve certificate request")
			return
		}
		ctx.Logger.Debug("request was auto-approved", "iss", user.Iss, "sub", user.Sub, "request_status", requestStatus)
	}

	updatedRequest, err := ctx.Storage.GetCertificateRequestByID(ctx, certRequest.ID)
	if err != nil {
		ctx.Logger.Error("failed to get certificate requests",
			"error", err)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to get certificate requests")
		return
	}

	ctx.WriteJSON(http.StatusCreated, updatedRequest)
}

// GETCertificateRequests is used to expose all certificate requests to admin users. Admin check done with middleware
func GETCertificateRequests(ctx *middlewares.AppContext) {
	user, ok := ctx.SessionManager.GetAuthenticatedUser(ctx)
	if !ok || user == nil {
		ctx.SetJSONError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	requests, err := ctx.Storage.GetCertificateRequests(ctx)
	if err != nil {
		ctx.Logger.Error("failed to get certificate requests",
			"error", err)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to get certificate requests")
		return
	}

	if requests == nil {
		ctx.WriteJSON(http.StatusOK, []interface{}{})
		return
	}

	ctx.WriteJSON(http.StatusOK, redactCertificateFields(requests))
}

// GETCertificateRequest is used to expose a single certificate request
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

	requests, err := ctx.Storage.GetCertificateRequestByID(ctx, requestId)
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
		ctx.WriteJSON(http.StatusOK, []interface{}{})
		return
	}

	ctx.WriteJSON(http.StatusOK, redactCertificateFields(
		[]*models.CertificateRequest{
			requests,
		}))
}

// POSTCertificateReview is used by admins to post a reject/approval for a specific certificate request with comments. Admin check done with middleware
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

	request, err := ctx.Storage.GetCertificateRequestByID(ctx, requestId)
	if err != nil {
		ctx.Logger.Error("failed to get certificate requests",
			"error", err)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to fetch certificate request")
		return
	}

	if request == nil {
		ctx.SetJSONStatus(http.StatusOK, "No certificate requests found")
		return
	}

	if !ctx.Config.Features.MTLSManagement.AllowAdminsToApproveOwnRequests && user.MatchesUser(request.OwnerIss, request.OwnerSub) {
		ctx.SetJSONError(http.StatusForbidden, "You are not allowed to approve your own requests")
		return
	}

	if request.Status != models.StatusAwaitingReview {
		ctx.SetJSONError(http.StatusBadRequest,
			fmt.Sprintf("Cannot review request with status '%s'. Only requests with status 'awaiting_review' can be reviewed.",
				request.Status))
		return
	}

	err = ctx.Storage.UpdateCertificateRequestStatus(ctx, request.ID, review.NewStatus, user.Iss, user.Sub, review.ReviewNotes)
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

	updatedRequest, err := ctx.Storage.GetCertificateRequestByID(ctx, requestId)
	if err != nil {
		ctx.Logger.Error("failed to get certificate requests",
			"error", err)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to get certificate requests")
		return
	}

	ctx.WriteJSON(http.StatusOK, updatedRequest)
}

// GETUserCertificateRequests exposes information about certificate requests to the owner
func GETUserCertificateRequests(ctx *middlewares.AppContext) {
	user, ok := ctx.SessionManager.GetAuthenticatedUser(ctx)
	if !ok || user == nil {
		ctx.SetJSONError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	requests, err := ctx.Storage.GetCertificateRequestsByUser(ctx, user.Sub, user.Iss)
	if err != nil {
		ctx.Logger.Error("failed to get certificate requests",
			"error", err)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to get certificate requests")
		return
	}

	if requests == nil {
		ctx.WriteJSON(http.StatusOK, []interface{}{})
		return
	}

	ctx.WriteJSON(http.StatusOK, redactCertificateFields(requests))
}

func redactCertificateFields(requests []*models.CertificateRequest) []*models.CertificateRequest {
	result := make([]*models.CertificateRequest, len(requests))
	for i, req := range requests {
		reqCopy := *req
		reqCopy.K8sCertificateName = nil
		reqCopy.K8sNamespace = nil
		reqCopy.K8sSecretName = nil
		reqCopy.CertificatePem = nil
		result[i] = &reqCopy
	}
	return result
}
