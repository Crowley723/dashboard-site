package handlers

import (
	"encoding/json"
)

// ResultData represents
type ResultData struct {
	QueryName     string          `json:"query_name"`
	Type          string          `json:"type"`
	Data          json.RawMessage `json:"data,omitempty"`
	Timestamp     int64           `json:"timestamp"`
	RequireAuth   bool            `json:"-"`
	RequiredGroup string          `json:"-"`
}
