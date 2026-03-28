package certificate

import (
	"context"
	"homelab-dashboard/internal/models"
)

type Provider interface {
	CreateCertificateFromRequest(ctx context.Context, request *models.CertificateRequest) (name string, err error)
	GetCertificateData(ctx context.Context, name string) (certPEM, keyPEM, caPEM []byte, err error)
	IsCertificateReady(ctx context.Context, name string) (ready bool, err error)
	DeleteCertificate(ctx context.Context, name string) error
}
