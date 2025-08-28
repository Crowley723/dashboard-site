package session

import (
	"encoding/gob"
	"fmt"
	"homelab-dashboard/auth"
	"homelab-dashboard/config"
	"homelab-dashboard/middlewares"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/coreos/go-oidc/v3/oidc"
)

type SessionManager struct {
	*scs.SessionManager
}

func NewSessionManager(cfg config.SessionConfig) (*SessionManager, error) {
	gob.Register(&auth.User{})
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

	return &SessionManager{SessionManager: sessionManager}, nil
}

func (s *SessionManager) SetUser(ctx *middlewares.AppContext, user *auth.User) {
	s.Put(ctx, string(SessionKeyUserData), user)
}

func (s *SessionManager) GetUser(ctx *middlewares.AppContext) (user *auth.User, ok bool) {
	data := s.Get(ctx, string(SessionKeyUserData))
	if data == nil {
		return nil, false
	}

	if user, ok := data.(*auth.User); ok {
		return user, true
	}

	return nil, false
}

func (s *SessionManager) SetAuthenticated(ctx *middlewares.AppContext, authenticated bool) {
	s.Put(ctx, string(SessionKeyAuthenticated), authenticated)
}

func (s *SessionManager) IsAuthenticated(ctx *middlewares.AppContext) bool {
	return s.GetBool(ctx, string(SessionKeyAuthenticated))
}

func (s *SessionManager) SetTokenExpiry(ctx *middlewares.AppContext, expiry time.Time) {
	s.Put(ctx, string(SessionKeyTokenExpiry), expiry.Unix())
}

func (s *SessionManager) GetTokenExpiry(ctx *middlewares.AppContext) (time.Time, bool) {
	timestamp := s.GetInt64(ctx, string(SessionKeyTokenExpiry))
	if timestamp == 0 {
		return time.Time{}, false
	}
	return time.Unix(timestamp, 0), true
}

func (s *SessionManager) SetCreatedAt(ctx *middlewares.AppContext, createdAt time.Time) {
	s.Put(ctx, string(SessionKeyCreatedAt), createdAt.Unix())
}

func (s *SessionManager) GetCreatedAt(ctx *middlewares.AppContext) (time.Time, bool) {
	timestamp := s.GetInt64(ctx, string(SessionKeyCreatedAt))
	if timestamp == 0 {
		return time.Time{}, false
	}
	return time.Unix(timestamp, 0), true
}

func (s *SessionManager) SetRedirectAfterLogin(ctx *middlewares.AppContext, redirectAfterLogin string) {
	s.Put(ctx, string(SessionKeyRedirectAfterLogin), redirectAfterLogin)
}

func (s *SessionManager) GetRedirectAfterLogin(ctx *middlewares.AppContext) string {
	return s.GetString(ctx, string(SessionKeyRedirectAfterLogin))
}

func (s *SessionManager) SetExpiresAt(ctx *middlewares.AppContext, expiresAt time.Time) {
	s.Put(ctx, string(SessionKeyExpiresAt), expiresAt.Unix())
}

func (s *SessionManager) GetExpiresAt(ctx *middlewares.AppContext) (time.Time, bool) {
	timestamp := s.GetInt64(ctx, string(SessionKeyExpiresAt))
	if timestamp == 0 {
		return time.Time{}, false
	}
	return time.Unix(timestamp, 0), true
}

func (s *SessionManager) CreateSessionWithTokenExpiry(ctx *middlewares.AppContext, idToken *oidc.IDToken, user *auth.User) error {
	now := time.Now()
	tokenExpiry := idToken.Expiry
	sessionDuration := tokenExpiry.Sub(now)

	if sessionDuration <= 0 {
		return fmt.Errorf("token already expired")
	}

	s.SetUser(ctx, user)
	s.SetAuthenticated(ctx, true)
	s.SetTokenExpiry(ctx, tokenExpiry)
	s.SetCreatedAt(ctx, now)
	s.SetExpiresAt(ctx, tokenExpiry)

	return nil
}

func (s *SessionManager) IsSessionValid(ctx *middlewares.AppContext) bool {
	expiresAt, exists := s.GetExpiresAt(ctx)
	if !exists {
		return false
	}

	return time.Now().Before(expiresAt)
}

func (s *SessionManager) IsUserAuthenticated(ctx *middlewares.AppContext) bool {
	if !s.IsAuthenticated(ctx) {
		return false
	}

	expiresAt, exists := s.GetExpiresAt(ctx)
	if exists && !time.Now().Before(expiresAt) {
		return false
	}

	return true
}

func (s *SessionManager) GetCurrentUser(ctx *middlewares.AppContext) (*auth.User, bool) {
	if !s.IsUserAuthenticated(ctx) {
		return nil, false
	}

	return s.GetUser(ctx)
}

func (s *SessionManager) Logout(ctx *middlewares.AppContext) error {
	return s.Destroy(ctx.Request.Context())
}
