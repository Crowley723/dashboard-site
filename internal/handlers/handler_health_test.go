// internal/handlers/handler_health_test.go
package handlers

import (
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/testutil"
	"testing"
)

func TestHandlerHealth(t *testing.T) {
	tc := testutil.NewTestContext(t, "GET", "/health")

	tc.CallHandler(HandlerHealth)

	tc.AssertStatus(t, 200)
	tc.AssertContentType(t, "application/json")
	tc.AssertJSONField(t, "status", "OK")
}

func TestHandlerError(t *testing.T) {
	tc := testutil.NewTestContext(t, "GET", "/error")

	errorHandler := func(ctx *middlewares.AppContext) {
		ctx.SetJSONError(400, "Bad Request")
	}

	tc.CallHandler(errorHandler)

	tc.AssertStatus(t, 400)
	tc.AssertJSONField(t, "error", "Bad Request")
}

func TestHandlerCustom(t *testing.T) {
	tc := testutil.NewTestContext(t, "GET", "/custom")

	tc.CallHandler(HandlerHealth)

	tc.AssertStatus(t, 200)

	response := tc.GetJSONResponse(t)
	if len(response) != 1 {
		t.Errorf("Expected 1 field in response, got %d", len(response))
	}

	body := tc.GetResponseBody()
	if body == "" {
		t.Error("Expected non-empty response body")
	}
}
