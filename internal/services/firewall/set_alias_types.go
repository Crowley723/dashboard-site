package firewall

// AliasSetRequest represents the POST body for setItem
type AliasSetRequest struct {
	Alias            AliasSetBody `json:"alias"`
	NetworkContent   string       `json:"network_content"`
	AuthGroupContent string       `json:"authgroup_content"`
}

// AliasSetBody is the simplified structure for POST
type AliasSetBody struct {
	Enabled        string `json:"enabled"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Proto          string `json:"proto"`
	Categories     string `json:"categories"`
	UpdateFreq     string `json:"updatefreq"`
	Content        string `json:"content"` // Newline-separated values
	PathExpression string `json:"path_expression"`
	AuthType       string `json:"authtype"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	Interface      string `json:"interface"`
	Counters       string `json:"counters"`
	Description    string `json:"description"`
}
