// internal/handlers/handler_data_test.go
package handlers

import (
	"homelab-dashboard/internal/models"
	"homelab-dashboard/internal/testutil"
	"strings"
	"testing"
)

func TestGetMetricsGET(t *testing.T) {
	tests := []struct {
		name           string
		queries        []string
		setupMocks     func(tc *testutil.TestContext)
		expectedStatus int
		expectedCount  int
		validate       func(t *testing.T, results []interface{})
	}{
		{
			name:           "SinglePublicMetricsShouldReturnMetric",
			queries:        []string{"cpu_usage"},
			expectedStatus: 200,
			expectedCount:  1,
			setupMocks: func(tc *testutil.TestContext) {
				cachedData := tc.CreateCachedDataWithScalar("cpu_usage", 85.5, false, "")

				tc.MockSession.EXPECT().
					GetCurrentUser(tc.AppContext).
					Return(nil, false).
					Times(1)

				tc.MockCache.EXPECT().
					Get("cpu_usage").
					Return(cachedData, true).
					Times(1)
			},
			validate: func(t *testing.T, results []interface{}) {
				metric := results[0].(map[string]interface{})
				if metric["query_name"] != "cpu_usage" {
					t.Errorf("Expected query_name 'cpu_usage', got %v", metric["query_name"])
				}
				if metric["type"] != "scalar" {
					t.Errorf("Expected type 'scalar', got %v", metric["type"])
				}
			},
		},
		{
			name:           "SinglePrivateMetricShouldReturnMetricForAuthorizedUser",
			queries:        []string{"cpu_usage"},
			expectedStatus: 200,
			expectedCount:  1,
			setupMocks: func(tc *testutil.TestContext) {
				cachedData := tc.CreateCachedDataWithScalar("cpu_usage", 85.5, true, "admin")

				tc.MockSession.EXPECT().
					GetCurrentUser(tc.AppContext).
					Return(&models.User{Groups: []string{"admin"}}, true).
					Times(1)

				tc.MockCache.EXPECT().
					Get("cpu_usage").
					Return(cachedData, true).
					Times(1)
			},
			validate: func(t *testing.T, results []interface{}) {
				metric := results[0].(map[string]interface{})
				if metric["query_name"] != "cpu_usage" {
					t.Errorf("Expected query_name 'cpu_usage', got %v", metric["query_name"])
				}
				if metric["type"] != "scalar" {
					t.Errorf("Expected type 'scalar', got %v", metric["type"])
				}
			},
		},
		{
			name:           "SinglePrivateMetricShouldReturnEmptyMetricsForUnauthorizedUser",
			queries:        []string{"cpu_usage"},
			expectedStatus: 200,
			expectedCount:  0,
			setupMocks: func(tc *testutil.TestContext) {
				cachedData := tc.CreateCachedDataWithScalar("cpu_usage", 85.5, true, "admin")

				tc.MockSession.EXPECT().
					GetCurrentUser(tc.AppContext).
					Return(&models.User{Groups: []string{}}, true).
					Times(1)

				tc.MockCache.EXPECT().
					Get("cpu_usage").
					Return(cachedData, true).
					Times(1)
			},
			validate: func(t *testing.T, results []interface{}) {
				if len(results) != 0 {
					t.Errorf("Expected 0 results for unauthorized user, got %d", len(results))
				}
			},
		},
		{
			name:           "UnknownMetricShouldReturnNoMetrics",
			queries:        []string{"nonexistent_metric"},
			expectedStatus: 200,
			expectedCount:  0,
			setupMocks: func(tc *testutil.TestContext) {
				tc.MockSession.EXPECT().
					GetCurrentUser(tc.AppContext).
					Return(nil, false).
					Times(1)

				tc.MockCache.EXPECT().
					Get("nonexistent_metric").
					Return(tc.CreateCachedDataWithScalar("", 0, false, ""), false).
					Times(1)
			},
			validate: func(t *testing.T, results []interface{}) {
			},
		},
		{
			name:           "MultipleMetricsShouldReturnAllMetrics",
			queries:        []string{"cpu_usage,memory_usage"},
			expectedStatus: 200,
			expectedCount:  2,
			setupMocks: func(tc *testutil.TestContext) {
				cpuData := tc.CreateCachedDataWithScalar("cpu_usage", 85.5, false, "")
				memData := tc.CreateCachedDataWithScalar("memory_usage", 65.2, false, "")

				tc.MockSession.EXPECT().
					GetCurrentUser(tc.AppContext).
					Return(nil, false).
					Times(1)

				tc.MockCache.EXPECT().
					Get("cpu_usage").
					Return(cpuData, true).
					Times(1)

				tc.MockCache.EXPECT().
					Get("memory_usage").
					Return(memData, true).
					Times(1)
			},
			validate: func(t *testing.T, results []interface{}) {
				queryNames := make(map[string]bool)
				for _, result := range results {
					metric := result.(map[string]interface{})
					if name, ok := metric["query_name"].(string); ok {
						queryNames[name] = true
					}
				}

				if !queryNames["cpu_usage"] {
					t.Error("Expected to find cpu_usage in results")
				}
				if !queryNames["memory_usage"] {
					t.Error("Expected to find memory_usage in results")
				}
			},
		},
		{
			name:           "MixedKnownAndUnknownMetricsShouldReturnKnownMetrics",
			queries:        []string{"cpu_usage", "nonexistent_metric"},
			expectedStatus: 200,
			expectedCount:  1,
			setupMocks: func(tc *testutil.TestContext) {
				cpuData := tc.CreateCachedDataWithScalar("cpu_usage", 85.5, false, "")

				tc.MockSession.EXPECT().
					GetCurrentUser(tc.AppContext).
					Return(nil, false).
					Times(1)

				tc.MockCache.EXPECT().
					Get("cpu_usage").
					Return(cpuData, true).
					Times(1)

				tc.MockCache.EXPECT().
					Get("nonexistent_metric").
					Return(tc.CreateCachedDataWithScalar("", 0, false, ""), false).
					Times(1)
			},
			validate: func(t *testing.T, results []interface{}) {
				metric := results[0].(map[string]interface{})
				if metric["query_name"] != "cpu_usage" {
					t.Errorf("Expected query_name 'cpu_usage', got %v", metric["query_name"])
				}
			},
		},
		{
			name:           "NoMetricsShouldReturnAllAvailableMetrics",
			expectedStatus: 200,
			expectedCount:  1,
			setupMocks: func(tc *testutil.TestContext) {
				cpuData := tc.CreateCachedDataWithScalar("cpu_usage", 85.5, false, "")

				tc.MockSession.EXPECT().
					GetCurrentUser(tc.AppContext).
					Return(nil, false).
					Times(1)

				tc.MockCache.EXPECT().
					ListAll().
					Return([]string{"cpu_usage"}).
					Times(1)

				tc.MockCache.EXPECT().
					Get("cpu_usage").
					Return(cpuData, true).
					Times(1)
			},
			validate: func(t *testing.T, results []interface{}) {
				metric := results[0].(map[string]interface{})
				if metric["query_name"] != "cpu_usage" {
					t.Errorf("Expected query_name 'cpu_usage', got %v", metric["query_name"])
				}
			},
		},
		{
			name:           "UnmarshalableMetricShouldBeSkippedAndContinue",
			queries:        []string{"invalid_metric", "cpu_usage"},
			expectedStatus: 200,
			expectedCount:  1,
			setupMocks: func(tc *testutil.TestContext) {
				invalidData := tc.CreateCachedDataWithUnmarshalableValue("invalid_metrics", false, "")
				cpuData := tc.CreateCachedDataWithScalar("cpu_usage", 85.5, false, "")

				tc.MockSession.EXPECT().
					GetCurrentUser(tc.AppContext).
					Return(nil, false).
					Times(1)

				tc.MockCache.EXPECT().
					Get("invalid_metric").
					Return(invalidData, true).
					Times(1)

				tc.MockCache.EXPECT().
					Get("cpu_usage").
					Return(cpuData, true).
					Times(1)
			},
			validate: func(t *testing.T, results []interface{}) {
				if len(results) != 1 {
					t.Errorf("Expected 1 result (invalid metric should be skipped), got %d", len(results))
				}

				if len(results) > 0 {
					metric := results[0].(map[string]interface{})
					if metric["query_name"] != "cpu_usage" {
						t.Errorf("Expected query_name 'cpu_usage', got %v", metric["query_name"])
					}
				}
			},
		},
		{
			name:           "UnmarshalableMetricWithAuthRequiredShouldBeSkippedAndContinue",
			queries:        []string{"invalid_metric", "cpu_usage"},
			expectedStatus: 200,
			expectedCount:  1,
			setupMocks: func(tc *testutil.TestContext) {
				invalidData := tc.CreateCachedDataWithUnmarshalableValue("invalid_metrics", true, "admin")
				cpuData := tc.CreateCachedDataWithScalar("cpu_usage", 85.5, false, "")

				tc.MockSession.EXPECT().
					GetCurrentUser(tc.AppContext).
					Return(&models.User{Groups: []string{"admin"}}, true).
					Times(1)

				tc.MockCache.EXPECT().
					Get("invalid_metric").
					Return(invalidData, true).
					Times(1)

				tc.MockCache.EXPECT().
					Get("cpu_usage").
					Return(cpuData, true).
					Times(1)
			},
			validate: func(t *testing.T, results []interface{}) {
				if len(results) != 1 {
					t.Errorf("Expected 1 result (invalid metric should be skipped), got %d", len(results))
				}

				if len(results) > 0 {
					metric := results[0].(map[string]interface{})
					if metric["query_name"] != "cpu_usage" {
						t.Errorf("Expected query_name 'cpu_usage', got %v", metric["query_name"])
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := testutil.NewTestContextWithURL(t, "GET", "/api/metrics")
			defer tc.Finish()

			if len(tt.queries) > 1 {
				tc = tc.WithQueryParam("queries", strings.Join(tt.queries, ","))
			} else if len(tt.queries) == 1 {
				tc = tc.WithQueryParam("queries", tt.queries[0])
			}

			tt.setupMocks(tc)

			tc.CallHandler(GetMetricsGET)

			tc.AssertStatus(t, tt.expectedStatus)
			tc.AssertContentType(t, "application/json")

			results := tc.GetJSONResponseArray(t)

			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d metrics, got %d", tt.expectedCount, len(results))
			}

			tt.validate(t, results)
		})
	}
}
