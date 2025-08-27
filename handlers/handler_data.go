package handlers

import (
	"homelab-dashboard/data"
	"homelab-dashboard/middlewares"
	"slices"

	"github.com/go-jose/go-jose/v4/json"
)

func GetMetricsGET(ctx *middlewares.AppContext) {
	queries := ctx.Request.URL.Query().Get("queries")
	var userGroups []string

	var data []ResultData
	if queries == "" {
		cacheNames := ctx.Cache.ListAll()
		data = make([]ResultData, 0, len(cacheNames))
		for _, entryName := range cacheNames {
			entry, exists := ctx.Cache.Get(entryName)
			if !exists {
				continue
			}

			canAccess := false
			if entry.RequireGroup {
				canAccess = slices.Contains(userGroups, entry.RequiredGroup)
			}

			if canAccess {
				data = append(data, *convertCachedDataToResultData(&entry))
			}

		}
	}

	//queryList := strings.Split(queries, ",")
}

func convertCachedDataToResultData(data *data.CachedData) *ResultData {
	return &ResultData{
		QueryName: data.Name,
		Type:      data.Value.Type().String(),
		Data:      json.RawMessage(data.Value.String()),
		Timestamp: data.Timestamp.Unix(),
	}
}
