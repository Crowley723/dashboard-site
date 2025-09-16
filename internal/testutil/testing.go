package testutil

import (
	"context"
	"encoding/json"
	"homelab-dashboard/internal/config"
	"homelab-dashboard/internal/data"
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/mocks"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"go.uber.org/mock/gomock"
)

// TestContext holds everything needed for testing
type TestContext struct {
	AppContext     *middlewares.AppContext
	Request        *http.Request
	Response       *httptest.ResponseRecorder
	MockController *gomock.Controller
	MockCache      *mocks.MockCacheProvider
	MockSession    *mocks.MockSessionProvider
	LogHandler     *TestLogHandler
}

func NewTestContext(t *testing.T) *TestContext {
	cfg := &config.Config{}

	logHandler := NewTestLogHandler()
	logger := slog.New(logHandler)

	// Create mock controller
	ctrl := gomock.NewController(t)

	// Create mocks
	mockCache := mocks.NewMockCacheProvider(ctrl)
	mockSession := mocks.NewMockSessionProvider(ctrl)

	rr := httptest.NewRecorder()

	appCtx := &middlewares.AppContext{
		Context:        context.Background(),
		Config:         cfg,
		Logger:         logger,
		SessionManager: mockSession,
		OIDCProvider:   nil,
		OauthConfig:    nil,
		Cache:          mockCache,
		Request:        nil,
		Response:       rr,
	}

	return &TestContext{
		AppContext:     appCtx,
		Request:        nil,
		Response:       rr,
		MockController: ctrl,
		MockCache:      mockCache,
		MockSession:    mockSession,
	}
}

// NewTestContextWithURL creates a complete test setup with sensible defaults
func NewTestContextWithURL(t *testing.T, method, url string) *TestContext {
	cfg := &config.Config{}

	logHandler := NewTestLogHandler()
	logger := slog.New(logHandler)

	// Create mock controller
	ctrl := gomock.NewController(t)

	// Create mocks
	mockCache := mocks.NewMockCacheProvider(ctrl)
	mockSession := mocks.NewMockSessionProvider(ctrl)

	req := httptest.NewRequest(method, url, nil)
	rr := httptest.NewRecorder()

	appCtx := &middlewares.AppContext{
		Context:        req.Context(),
		Config:         cfg,
		Logger:         logger,
		SessionManager: mockSession,
		OIDCProvider:   nil,
		OauthConfig:    nil,
		Cache:          mockCache,
		Request:        req,
		Response:       rr,
	}

	return &TestContext{
		AppContext:     appCtx,
		Request:        req,
		Response:       rr,
		MockController: ctrl,
		MockCache:      mockCache,
		MockSession:    mockSession,
		LogHandler:     logHandler,
	}
}

// NewTestContextWithRealCache creates a test context with a real MemCache instead of mock
func NewTestContextWithRealCache(method, url string) *TestContext {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &config.Config{}
	cache := &data.MemCache{}

	req := httptest.NewRequest(method, url, nil)
	rr := httptest.NewRecorder()

	appCtx := &middlewares.AppContext{
		Context:        req.Context(),
		Config:         cfg,
		Logger:         logger,
		SessionManager: nil,
		OIDCProvider:   nil,
		OauthConfig:    nil,
		Cache:          cache,
		Request:        req,
		Response:       rr,
	}

	return &TestContext{
		AppContext: appCtx,
		Request:    req,
		Response:   rr,
	}
}

// Finish should be called at the end of tests to clean up mocks
func (tc *TestContext) Finish() {
	if tc.MockController != nil {
		tc.MockController.Finish()
	}
}

func (tc *TestContext) AssertLogContains(t *testing.T, level slog.Level, message string) {
	if !tc.LogHandler.ContainsMessage(level, message) {
		t.Errorf("Expected to find log entry with level %v containing message: %s", level, message)
	}
}

func (tc *TestContext) AssertLogCount(t *testing.T, level slog.Level, expectedCount int) {
	count := tc.LogHandler.CountByLevel(level)
	if count != expectedCount {
		t.Errorf("Expected %d log entries at level %v, got %d", expectedCount, level, count)
	}
}

func (tc *TestContext) GetLogRecords() []TestLogRecord {
	return tc.LogHandler.GetRecords()
}

func (tc *TestContext) ClearLogRecords() {
	tc.LogHandler.Reset()
}

// CallHandler executes a handler with the test context
func (tc *TestContext) CallHandler(handler middlewares.AppHandler) {
	handler(tc.AppContext)
}

// AssertStatus checks the HTTP status code
func (tc *TestContext) AssertStatus(t *testing.T, expectedStatus int) {
	if tc.Response.Code != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, tc.Response.Code)
	}
}

// AssertContentType checks the content type header
func (tc *TestContext) AssertContentType(t *testing.T, expectedType string) {
	if ct := tc.Response.Header().Get("Content-Type"); ct != expectedType {
		t.Errorf("Expected content type %s, got %s", expectedType, ct)
	}
}

