package handlers

import (
	"homelab-dashboard/internal/models"
	"homelab-dashboard/internal/testutil"
	"net/http"
	"testing"
)

func TestAuthStatusHandler_ShouldReturnUnauthorizedForAnonymousUser(t *testing.T) {
	tc := testutil.NewTestContext(t)
	defer tc.Finish()

	tc.MockSession.EXPECT().IsUserAuthenticated(tc.AppContext).Return(false)

	tc.CallHandler(GETAuthStatusHandler)

	tc.AssertStatus(t, http.StatusUnauthorized)
	tc.AssertContentType(t, "application/json")
	tc.AssertJSONField(t, "authenticated", false)
}

func TestAuthStatusHandler_ShouldReturnAuthorizedForKnownUser(t *testing.T) {
	tc := testutil.NewTestContextWithURL(t, "GET", "/api/auth/status")
	defer tc.Finish()

	tc.MockSession.EXPECT().IsUserAuthenticated(tc.AppContext).Return(true)

	testUser := &models.User{
		Sub:      "sub_claim",
		Iss:      "iss_claim",
		Username: "steve",
	}

	tc.MockSession.EXPECT().GetCurrentUser(tc.AppContext).Return(testUser, true)

	tc.CallHandler(GETAuthStatusHandler)

	tc.AssertStatus(t, http.StatusOK)
	tc.AssertContentType(t, "application/json")
	tc.AssertJSONField(t, "authenticated", true)
	tc.AssertUser(t, "user", testUser)
}

func TestAuthStatusHandler_ShouldReturnUnauthorizedOnNotOK(t *testing.T) {
	tc := testutil.NewTestContextWithURL(t, "GET", "/api/auth/status")
	defer tc.Finish()

	tc.MockSession.EXPECT().IsUserAuthenticated(tc.AppContext).Return(true)

	testUser := &models.User{
		Sub:      "sub_claim",
		Iss:      "iss_claim",
		Username: "steve",
	}

	tc.MockSession.EXPECT().GetCurrentUser(tc.AppContext).Return(testUser, false)

	tc.CallHandler(GETAuthStatusHandler)

	tc.AssertStatus(t, http.StatusUnauthorized)
	tc.AssertContentType(t, "application/json")
	tc.AssertJSONField(t, "authenticated", false)
}

func TestAuthStatusHandler_ShouldReturnUnauthorizedOnNilUser(t *testing.T) {
	tc := testutil.NewTestContextWithURL(t, "GET", "/api/auth/status")
	defer tc.Finish()

	tc.MockSession.EXPECT().IsUserAuthenticated(tc.AppContext).Return(true)

	tc.MockSession.EXPECT().GetCurrentUser(tc.AppContext).Return(nil, true)

	tc.CallHandler(GETAuthStatusHandler)

	tc.AssertStatus(t, http.StatusUnauthorized)
	tc.AssertContentType(t, "application/json")
	tc.AssertJSONField(t, "authenticated", false)
}
