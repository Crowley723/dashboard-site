package utils

import (
	v1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
)

func IsCertificateReady(cert *v1.Certificate) bool {
	for _, condition := range cert.Status.Conditions {
		if condition.Type == v1.CertificateConditionReady {
			return condition.Status == cmmeta.ConditionTrue
		}
	}

	return false
}
