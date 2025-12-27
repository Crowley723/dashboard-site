package server

import (
	"context"
	"errors"
	"fmt"
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/models"
	"homelab-dashboard/internal/utils"
	"time"
)

type CertificateIssuedStatusJob struct {
	appCtx   *middlewares.AppContext
	interval time.Duration
}

func NewCertificateIssuedStatusJob(appCtx *middlewares.AppContext, interval time.Duration) *CertificateIssuedStatusJob {
	return &CertificateIssuedStatusJob{
		appCtx:   appCtx,
		interval: interval,
	}
}

func (j *CertificateIssuedStatusJob) Name() string {
	return "certificate_issued_status"
}

func (j *CertificateIssuedStatusJob) RequiresLeadership() bool {
	return true
}

func (j *CertificateIssuedStatusJob) Interval() time.Duration {
	return j.interval
}

func (j *CertificateIssuedStatusJob) Run(ctx context.Context) error {
	if j.interval <= 0 {
		j.appCtx.Logger.Error("certificate issued status job failed: ticker interval must not be zero")
		return fmt.Errorf("non-positive ticker interval: %s", j.interval)
	}

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	// Initial run
	certs, err := getIssuedCertificates(j.appCtx)
	if err != nil {
		if !errors.Is(err, errNoIssuedCertificates) {
			j.appCtx.Logger.Error("error checking for issued certificates", "error", err)
		}
	} else {
		err = handleIssuedCertificates(j.appCtx, certs)
		if err != nil {
			j.appCtx.Logger.Error("unable to update issued certificate status", "error", err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			j.appCtx.Logger.Info("Background issued certificate job canceled")
			return ctx.Err()
		case <-ticker.C:
			certs, err := getIssuedCertificates(j.appCtx)
			if err != nil {
				if !errors.Is(err, errNoIssuedCertificates) {
					j.appCtx.Logger.Error("unable to check for issued certificates", "error", err)
				}
				continue
			}

			err = handleIssuedCertificates(j.appCtx, certs)
			if err != nil {
				j.appCtx.Logger.Error("unable to update issued certificate status", "error", err)
			}
		}
	}
}

func getIssuedCertificates(ctx *middlewares.AppContext) ([]*models.CertificateRequest, error) {
	certs, err := ctx.Storage.Certificates().GetPendingRequests(ctx)
	if err != nil {
		return nil, err
	}

	if len(certs) == 0 || certs == nil {
		return nil, errNoIssuedCertificates
	}

	return certs, nil
}

func handleIssuedCertificates(ctx *middlewares.AppContext, certs []*models.CertificateRequest) error {
	systemIss, systemSub, err := ctx.Storage.GetSystemUser(ctx)
	if err != nil {
		return fmt.Errorf("error getting system user: %w", err)
	}

	for _, cert := range certs {
		k8sCert, err := ctx.KubernetesClient.GetCertificate(ctx, *cert.K8sNamespace, *cert.K8sCertificateName)
		if err != nil {
			ctx.Logger.Error("error getting certificate", "error", err, "name", *cert.K8sCertificateName)
			continue
		}

		if utils.IsCertificateReady(k8sCert) {
			certPEM, _, _, err := ctx.KubernetesClient.GetCertificatePEM(ctx, *cert.K8sNamespace, *cert.K8sCertificateName)
			if err != nil {
				ctx.Logger.Error("unable to get certificate PEM", "error", err, "name", *cert.K8sCertificateName)
				continue
			}

			certDetails, err := utils.ParseCertificateDetails(certPEM)
			if err != nil {
				ctx.Logger.Error("unable to parse certificate PEM", "error", err, "name", *cert.K8sCertificateName)
				continue
			}

			err = ctx.Storage.Certificates().UpdateCertificateIssued(ctx, cert.ID, string(certPEM), certDetails.SerialNumber, certDetails.NotBefore, certDetails.NotAfter, systemIss, systemSub)
			if err != nil {
				ctx.Logger.Error("unable to update certificate status", "error", err, "request_id", cert.ID)
				continue
			}

			ctx.Logger.Info("Certificate Issuance Completed",
				"request_id", cert.ID,
				"k8s_name", cert.K8sCertificateName,
				"namespace", cert.K8sNamespace)
		}
	}

	return nil
}
