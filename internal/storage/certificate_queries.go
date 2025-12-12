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

func (q *CertificateQueries) GetRequests(ctx context.Context) ([]*models.CertificateRequest, error) {
	query := `
       SELECT id, owner_iss, owner_sub, message, common_name, dns_names, organizational_units, validity_days, status, requested_at, issued_at, expires_at, serial_number, certificate_pem
       FROM certificate_requests
       ORDER BY requested_at DESC
    `

	eventsQuery := `
       SELECT id, certificate_request_id, requester_iss, requester_sub, reviewer_iss, reviewer_sub, new_status, review_notes, created_at
       FROM certificate_events
       WHERE certificate_request_id = ANY($1)
       ORDER BY certificate_request_id, created_at
    `

	rows, err := q.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate requests: %w", err)
	}
	defer rows.Close()

	var requests []*models.CertificateRequest
	var requestIDs []int

	for rows.Next() {
		var req models.CertificateRequest
		if err := rows.Scan(
			&req.ID,
			&req.OwnerIss,
			&req.OwnerSub,
			&req.Message,
			&req.CommonName,
			&req.DNSNames,
			&req.OrganizationalUnits,
			&req.ValidityDays,
			&req.Status,
			&req.RequestedAt,
			&req.IssuedAt,
			&req.ExpiresAt,
			&req.SerialNumber,
			&req.CertificatePem,
		); err != nil {
			return nil, fmt.Errorf("failed to scan certificate request: %w", err)
		}
		req.Events = []models.CertificateEvent{}
		requests = append(requests, &req)
		requestIDs = append(requestIDs, req.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate certificate requests: %w", err)
	}

	if len(requests) == 0 {
		return requests, nil
	}

	// Fetch all events for all requests in one query
	eventRows, err := q.pool.Query(ctx, eventsQuery, requestIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for certificate requests: %w", err)
	}
	defer eventRows.Close()

	// Create a map for quick lookup
	requestMap := make(map[int]*models.CertificateRequest)
	for _, req := range requests {
		requestMap[req.ID] = req
	}

	for eventRows.Next() {
		var event models.CertificateEvent
		if err := eventRows.Scan(
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

		if req, ok := requestMap[event.CertificateRequestID]; ok {
			req.Events = append(req.Events, event)
		}
	}

	if err := eventRows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate certificate events: %w", err)
	}

	return requests, nil
}

func (q *CertificateQueries) GetRequestsByUser(ctx context.Context, sub, iss string) ([]*models.CertificateRequest, error) {
	query := `
       SELECT id, owner_iss, owner_sub, message, common_name, dns_names, organizational_units, validity_days, status, requested_at, issued_at, expires_at, serial_number, certificate_pem
       FROM certificate_requests
       WHERE owner_sub = $1 AND  owner_iss = $2
       ORDER BY requested_at DESC
    `

	eventsQuery := `
       SELECT id, certificate_request_id, requester_iss, requester_sub, reviewer_iss, reviewer_sub, new_status, review_notes, created_at
       FROM certificate_events
       WHERE certificate_request_id = ANY($1)
       ORDER BY certificate_request_id, created_at
    `

	rows, err := q.pool.Query(ctx, query, sub, iss)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate requests for user '%s/%s': %w", iss, sub, err)
	}
	defer rows.Close()

	var requests []*models.CertificateRequest
	var requestIDs []int

	for rows.Next() {
		var req models.CertificateRequest
		if err := rows.Scan(
			&req.ID,
			&req.OwnerIss,
			&req.OwnerSub,
			&req.Message,
			&req.CommonName,
			&req.DNSNames,
			&req.OrganizationalUnits,
			&req.ValidityDays,
			&req.Status,
			&req.RequestedAt,
			&req.IssuedAt,
			&req.ExpiresAt,
			&req.SerialNumber,
			&req.CertificatePem,
		); err != nil {
			return nil, fmt.Errorf("failed to scan certificate request: %w", err)
		}
		req.Events = []models.CertificateEvent{}
		requests = append(requests, &req)
		requestIDs = append(requestIDs, req.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate certificate requests: %w", err)
	}

	if len(requests) == 0 {
		return requests, nil
	}

	// Fetch all events for all requests in one query
	eventRows, err := q.pool.Query(ctx, eventsQuery, requestIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for certificate requests: %w", err)
	}
	defer eventRows.Close()

	// Create a map for quick lookup
	requestMap := make(map[int]*models.CertificateRequest)
	for _, req := range requests {
		requestMap[req.ID] = req
	}

	for eventRows.Next() {
		var event models.CertificateEvent
		if err := eventRows.Scan(
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

		if req, ok := requestMap[event.CertificateRequestID]; ok {
			req.Events = append(req.Events, event)
		}
	}

	if err := eventRows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate certificate events: %w", err)
	}

	return requests, nil
}

func (q *CertificateQueries) GetRequestsPaginated(ctx context.Context, params models.PaginationParams) (*models.PaginatedResult, error) {
	// Set default limit if not provided
	if params.Limit <= 0 {
		params.Limit = 20
	}
	// Cap maximum limit to prevent abuse
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM certificate_requests`
	var total int
	if err := q.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get paginated requests
	query := `
       SELECT id, owner_iss, owner_sub, message, common_name, dns_names, organizational_units, validity_days, status, requested_at, issued_at, expires_at, serial_number, certificate_pem
       FROM certificate_requests
       ORDER BY requested_at DESC
       LIMIT $1 OFFSET $2
    `

	rows, err := q.pool.Query(ctx, query, params.Limit, params.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate requests: %w", err)
	}
	defer rows.Close()

	var requests []*models.CertificateRequest
	var requestIDs []int

	for rows.Next() {
		var req models.CertificateRequest
		if err := rows.Scan(
			&req.ID,
			&req.OwnerIss,
			&req.OwnerSub,
			&req.Message,
			&req.CommonName,
			&req.DNSNames,
			&req.OrganizationalUnits,
			&req.ValidityDays,
			&req.Status,
			&req.RequestedAt,
			&req.IssuedAt,
			&req.ExpiresAt,
			&req.SerialNumber,
			&req.CertificatePem,
		); err != nil {
			return nil, fmt.Errorf("failed to scan certificate request: %w", err)
		}
		req.Events = []models.CertificateEvent{}
		requests = append(requests, &req)
		requestIDs = append(requestIDs, req.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate certificate requests: %w", err)
	}

	result := &models.PaginatedResult{
		Requests: requests,
		Total:    total,
		Limit:    params.Limit,
		Offset:   params.Offset,
		HasMore:  params.Offset+len(requests) < total,
	}

	if len(requests) == 0 {
		return result, nil
	}

	// Fetch all events for the paginated requests
	eventsQuery := `
       SELECT id, certificate_request_id, requester_iss, requester_sub, reviewer_iss, reviewer_sub, new_status, review_notes, created_at
       FROM certificate_events
       WHERE certificate_request_id = ANY($1)
       ORDER BY certificate_request_id, created_at
    `

	eventRows, err := q.pool.Query(ctx, eventsQuery, requestIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for certificate requests: %w", err)
	}
	defer eventRows.Close()

	// Create a map for quick lookup
	requestMap := make(map[int]*models.CertificateRequest)
	for _, req := range requests {
		requestMap[req.ID] = req
	}

	for eventRows.Next() {
		var event models.CertificateEvent
		if err := eventRows.Scan(
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

		if req, ok := requestMap[event.CertificateRequestID]; ok {
			req.Events = append(req.Events, event)
		}
	}

	if err := eventRows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate certificate events: %w", err)
	}

	return result, nil
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
