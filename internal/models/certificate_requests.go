package models

type CertificateRequestStatus int

const (
	_ CertificateRequestStatus = iota
	PENDING
	COMPLETED
	REJECTED
	ISSUED
)

type CertificateRequest struct {
	ID                  string   `json:"id"`
	OwnerIss            string   `json:"owner_iss"`
	OwnerSub            string   `json:"owner_sub"`
	Message             string   `json:"message"`
	CommonName          string   `json:"common_name"`
	DNSNames            []string `json:"dns_names"`
	OrganizationalUnits []string `json:"organizational_units"`
	ValidityDays        int      `json:"validity_days"`
	Status              string   `json:"status"`
}
