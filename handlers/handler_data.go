package handlers

import (
	"homelab-dashboard/data"
	"homelab-dashboard/middlewares"
	"net/http"
	"slices"
	"strings"

	"github.com/go-jose/go-jose/v4/json"
)

func GetMetricsGET(ctx *middlewares.AppContext) {
	queryParam := ctx.Request.URL.Query().Get("queries")
	queries := strings.Split(queryParam, ",")
	user, userExists := ctx.SessionManager.GetCurrentUser(ctx)

	var userGroups []string
	if userExists {
		userGroups = user.Groups
	}

	var resultData []ResultData
	if queryParam == "" {
		resultData = addRecordsIfAuthorized(ctx, ctx.Cache.ListAll(), userGroups)
	} else {
		resultData = addRecordsIfAuthorized(ctx, queries, userGroups)

	}

	ctx.WriteJSON(http.StatusOK, resultData)
}

func convertCachedDataToResultData(data *data.CachedData) (*ResultData, error) {
	jsonBytes, err := json.Marshal(data.Value)

	return &ResultData{
		QueryName: data.Name,
		Type:      data.Value.Type().String(),
		Data:      jsonBytes,
		Timestamp: data.Timestamp.Unix(),
	}, err
}

func addRecordsIfAuthorized(ctx *middlewares.AppContext, queryNames []string, userGroups []string) []ResultData {
	var resultData []ResultData
	resultData = make([]ResultData, 0, len(queryNames))

	for _, entryName := range queryNames {
		entry, exists := ctx.Cache.Get(entryName)
		if !exists {
			continue
		}

		if entry.RequireAuth {
			if canAccess := slices.Contains(userGroups, entry.RequiredGroup); canAccess {
				dataRecord, err := convertCachedDataToResultData(&entry)
				if err != nil {
					ctx.Logger.Error("failed to add cached data to result", err)
					continue
				}

				resultData = append(resultData, *dataRecord)
			}
		} else {
			dataRecord, err := convertCachedDataToResultData(&entry)
			if err != nil {
				ctx.Logger.Error("failed to add cached data to result", err)
				continue
			}

			resultData = append(resultData, *dataRecord)
		}
	}

	return resultData
}
