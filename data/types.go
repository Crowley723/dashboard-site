package data

import (
	"time"

	"github.com/prometheus/common/model"
)

type QueryResult struct {
	Name      string      `json:"name"`
	Value     model.Value `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
}

type DashboardData struct {
	LastUpdated time.Time     `json:"last_updated"`
	Queries     []QueryResult `json:"queries"`
}
