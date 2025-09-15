// internal/handlers/handler_health_test.go
package handlers

import (
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/testutil"
	"testing"
)

func TestHandlerHealth(t *testing.T) {
	// Create test context
	tc := testutil.NewTestContext("GET", "/health")

	// Call the handler
	tc.CallHandler(HandlerHealth)

	// Assert the results
	tc.AssertStatus(t, 200)
	tc.AssertContentType(t, "application/json")
	tc.AssertJSONField(t, "status", "OK")
}

// Example: Testing error cases
func TestHandlerError(t *testing.T) {
	tc := testutil.NewTestContext("GET", "/error")

	// Create a handler that returns an error
	errorHandler := func(ctx *middlewares.AppContext) {
		ctx.SetJSONError(400, "Bad Request")
	}

	tc.CallHandler(errorHandler)

	tc.AssertStatus(t, 400)
	tc.AssertJSONField(t, "error", "Bad Request")
}

// Example: Custom assertions when you need more control
func TestHandlerCustom(t *testing.T) {
	tc := testutil.NewTestContext("GET", "/custom")

	tc.CallHandler(HandlerHealth)

	// Use the utility methods
	tc.AssertStatus(t, 200)

	// Or do custom assertions
	response := tc.GetJSONResponse(t)
	if len(response) != 1 {
		t.Errorf("Expected 1 field in response, got %d", len(response))
	}

	// Or check raw response
	body := tc.GetResponseBody()
	if body == "" {
		t.Error("Expected non-empty response body")
	}
}
