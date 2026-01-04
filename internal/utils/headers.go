package utils

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrMissingAuthzHeader     = errors.New("missing authorization header")
	ErrInvalidAuthzHeader     = errors.New("invalid authorization header")
	ErrUnsupportedAuthzScheme = errors.New("unsupported authorization scheme")
	ErrMissingAuthzToken      = errors.New("missing authorization token")
)

func ExtractAuthorizationHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrMissingAuthzHeader
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return "", ErrInvalidAuthzHeader
	}

	scheme := parts[0]
	token := parts[1]

	if !strings.EqualFold(scheme, "Bearer") {
		return "", fmt.Errorf("%v: %s", ErrUnsupportedAuthzScheme, scheme)
	}

	if token == "" {
		return "", ErrMissingAuthzToken
	}

	return token, nil
}
