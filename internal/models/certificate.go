package models

import (
	"time"
)

type CertificateRequest struct {
	ID       int                `json:"id"`
	OwnerIss string             `json:"owner_iss"`
	OwnerSub string             `json:"owner_sub"`
	Message  string             `json:"message,omitempty"`
	Events   []CertificateEvent `json:"events,omitempty"`

	CommonName          string   `json:"common_name"`
	DNSNames            []string `json:"dns_names,omitempty"`
	OrganizationalUnits []string `json:"organizational_units,omitempty"`
	ValidityDays        int      `json:"validity_days"`

	Status      CertificateRequestStatus `json:"status,omitempty"`
	RequestedAt time.Time                `json:"requested_at"`

	IssuedAt       *time.Time `json:"issued_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	SerialNumber   *string    `json:"serial_number,omitempty"`
	CertificatePem *string    `json:"certificate_pem,omitempty"`
}

type CertificateEvent struct {
	ID                   int    `json:"id"`
	CertificateRequestID int    `json:"certificate_request_id"`
	RequesterIss         string `json:"requester_iss"`
	RequesterSub         string `json:"requester_sub"`
	ReviewerIss          string `json:"reviewer_iss"`
	ReviewerSub          string `json:"reviewer_sub"`

	NewStatus   CertificateRequestStatus `json:"new_status"`
	ReviewNotes string                   `json:"review_notes"`
	CreatedAt   time.Time                `json:"created_at"`
}

type CertificateRequestStatus string

const (
	StatusAwaitingReview CertificateRequestStatus = "awaiting_review"
	StatusApproved       CertificateRequestStatus = "approved"
	StatusRejected       CertificateRequestStatus = "rejected"
	StatusPending        CertificateRequestStatus = "pending" // waiting for issuance
	StatusIssued         CertificateRequestStatus = "issued"
	StatusFailed         CertificateRequestStatus = "failed"
	StatusCompleted      CertificateRequestStatus = "completed"
)

func (s CertificateRequestStatus) CanTransitionTo(next CertificateRequestStatus) bool {
	// Rejected, Failed, and Completed are final states
	validTransitions := map[CertificateRequestStatus][]CertificateRequestStatus{
		StatusAwaitingReview: {StatusApproved, StatusRejected},
		StatusApproved:       {StatusPending},
		StatusPending:        {StatusIssued, StatusFailed},
		StatusIssued:         {StatusCompleted},
	}

	allowed, ok := validTransitions[s]
	if !ok {
		return false
	}

	for _, valid := range allowed {
		if valid == next {
			return true
		}
	}

	return false
}

func (s CertificateRequestStatus) IsTerminal() bool {
	return s == StatusRejected || s == StatusFailed || s == StatusCompleted
}

func (s CertificateRequestStatus) RequiresAction() bool {
	return s == StatusAwaitingReview || s == StatusIssued
}
