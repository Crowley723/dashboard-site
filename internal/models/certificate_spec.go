package models

// CertificateSpec contains the details for creating a Certificate resource
type CertificateSpec struct {
	Name                string
	Namespace           string
	CommonName          string
	DNSNames            []string
	OrganizationalUnits []string
	ValidityDays        int
	OwnerSub            string
	OwnerIss            string
	RequestID           int
}
