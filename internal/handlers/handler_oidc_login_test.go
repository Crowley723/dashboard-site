package handlers

import (
	"errors"
	"homelab-dashboard/internal/testutil"
	"log/slog"
	"net/http"
	"testing"
)

const (
	expectedRedirectURL   = "https://auth.example.com/api/authorize?state=12345"
	defaultRedirectPath   = "/"
	nonDefaultRedirectURL = "https://example.com/abc123"
)

func TestGetLoginHandler_ShouldReturnRedirectURL(t *testing.T) {
	tc := testutil.NewTestContextWithURL(t, "GET", "/api/auth/login")
	defer tc.Finish()

	tc.MockSession.EXPECT().IsAuthenticated(tc.AppContext).Return(false).Times(1)

	tc.MockSession.EXPECT().SetRedirectAfterLogin(tc.AppContext, defaultRedirectPath).Return().Times(1)

	tc.MockOidcProvider.EXPECT().StartLogin(tc.AppContext).Return(expectedRedirectURL, nil).Times(1)

	tc.CallHandler(GETLoginHandler)

	tc.AssertStatus(t, http.StatusOK)
	tc.AssertContentType(t, "application/json")
	tc.AssertJSONField(t, "status", "redirect_required")
	tc.AssertJSONField(t, "redirect_url", expectedRedirectURL)
	tc.AssertLogsContainMessage(t, slog.LevelInfo, "Redirecting to OIDC Provider")
}

func TestGetLoginHandler_ShouldRedirectAlreadyAuthenticatedUser(t *testing.T) {
	tc := testutil.NewTestContextWithURL(t, "GET", "/api/auth/login")
	defer tc.Finish()

	tc.MockSession.EXPECT().IsAuthenticated(tc.AppContext).Return(true).Times(1)

	tc.CallHandler(GETLoginHandler)

	tc.AssertStatus(t, http.StatusOK)
	tc.AssertContentType(t, "application/json")
	tc.AssertJSONField(t, "status", "ok")
	tc.AssertLogsContainMessage(t, slog.LevelInfo, "User already authenticated")

}

func TestGetLoginHandler_ShouldRedirectWithReferrer(t *testing.T) {
	tc := testutil.NewTestContextWithURL(t, "GET", "/api/auth/login")
	tc.Request.Header.Set("Referer", nonDefaultRedirectURL)
	defer tc.Finish()

	tc.MockSession.EXPECT().IsAuthenticated(tc.AppContext).Return(false).Times(1)

	tc.MockSession.EXPECT().SetRedirectAfterLogin(tc.AppContext, nonDefaultRedirectURL).Return().Times(1)

	tc.MockOidcProvider.EXPECT().StartLogin(tc.AppContext).Return(expectedRedirectURL, nil).Times(1)

	tc.CallHandler(GETLoginHandler)

	tc.AssertStatus(t, http.StatusOK)
	tc.AssertContentType(t, "application/json")
	tc.AssertJSONField(t, "status", "redirect_required")
	tc.AssertJSONField(t, "redirect_url", expectedRedirectURL)
	tc.AssertLogsContainMessage(t, slog.LevelInfo, "Redirecting to OIDC Provider")
}

func TestGetLoginHandler_ShouldErrorOnStartLoginError(t *testing.T) {
	tc := testutil.NewTestContextWithURL(t, "GET", "/api/auth/login")
	defer tc.Finish()

	tc.MockSession.EXPECT().IsAuthenticated(tc.AppContext).Return(false).Times(1)

	tc.MockSession.EXPECT().SetRedirectAfterLogin(tc.AppContext, defaultRedirectPath).Return().Times(1)

	tc.MockOidcProvider.EXPECT().StartLogin(tc.AppContext).Return(expectedRedirectURL, errors.New("failed")).Times(1)

	tc.CallHandler(GETLoginHandler)

	tc.AssertStatus(t, http.StatusInternalServerError)
	tc.AssertContentType(t, "application/json")
	tc.AssertJSONField(t, "error", "Internal Server Error")
	tc.AssertLogsContainMessage(t, slog.LevelError, "Failed to start login")
}
