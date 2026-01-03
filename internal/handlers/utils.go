package handlers

import (
	"fmt"
	"homelab-dashboard/internal/models"
	"strings"
)

// RedactEmail is used to redact emails (mostly for logs)
func RedactEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}

	localRunes := []rune(parts[0])
	domain := parts[1]

	if len(localRunes) <= 2 {
		return strings.Repeat("*", len(localRunes)) + "@" + domain
	}

	first := string(localRunes[0])
	last := string(localRunes[len(localRunes)-1])
	middle := strings.Repeat("*", len(localRunes)-2)

	return first + middle + last + "@" + domain
}

// deriveCommonName gets a CN for the user's mTLS certificate
func deriveCommonName(user *models.User) string {
	if user.DisplayName != "" {
		return user.DisplayName
	}

	if user.Username != "" {
		return user.Username
	}

	if user.Email != "" {
		return user.Email
	}

	issuerDomain := strings.TrimPrefix(user.Iss, "https://")
	issuerDomain = strings.TrimPrefix(issuerDomain, "http://")
	return fmt.Sprintf("%s@%s", user.Sub, issuerDomain)
}

func deriveOrganizationalUnits(user *models.User) []string {
	if len(user.Groups) == 0 {
		return []string{"Users"}
	}
	return user.Groups
}
