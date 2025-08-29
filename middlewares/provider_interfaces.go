package middlewares

import (
	"homelab-dashboard/models"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
)

type SessionProvider interface {
	SetUser(ctx *AppContext, user *models.User)
	GetUser(ctx *AppContext) (user *models.User, ok bool)
	SetAuthenticated(ctx *AppContext, authenticated bool)
	IsAuthenticated(ctx *AppContext) bool
	SetTokenExpiry(ctx *AppContext, expiry time.Time)
	GetTokenExpiry(ctx *AppContext) (time.Time, bool)
	SetCreatedAt(ctx *AppContext, createdAt time.Time)
	GetCreatedAt(ctx *AppContext) (time.Time, bool)
	SetRedirectAfterLogin(ctx *AppContext, redirectAfterLogin string)
	GetRedirectAfterLogin(ctx *AppContext) string
	SetOauthState(ctx *AppContext, state string)
	GetOauthState(ctx *AppContext) string
	ClearOauthState(ctx *AppContext)
	SetExpiresAt(ctx *AppContext, expiresAt time.Time)
	GetExpiresAt(ctx *AppContext) (time.Time, bool)
	CreateSessionWithTokenExpiry(ctx *AppContext, idToken *oidc.IDToken, user *models.User) error
	IsSessionValid(ctx *AppContext) bool
	IsUserAuthenticated(ctx *AppContext) bool
	GetCurrentUser(ctx *AppContext) (user *models.User, ok bool)
	Logout(ctx *AppContext) error

	LoadAndSave(next http.Handler) http.Handler
}
