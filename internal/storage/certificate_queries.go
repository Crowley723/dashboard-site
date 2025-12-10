package storage

import (
	"context"
	"fmt"
	"homelab-dashboard/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CertificateQueries struct {
	pool *pgxpool.Pool
}

func NewCertificateQueries(pool *pgxpool.Pool) *CertificateQueries {
	return &CertificateQueries{pool: pool}
}

// CreateRequest adds a certificate request to the database.
func (q *CertificateQueries) CreateRequest(ctx context.Context, sub, iss, commonName, status, message string, dnsNames, organizationalUnits []string, validityDays int) (int, error) {
	query := `
		INSERT INTO certificate_requests (sub, iss, common_name, status, message, dns_names, organizational_units, validity_days )
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`

	var requestID int
	err := q.pool.QueryRow(ctx, query,
		sub, iss, commonName, message,
		dnsNames, organizationalUnits, validityDays,
	).Scan(&requestID)

	if err != nil {
		return -1, fmt.Errorf("failed to create certificate request: %w", err)
	}

	return q.GetByID(ctx, iss, sub)
}

func (q *CertificateQueries) GetRequestByID(ctx context.Context, id int) (models.CertificateRequest, error) {
}