// GetJSONResponse parses the response body as JSON
func (tc *TestContext) GetJSONResponse(t *testing.T) map[string]interface{} {
	var response map[string]interface{}
	if err := json.Unmarshal(tc.Response.Body.Bytes(), &response); err != nil {
		t.Fatalf("Could not parse JSON response: %v", err)
	}
	return response
}

// GetJSONResponseArray parses the response body as a JSON array
func (tc *TestContext) GetJSONResponseArray(t *testing.T) []interface{} {
	var response []interface{}
	if err := json.Unmarshal(tc.Response.Body.Bytes(), &response); err != nil {
		t.Fatalf("Could not parse JSON array response: %v", err)
	}
	return response
}

// AssertJSONField checks a specific field in a JSON response
func (tc *TestContext) AssertJSONField(t *testing.T, field string, expected any) {
	response := tc.GetJSONResponse(t)
	if actual, ok := response[field]; !ok || actual != expected {
		t.Errorf("Expected %s to be %s, got %v", field, expected, response[field])
	}
}

func (tc *TestContext) AssertJSONBool(t *testing.T, field string, expected bool) {
	response := tc.GetJSONResponse(t)
	actual, exists := response[field]

	if !exists {
		t.Errorf("Field %s not found in response", field)
		return
	}

	actualBool, ok := actual.(bool)
	if !ok {
		t.Errorf("Expected %s to be a boolean, got %T", field, actual)
		return
	}

	if actualBool != expected {
		t.Errorf("Expected %s to be %v, got %v", field, expected, actualBool)
	}
}

// AssertJSONString checks a specific string field in a JSON response
func (tc *TestContext) AssertJSONString(t *testing.T, field string, expected string) {
	response := tc.GetJSONResponse(t)
	actual, exists := response[field]

	if !exists {
		t.Errorf("Field %s not found in response", field)
		return
	}

	actualString, ok := actual.(string)
	if !ok {
		t.Errorf("Expected %s to be a string, got %T", field, actual)
		return
	}

	if actualString != expected {
		t.Errorf("Expected %s to be %q, got %q", field, expected, actualString)
	}
}

