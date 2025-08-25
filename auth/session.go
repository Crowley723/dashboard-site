package auth

import (
	"encoding/gob"
	"fmt"
	"homelab-dashboard/config"
	"homelab-dashboard/middlewares"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/coreos/go-oidc/v3/oidc"
)

func NewSessionManager(cfg config.SessionConfig) (*scs.SessionManager, error) {
	gob.Register(&User{})
	sessionManager := scs.New()

	switch cfg.Store {
	case "memory":
		sessionManager.Store = memstore.New()
	default:
		return nil, fmt.Errorf("unsupported session store: %s", cfg.Store)
	}

	sessionManager.Lifetime = cfg.FixedTimeout

	sessionManager.Cookie.Name = cfg.Name
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = cfg.Secure
	sessionManager.Cookie.Path = "/"

	return sessionManager, nil
}

func CreateSessionWithTokenExpiry(ctx *middlewares.AppContext, idToken *oidc.IDToken, user *User) error {
	now := time.Now()
	tokenExpiry := idToken.Expiry
	sessionDuration := tokenExpiry.Sub(now)

	if sessionDuration <= 0 {
		return fmt.Errorf("token already expired")
	}

	ctx.SessionManager.Put(ctx.Request.Context(), "user_data", user)
	ctx.SessionManager.Put(ctx.Request.Context(), "authenticated", true)
	ctx.SessionManager.Put(ctx.Request.Context(), "token_expiry", tokenExpiry.Unix())
	ctx.SessionManager.Put(ctx.Request.Context(), "created_at", now.Unix())

	ctx.SessionManager.Put(ctx.Request.Context(), "expires_at", tokenExpiry.Unix())

	return nil
}

func IsSessionValid(sessionManager *scs.SessionManager, r *http.Request) bool {
	expiresAt := sessionManager.Get(r.Context(), "expires_at").(int64)
	if expiresAt == 0 {
		return false
	}

	return time.Now().Unix() < expiresAt
}

func GetCurrentUser(ctx *middlewares.AppContext) (*User, bool) {
	if !IsAuthenticated(ctx) {
		return nil, false
	}

	exists := ctx.SessionManager.Get(ctx.Request.Context(), "user_data")
	if exists == nil {
		return nil, false
	}

	if user, ok := exists.(*User); ok {
		return user, true
	}

	return nil, false
}

// IsAuthenticated checks if the user is authenticated and token is still valid
func IsAuthenticated(ctx *middlewares.AppContext) bool {
	authenticated := ctx.SessionManager.GetBool(ctx.Request.Context(), "authenticated")
	if !authenticated {
		return false
	}

	// Check if token is still valid
	expiresAt := ctx.SessionManager.GetInt64(ctx.Request.Context(), "token_expires_at")
	if expiresAt > 0 && time.Now().Unix() >= expiresAt {
		return false // token expired
	}

	return true
}

// Logout user
func Logout(ctx *middlewares.AppContext) error {
	return ctx.SessionManager.Destroy(ctx.Request.Context())
}
