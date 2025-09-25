package handlers

import (
	"strings"
)

func RedactEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}

	localPart := parts[0]
	domain := parts[1]

	if len(localPart) <= 2 {
		return strings.Repeat("*", len(localPart)) + "@" + domain
	}

	first := string(localPart[0])
	last := string(localPart[len(localPart)-1])
	middle := strings.Repeat("*", len(localPart)-2)

	return first + middle + last + "@" + domain
}
