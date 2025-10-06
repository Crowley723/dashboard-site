package handlers

import (
	"homelab-dashboard/internal/data"
	"homelab-dashboard/internal/middlewares"
	"net/http"
	"slices"
	"strings"
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
		resultData = addRecordsIfAuthorized(ctx, ctx.Cache.ListAll(ctx.Context), userGroups)
	} else {
		resultData = addRecordsIfAuthorized(ctx, queries, userGroups)

	}

	ctx.WriteJSON(http.StatusOK, resultData)
}

func convertCachedDataToResultData(data *data.CachedData) (*ResultData, error) {
	return &ResultData{
		QueryName: data.Name,
		Type:      data.Value.Type().String(),
		Data:      data.JSONBytes,
	}, nil
}

func addRecordsIfAuthorized(ctx *middlewares.AppContext, queryNames []string, userGroups []string) []ResultData {
	var resultData []ResultData
	resultData = make([]ResultData, 0, len(queryNames))

	for _, entryName := range queryNames {
		entry, exists := ctx.Cache.Get(ctx.Context, entryName)
		if !exists {
			continue
		}

		if entry.RequireAuth {
			if canAccess := slices.Contains(userGroups, entry.RequiredGroup); canAccess {
				dataRecord, err := convertCachedDataToResultData(&entry)
				if err != nil {
					ctx.Logger.Error("failed to add cached data to result", "error", err)
					continue
				}

				resultData = append(resultData, *dataRecord)
			}
		} else {
			dataRecord, err := convertCachedDataToResultData(&entry)
			if err != nil {
				ctx.Logger.Error("failed to add cached data to result", "error", err)
				continue
			}

			resultData = append(resultData, *dataRecord)
		}
	}

	return resultData
}

func GetQueriesGET(ctx *middlewares.AppContext) {
	queryParam := ctx.Request.URL.Query().Get("queries")
	queries := strings.Split(queryParam, ",")
	user, userExists := ctx.SessionManager.GetCurrentUser(ctx)

	var userGroups []string
	if userExists {
		userGroups = user.Groups
	}

	var resultData []ResultData
	if queryParam == "" {
		resultData = addRecordsIfAuthorized(ctx, ctx.Cache.ListAll(ctx.Context), userGroups)
	} else {
		resultData = addRecordsIfAuthorized(ctx, queries, userGroups)

	}

	dataNames := make([]string, 0, len(resultData))
	for _, query := range resultData {
		dataNames = append(dataNames, query.QueryName)
	}

	ctx.WriteJSON(http.StatusOK, dataNames)
}
