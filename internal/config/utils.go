package config

import (
	"fmt"
	"net/url"
)

func validateURL(urlStr, fieldName string) error {
	if urlStr == "" {
		return fmt.Errorf("OIDC %s is required", fieldName)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("OIDC %s is not a valid URL: %w", fieldName, err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("OIDC %s must have http or https scheme", fieldName)
	}

	return nil
}
