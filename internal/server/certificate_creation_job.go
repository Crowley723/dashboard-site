package server

import (
	"context"
	"errors"
	"fmt"
	"homelab-dashboard/internal/k8s"
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/models"
	"strings"
	"time"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
)

type CertificateCreationJob struct {
	appCtx   *middlewares.AppContext
	interval time.Duration
}

func NewCertificateCreationJob(appCtx *middlewares.AppContext, interval time.Duration) *CertificateCreationJob {
	return &CertificateCreationJob{
		appCtx:   appCtx,
		interval: interval,
	}
}

func (j *CertificateCreationJob) Name() string {
	return "create_certificate"
}

func (j *CertificateCreationJob) RequiresLeadership() bool {
	return true
}

func (j *CertificateCreationJob) Interval() time.Duration {
	return j.interval
}

func (j *CertificateCreationJob) Run(ctx context.Context) error {
	if j.interval <= 0 {
		j.appCtx.Logger.Error("certificate creation job failed: ticker interval must not be zero")
		return fmt.Errorf("non-positive ticker interval: %s", j.interval)
	}

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	certs, err := getApprovedCertificates(j.appCtx)
	if err != nil {
		if !errors.Is(err, errNoApprovedCertificates) {
			j.appCtx.Logger.Error("error checking for approved certificates", "error", err)
		}
	}

	err = handleApprovedCertificates(j.appCtx, certs)
	if err != nil {
		return fmt.Errorf("unable to create certificate CRD: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			j.appCtx.Logger.Debug("Background data fetching canceled")
			return ctx.Err()
		case <-ticker.C:
			certs, err := getApprovedCertificates(j.appCtx)
			if err != nil {
				if !errors.Is(err, errNoApprovedCertificates) {
					j.appCtx.Logger.Error("unable to check for approved certificates", "error", err)
				}
			}

			err = handleApprovedCertificates(j.appCtx, certs)
			if err != nil {
				return fmt.Errorf("unable to create certificate CRD: %w", err)
			}
		}
	}
}

func getApprovedCertificates(ctx *middlewares.AppContext) ([]*models.CertificateRequest, error) {
	certs, err := ctx.Storage.Certificates().GetApprovedRequests(ctx)
	if err != nil {
		return nil, err
	}

	if len(certs) == 0 || certs == nil {
		return nil, errNoApprovedCertificates
	}

	return certs, nil
}

func handleApprovedCertificates(ctx *middlewares.AppContext, certs []*models.CertificateRequest) error {
	for _, cert := range certs {
		systemIss, systemSub, err := ctx.Storage.GetSystemUser(ctx)
		if err != nil {
			ctx.Logger.Error("error getting system user", "error", err)
			continue
		}

		// Mark it as pending
		err = ctx.Storage.Certificates().UpdateCertificateStatus(
			ctx,
			cert.ID,
			models.StatusPending,
			systemIss,
			systemSub,
			"Pending certificate issuance",
		)
		if err != nil {
			ctx.Logger.Error("error reserving certificate request", "error", err, "request_id", cert.ID)
			continue
		}

		// Check if K8s certificate already exists
		// This handles cases where cert was created but DB update failed
		certName := k8s.GenerateCertificateName(cert.OwnerSub, cert.OwnerIss, cert.RequestedAt)
		existingCert, err := ctx.KubernetesClient.GetCertificate(ctx, ctx.KubernetesClient.Namespace, certName)

		var createdCert *certmanagerv1.Certificate
		if err != nil {
			if isNotFoundError(err) {
				createdCert, err = ctx.KubernetesClient.CreateCertificateFromRequest(ctx, cert)
				if err != nil {
					ctx.Logger.Error("error creating certificate from request", "error", err, "request_id", cert.ID)
					// Rollback: mark as APPROVED again so it can be retried later
					rollbackErr := ctx.Storage.Certificates().UpdateCertificateStatus(
						ctx,
						cert.ID,
						models.StatusApproved,
						systemIss,
						systemSub,
						fmt.Sprintf("K8s certificate creation failed"),
					)
					if rollbackErr != nil {
						ctx.Logger.Error("error rolling back certificate status", "error", rollbackErr, "request_id", cert.ID)
					}
					continue
				}
			} else {
				ctx.Logger.Error("error checking for existing certificate", "error", err, "request_id", cert.ID)
				continue
			}
		} else {
			ctx.Logger.Debug("certificate already exists in k8s, reusing", "cert_name", certName, "request_id", cert.ID)
			createdCert = existingCert
		}

		err = ctx.Storage.Certificates().UpdateCertificateK8sMetadata(
			ctx,
			cert.ID,
			createdCert.Name,
			createdCert.Namespace,
			createdCert.Spec.SecretName,
		)
		if err != nil {
			ctx.Logger.Error("error updating certificate metadata", "error", err, "request_id", cert.ID)
			// Don't continue - metadata update is critical but cert exists in K8s
			// Status polling job will eventually detect this certificate
		}

		ctx.Logger.Debug("Certificate Creation Completed",
			"request_id", cert.ID,
			"k8s_name", createdCert.Name,
			"namespace", createdCert.Namespace)
	}
	return nil
}

// isNotFoundError checks if the error is a Kubernetes NotFound error
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Check if error message contains "not found" (GetCertificate returns formatted error)
	return strings.Contains(strings.ToLower(err.Error()), "not found")
}
