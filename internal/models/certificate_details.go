package models

import (
	"time"
)

type CertificateDetails struct {
	SerialNumber string
	Subject      string
	Issuer       string
	NotBefore    time.Time
	NotAfter     time.Time
	DNSNames     []string
	CommonName   string
	Organization []string
}
