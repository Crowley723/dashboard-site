package storage

import (
	"context"
	"errors"
	"fmt"
	"homelab-dashboard/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CertificateQueries struct {
	pool *pgxpool.Pool
}

func NewCertificateQueries(pool *pgxpool.Pool) *CertificateQueries {
	return &CertificateQueries{pool: pool}
}

// CreateRequest adds a certificate request to the database.
func (q *CertificateQueries) CreateRequest(ctx context.Context, sub, iss, commonName, status, message string, dnsNames, organizationalUnits []string, validityDays int) (*models.CertificateRequest, error) {
	query := `
		INSERT INTO certificate_requests (owner_sub, owner_iss, common_name, status, message, dns_names, organizational_units, validity_days)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var requestID int
	err := q.pool.QueryRow(ctx, query,
		sub, iss, commonName, status, message,
		dnsNames, organizationalUnits, validityDays,
	).Scan(&requestID)

	if err != nil {
		return nil, fmt.Errorf("failed to create certificate request: %w", err)
	}

	return q.GetRequestByID(ctx, requestID)
}

func (q *CertificateQueries) GetRequestByID(ctx context.Context, id int) (*models.CertificateRequest, error) {
	query := `
		SELECT id, owner_iss, owner_sub, message, common_name, dns_names, organizational_units, validity_days, status, requested_at, issued_at, expires_at, serial_number, certificate_pem
		FROM certificate_requests
		WHERE id = $1
	`

	eventsQuery := `
		SELECT id, certificate_request_id, requester_iss, requester_sub, reviewer_iss, reviewer_sub, new_status, review_notes, created_at
		FROM certificate_events
		WHERE certificate_request_id = $1
		`

	var certificateRequest models.CertificateRequest
	err := q.pool.QueryRow(ctx, query, id).Scan(
		&certificateRequest.ID,
		&certificateRequest.OwnerIss,
		&certificateRequest.OwnerSub,
		&certificateRequest.Message,
		&certificateRequest.CommonName,
		&certificateRequest.DNSNames,
		&certificateRequest.OrganizationalUnits,
		&certificateRequest.ValidityDays,
		&certificateRequest.Status,
		&certificateRequest.RequestedAt,
		&certificateRequest.IssuedAt,
		&certificateRequest.ExpiresAt,
		&certificateRequest.SerialNumber,
		&certificateRequest.CertificatePem,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("certificate request not found")
		}
		return nil, fmt.Errorf("failed to get certificate request by id: %w", err)
	}

	rows, err := q.pool.Query(ctx, eventsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for certificate request '%d': %w", id, err)
	}
	defer rows.Close()

	var events []models.CertificateEvent
	for rows.Next() {
		var event models.CertificateEvent
		if err := rows.Scan(
			&event.ID,
			&event.CertificateRequestID,
			&event.RequesterIss,
			&event.RequesterSub,
			&event.ReviewerIss,
			&event.ReviewerSub,
			&event.NewStatus,
			&event.ReviewNotes,
			&event.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate certificate events: %w", err)
	}

	certificateRequest.Events = events
	if certificateRequest.Events == nil {
		certificateRequest.Events = []models.CertificateEvent{}
	}

	return &certificateRequest, nil
}

func (q *CertificateQueries) UpdateCertificateStatus(ctx context.Context, requestId int, newStatus models.CertificateRequestStatus, reviewerIss, reviewerSub, notes string) error {
	tx, err := q.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("transaction start failed: %w", err)
	}
	defer tx.Rollback(ctx)

	getRequestStatusQuery := `
		SELECT status, owner_iss, owner_sub 
		FROM certificate_requests 
		WHERE id = $1
		`

	var currentStatus models.CertificateRequestStatus
	var requesterIss, requesterSub string
	err = tx.QueryRow(ctx, getRequestStatusQuery, requestId).Scan(&currentStatus, &requesterIss, &requesterSub)
	if err != nil {
		return fmt.Errorf("failed to get current status for certificate request '%d': %w", requestId, err)
	}

	updateStatusQuery := `
		UPDATE certificate_requests
		SET status = $1
		WHERE id = $2
		`

	_, err = tx.Exec(ctx, updateStatusQuery, newStatus, requestId)
	if err != nil {
		return fmt.Errorf("failed to update status for certificate request '%d': %w", requestId, err)
	}

	insertEventQuery := `
		INSERT INTO certificate_events
		(certificate_request_id, requester_iss, requester_sub, reviewer_iss, reviewer_sub, new_status, review_notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

	_, err = tx.Exec(ctx, insertEventQuery, requestId, requesterIss, requesterSub, reviewerIss, reviewerSub, newStatus, notes)
	if err != nil {
		return fmt.Errorf("failed to insert event for certificate request '%d': %w", requestId, err)
	}

	return tx.Commit(ctx)
}
