package firewall

import "strings"

type AliasGetResponse struct {
	Alias AliasDetail `json:"alias"`
}

// AliasDetail contains all the details of an alias
type AliasDetail struct {
	Enabled        string                  `json:"enabled"`
	Name           string                  `json:"name"`
	Type           map[string]SelectOption `json:"type"`
	PathExpression string                  `json:"path_expression"`
	Proto          map[string]SelectOption `json:"proto"`
	Interface      map[string]SelectOption `json:"interface"`
	Counters       string                  `json:"counters"`
	UpdateFreq     string                  `json:"updatefreq"`
	Content        map[string]ContentItem  `json:"content"`
	Password       string                  `json:"password"`
	Username       string                  `json:"username"`
	AuthType       map[string]SelectOption `json:"authtype"`
	Categories     map[string]SelectOption `json:"categories"`
	CurrentItems   string                  `json:"current_items"`
	LastUpdated    string                  `json:"last_updated"`
	Description    string                  `json:"description"`
}

// SelectOption represents a selectable option with value and selected state
type SelectOption struct {
	Value       string `json:"value"`
	Selected    int    `json:"selected"`
	Description string `json:"description,omitempty"`
}

// ContentItem represents an IP address or alias reference
type ContentItem struct {
	Value       string `json:"value"`
	Selected    int    `json:"selected"`
	Description string `json:"description,omitempty"`
}

// GetSelectedIPs extracts selected IPs from content
func (a *AliasDetail) GetSelectedIPs() []string {
	var ips []string
	for key, item := range a.Content {
		if item.Selected == 1 && !isInternalAlias(key) && !isAliasReference(key) {
			// Strip CIDR notation if present (e.g., "192.168.1.1/32" -> "192.168.1.1")
			ip := item.Value
			if strings.Contains(ip, "/") {
				ip = strings.Split(ip, "/")[0]
			}
			ips = append(ips, ip)
		}
	}
	return ips
}

// isInternalAlias checks if a key is an internal alias
func isInternalAlias(key string) bool {
	internalPrefixes := []string{"__", "bogons", "virusprot", "sshlockout"}
	for _, prefix := range internalPrefixes {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// Helper to check if it's an alias reference (not a direct IP)
func isAliasReference(key string) bool {
	// Simple heuristic: if it doesn't contain dots, colons, or dashes at the start
	// it's probably an alias name
	// IP addresses and ranges will have these characters
	hasIPChars := false
	for _, char := range key {
		if char == '.' || char == ':' || (char == '-' && key[0] != '-') {
			hasIPChars = true
			break
		}
	}
	return !hasIPChars
}

// Helper function to get the selected value from a map of SelectOptions
func getSelected(options map[string]SelectOption) string {
	for key, opt := range options {
		if opt.Selected == 1 {
			return key
		}
	}
	return ""
}
