package models

import (
	"time"
)

type CertificateRequest struct {
	ID               int                `json:"id"`
	OwnerIss         string             `json:"owner_iss"`
	OwnerSub         string             `json:"owner_sub"`
	OwnerUsername    string             `json:"owner_username"`
	OwnerDisplayName string             `json:"owner_display_name"`
	Message          string             `json:"message,omitempty"`
	Events           []CertificateEvent `json:"events,omitempty"`

	CommonName          string   `json:"common_name"`
	DNSNames            []string `json:"dns_names,omitempty"`
	OrganizationalUnits []string `json:"organizational_units,omitempty"`
	ValidityDays        int      `json:"validity_days"`

	Status      CertificateRequestStatus `json:"status,omitempty"`
	RequestedAt time.Time                `json:"requested_at"`

	// Kubernetes Certificate metadata
	K8sCertificateName *string `json:"k8s_certificate_name,omitempty"`
	K8sNamespace       *string `json:"k8s_namespace,omitempty"`
	K8sSecretName      *string `json:"k8s_secret_name,omitempty"`

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
	RequesterUsername    string `json:"requester_username"`
	RequesterDisplayName string `json:"requester_display_name"`
	ReviewerIss          string `json:"reviewer_iss"`
	ReviewerSub          string `json:"reviewer_sub"`
	ReviewerUsername     string `json:"reviewer_username"`
	ReviewerDisplayName  string `json:"reviewer_display_name"`

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

// Rejected, Failed, and Completed are final states
var validTransitions = map[CertificateRequestStatus][]CertificateRequestStatus{
	StatusAwaitingReview: {StatusApproved, StatusRejected},
	StatusApproved:       {StatusPending},
	StatusPending:        {StatusIssued, StatusFailed},
	StatusIssued:         {StatusCompleted},
}

func (s CertificateRequestStatus) CanTransitionTo(next CertificateRequestStatus) bool {
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

func (s CertificateRequestStatus) GetValidTransitions() []CertificateRequestStatus {
	return validTransitions[s]
}

func (s CertificateRequestStatus) IsTerminal() bool {
	return s == StatusRejected || s == StatusFailed || s == StatusCompleted
}

func (s CertificateRequestStatus) RequiresAction() bool {
	return s == StatusAwaitingReview || s == StatusIssued
}

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Limit  int
	Offset int
}

// PaginatedCertResult holds paginated results
type PaginatedCertResult struct {
	Requests []*CertificateRequest
	Total    int
	Limit    int
	Offset   int
	HasMore  bool
}

// CertificateSpec contains the details for creating a Certificate resource
type CertificateSpec struct {
	Name                string
	Namespace           string
	CommonName          string
	DNSNames            []string
	OrganizationalUnits []string
	ValidityDays        int
	OwnerSub            string
	OwnerIss            string
	RequestID           int
}

// IssuedCertificateDetails contains the additional details related to certificates that have been issued
type IssuedCertificateDetails struct {
	SerialNumber string
	Subject      string
	Issuer       string
	NotBefore    time.Time
	NotAfter     time.Time
	DNSNames     []string
	CommonName   string
	Organization []string
}
