package jobs

import (
	"errors"
)

var (
	errNoApprovedCertificates = errors.New("no approved certificate")
	errNoIssuedCertificates   = errors.New("no issued certificate")
)
