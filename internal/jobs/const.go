package jobs

import (
	"errors"
)

var (
	errNoApprovedCertificates = errors.New("no approved certificates")
	errNoIssuedCertificates   = errors.New("no issued certificates")
)
