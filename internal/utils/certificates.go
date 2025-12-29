package utils

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"homelab-dashboard/internal/models"
)

// ParseCertificateDetails extracts details from a PEM-encoded certificate
func ParseCertificateDetails(certPEM []byte) (*models.IssuedCertificateDetails, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return &models.IssuedCertificateDetails{
		SerialNumber: cert.SerialNumber.String(),
		Subject:      cert.Subject.String(),
		Issuer:       cert.Issuer.String(),
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		DNSNames:     cert.DNSNames,
		CommonName:   cert.Subject.CommonName,
		Organization: cert.Subject.Organization,
	}, nil
}