// AssertJSONObject validates an object field with expected key-value pairs
func (tc *TestContext) AssertJSONObject(t *testing.T, field string, expectedFields map[string]interface{}) {
	response := tc.GetJSONResponse(t)
	actual, exists := response[field]

	if !exists {
		t.Errorf("Field %s not found in response", field)
		return
	}

	actualObj, ok := actual.(map[string]interface{})
	if !ok {
		t.Errorf("Expected %s to be an object, got %T", field, actual)
		return
	}

	for key, expectedValue := range expectedFields {
		if actualValue, keyExists := actualObj[key]; !keyExists {
			t.Errorf("Expected field %s.%s to exist", field, key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected %s.%s to be %v, got %v", field, key, expectedValue, actualValue)
		}
	}
}

// AssertUser validates a user object in the JSON response
func (tc *TestContext) AssertUser(t *testing.T, field string, expectedUser interface{}) {
	response := tc.GetJSONResponse(t)
	actual, exists := response[field]

	if !exists {
		t.Errorf("Field %s not found in response", field)
		return
	}

	user, ok := actual.(map[string]interface{})
	if !ok {
		t.Errorf("Expected %s to be a user object, got %T", field, actual)
		return
	}

	// Handle different user types - you'll need to import your models package
	switch u := expectedUser.(type) {
	case map[string]interface{}:
		// Compare as key-value pairs
		for key, expectedValue := range u {
			if actualValue, keyExists := user[key]; !keyExists {
				t.Errorf("Expected field %s.%s to exist", field, key)
			} else if actualValue != expectedValue {
				t.Errorf("Expected %s.%s to be %v, got %v", field, key, expectedValue, actualValue)
			}
		}
	default:
		// For any struct type, convert to map for comparison
		userBytes, err := json.Marshal(expectedUser)
		if err != nil {
			t.Errorf("Failed to marshal expected user: %v", err)
			return
		}

		var expectedUserMap map[string]interface{}
		if err := json.Unmarshal(userBytes, &expectedUserMap); err != nil {
			t.Errorf("Failed to unmarshal expected user: %v", err)
			return
		}

		// Compare only non-empty/non-nil fields from expected user
		for key, expectedValue := range expectedUserMap {
			// Skip nil values and empty strings unless they're explicitly set
			if expectedValue == nil || expectedValue == "" {
				continue
			}

			if actualValue, keyExists := user[key]; !keyExists {
				t.Errorf("Expected field %s.%s to exist", field, key)
			} else if actualValue != expectedValue {
				t.Errorf("Expected %s.%s to be %v, got %v", field, key, expectedValue, actualValue)
			}
		}
	}
}

// WithConfig allows you to override the default config for specific tests
func (tc *TestContext) WithConfig(cfg *config.Config) *TestContext {
	tc.AppContext.Config = cfg
	return tc
}

// WithLogger allows you to override the default logger for specific tests
func (tc *TestContext) WithLogger(logger *slog.Logger) *TestContext {
	tc.AppContext.Logger = logger
	return tc
}

// WithCache allows you to override the cache with a different mock or implementation
func (tc *TestContext) WithCache(cache data.CacheProvider) *TestContext {
	tc.AppContext.Cache = cache
	return tc
}

// WithSessionManager allows you to override the session manager with a different mock or implementation
func (tc *TestContext) WithSessionManager(sm middlewares.SessionProvider) *TestContext {
	tc.AppContext.SessionManager = sm
	return tc
}

// Helper to add query parameters to the request
func (tc *TestContext) WithQueryParam(key, value string) *TestContext {
	q := tc.Request.URL.Query()
	q.Add(key, value)
	tc.Request.URL.RawQuery = q.Encode()
	return tc
}

// Helper to add headers
func (tc *TestContext) WithHeader(key, value string) *TestContext {
	tc.Request.Header.Set(key, value)
	return tc
}

// Assertion helpers for common patterns
func (tc *TestContext) AssertJSONArrayLength(t *testing.T, expected int) {
	response := tc.GetJSONResponseArray(t)
	if len(response) != expected {
		t.Errorf("Expected JSON array length %d, got %d", expected, len(response))
	}
}

// WithRequest allows you to set a custom request (useful for tests that don't use URL constructor)
func (tc *TestContext) WithRequest(req *http.Request) *TestContext {
	tc.Request = req
	tc.AppContext.Request = req
	tc.AppContext.Context = req.Context()
	return tc
}

// ExpectCacheGet sets up an expectation for cache.Get()
func (tc *TestContext) ExpectCacheGet(queryName string, returnData data.CachedData, found bool) *gomock.Call {
	return tc.MockCache.EXPECT().Get(queryName).Return(returnData, found)
}

// ExpectCacheSet sets up an expectation for cache.Set()
func (tc *TestContext) ExpectCacheSet(queryName string, value interface{}, requireAuth bool, requiredGroup string) *gomock.Call {
	return tc.MockCache.EXPECT().Set(queryName, value, requireAuth, requiredGroup)
}

// ExpectSessionIsAuthenticated sets up an expectation for session.IsAuthenticated()
func (tc *TestContext) ExpectSessionIsAuthenticated(result bool) *gomock.Call {
	return tc.MockSession.EXPECT().IsAuthenticated(tc.AppContext).Return(result)
}

// ExpectSessionGetUser sets up an expectation for session.GetUser()
func (tc *TestContext) ExpectSessionGetUser(user interface{}, ok bool) *gomock.Call {
	return tc.MockSession.EXPECT().GetUser(tc.AppContext).Return(user, ok)
}

func (tc *TestContext) CreateCachedDataWithScalar(name string, value float64, requireAuth bool, requiredGroup string) data.CachedData {
	scalar := &model.Scalar{
		Value:     model.SampleValue(value),
		Timestamp: model.Time(time.Now().Unix() * 1000), // Prometheus uses milliseconds
	}
	return data.CachedData{
		Name:          name,
		Value:         scalar,
		Timestamp:     time.Now(),
		RequireAuth:   requireAuth,
		RequiredGroup: requiredGroup,
	}
}

// CreateCachedDataWithVector creates a CachedData instance with a Vector value
func (tc *TestContext) CreateCachedDataWithVector(name string, samples []*model.Sample, requireAuth bool, requiredGroup string) data.CachedData {
	vector := model.Vector(samples)
	return data.CachedData{
		Name:          name,
		Value:         vector,
		Timestamp:     time.Now(),
		RequireAuth:   requireAuth,
		RequiredGroup: requiredGroup,
	}
}

// CreateSample creates a model.Sample for use in vectors
func (tc *TestContext) CreateSample(labels model.LabelSet, value float64, timestamp time.Time) *model.Sample {
	return &model.Sample{
		Metric:    model.Metric(labels),
		Value:     model.SampleValue(value),
		Timestamp: model.Time(timestamp.Unix() * 1000),
	}
}

type UnmarshalableValue struct {
	Channel chan int
}

func (u UnmarshalableValue) Type() model.ValueType {
	return model.ValScalar
}

func (u UnmarshalableValue) String() string {
	return "unmarshalable"
}

func (tc *TestContext) CreateCachedDataWithUnmarshalableValue(name string, requireAuth bool, requiredGroup string) data.CachedData {
	unmarshalableValue := UnmarshalableValue{
		Channel: make(chan int),
	}

	return data.CachedData{
		Name:          name,
		Value:         unmarshalableValue,
		RequireAuth:   requireAuth,
		RequiredGroup: requiredGroup,
		Timestamp:     time.Now(),
	}
}
