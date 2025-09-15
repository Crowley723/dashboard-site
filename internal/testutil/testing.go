// internal/testutil/testutil.go
package testutil

import (
	"encoding/json"
	"homelab-dashboard/internal/config"
	"homelab-dashboard/internal/data"
	"homelab-dashboard/internal/middlewares"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestContext holds everything needed for testing
type TestContext struct {
	AppContext *middlewares.AppContext
	Request    *http.Request
	Response   *httptest.ResponseRecorder
}

// NewTestContext creates a complete test setup with sensible defaults
func NewTestContext(method, url string) *TestContext {
	// Create minimal dependencies
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &config.Config{} // Add any default test config values here
	cache := &data.Cache{}

	// Create request and response
	req := httptest.NewRequest(method, url, nil)
	rr := httptest.NewRecorder()

	// Create AppContext
	appCtx := &middlewares.AppContext{
		Context:        req.Context(),
		Config:         cfg,
		Logger:         logger,
		SessionManager: nil, // Mock as needed
		OIDCProvider:   nil, // Mock as needed
		OauthConfig:    nil, // Mock as needed
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

// AssertJSONField checks a specific field in a JSON response
func (tc *TestContext) AssertJSONField(t *testing.T, field, expected string) {
	response := tc.GetJSONResponse(t)
	if actual, ok := response[field].(string); !ok || actual != expected {
		t.Errorf("Expected %s to be %s, got %v", field, expected, response[field])
	}
}

// GetResponseBody returns the response body as a string
func (tc *TestContext) GetResponseBody() string {
	return tc.Response.Body.String()
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

// WithCache allows you to override the default cache for specific tests
func (tc *TestContext) WithCache(cache *data.Cache) *TestContext {
	tc.AppContext.Cache = cache
	return tc
}
