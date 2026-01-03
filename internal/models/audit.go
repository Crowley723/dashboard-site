package models

import (
	"net/netip"
	"time"
)

type CertificateDownload struct {
	ID                   int
	CertificateRequestID int
	Sub                  string
	Iss                  string
	IPAddress            netip.Addr
	UserAgent            string
	BrowserName          string
	BrowserVersion       string
	OSName               string
	OSVersion            string
	DeviceType           string
	DownloadedAt         time.Time
}
