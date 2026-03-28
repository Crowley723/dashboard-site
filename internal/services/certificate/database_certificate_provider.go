package certificate

import (
	"context"
	"homelab-dashboard/internal/models"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
)

type DatabaseCertificateProvider struct{}

func (DatabaseCertificateProvider) CreateCertificateFromRequest(ctx context.Context, request *models.CertificateRequest) (*certmanagerv1.Certificate, error) {
	//TODO implement me
	panic("implement me")
}

func (DatabaseCertificateProvider) GetCertificateData(ctx context.Context, name string) (certPEM, keyPEM, caPEM []byte, err error) {
	//TODO implement me
	panic("implement me")
}

func (DatabaseCertificateProvider) IsCertificateReady(ctx context.Context, name string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (DatabaseCertificateProvider) DeleteCertificate(ctx context.Context, name string) error {
	//TODO implement me
	panic("implement me")
}
