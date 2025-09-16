package handlers

import (
	"fmt"
	"homelab-dashboard/internal/models"
	"homelab-dashboard/internal/testutil"
	"log/slog"
	"net/http"
	"testing"
)

var (
	failedAuth = "auth_failed"
)

func TestGetCallbackHandler_ShouldRedirectOnSuccess(t *testing.T) {
	tc := testutil.NewTestContextWithURL(t, "GET", "/api/auth/callback")
	tc.Request.URL.Query().Set("error", "")
	defer tc.Finish()

	testUser := &models.User{
		Sub:      "sub_claim",
		Iss:      "iss_claim",
		Username: "steve",
		Email:    "steve@example.com",
		Groups:   []string{"admin", "dev"},
	}

	tc.MockOidcProvider.EXPECT().HandleCallback(tc.AppContext).Return(testUser, nil).Times(1)

	tc.MockSession.EXPECT().GetRedirectAfterLogin(tc.AppContext).Return("").Times(1)

	tc.CallHandler(GETCallbackHandler)

	tc.AssertStatus(t, http.StatusFound)
	tc.AssertContentType(t, "text/html; charset=utf-8")
	tc.AssertLogsContainMessage(t, slog.LevelInfo, "User successfully authenticated")
}

func TestGetCallbackHandler_ShouldErrorOnError(t *testing.T) {
	tc := testutil.NewTestContextWithURL(t, "GET", "/api/auth/callback")
	tc.WithQueryParam("error", failedAuth)
	tc.WithQueryParam("error_description", failedAuth)
	defer tc.Finish()

	tc.CallHandler(GETCallbackHandler)

	tc.AssertStatus(t, http.StatusFound)
	tc.AssertContentType(t, "text/html; charset=utf-8")
	tc.AssertLocationHeader(t, fmt.Sprintf("/callback?error=%s", failedAuth))
	tc.AssertLogsContainMessage(t, slog.LevelWarn, "OIDC callback error")
}

func TestGetCallbackHandler_ShouldErrorOnCallbackError(t *testing.T) {
	tc := testutil.NewTestContextWithURL(t, "GET", "/api/auth/callback")
	defer tc.Finish()

	tc.MockOidcProvider.EXPECT().HandleCallback(tc.AppContext).Return(nil, fmt.Errorf("error")).Times(1)

	tc.CallHandler(GETCallbackHandler)

	tc.AssertStatus(t, http.StatusFound)
	tc.AssertContentType(t, "text/html; charset=utf-8")
	tc.AssertLocationHeader(t, fmt.Sprintf("/callback?error=%s", failedAuth))
	tc.AssertLogsContainMessage(t, slog.LevelError, "Failed to handle OIDC callback")
}

func TestGetCallbackHandler_ShouldRedirectToPreAuthLocation(t *testing.T) {
	tc := testutil.NewTestContextWithURL(t, "GET", "/api/auth/callback")
	defer tc.Finish()

	testUser := &models.User{
		Sub:      "sub_claim",
		Iss:      "iss_claim",
		Username: "steve",
		Email:    "steve@example.com",
		Groups:   []string{"admin", "dev"},
	}

	tc.MockOidcProvider.EXPECT().HandleCallback(tc.AppContext).Return(testUser, nil).Times(1)

	tc.MockSession.EXPECT().GetRedirectAfterLogin(tc.AppContext).Return(nonDefaultRedirectURL).Times(1)

	tc.CallHandler(GETCallbackHandler)

	tc.AssertStatus(t, http.StatusFound)
	tc.AssertContentType(t, "text/html; charset=utf-8")
	tc.AssertLocationHeader(t, nonDefaultRedirectURL)
}
