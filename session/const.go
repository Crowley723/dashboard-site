package session

import "homelab-dashboard/auth"

var (
	SessionKeyUserData           auth.SessionKey = "user_data"
	SessionKeyAuthenticated      auth.SessionKey = "authenticated"
	SessionKeyTokenExpiry        auth.SessionKey = "token_expiry"
	SessionKeyCreatedAt          auth.SessionKey = "created_at"
	SessionKeyExpiresAt          auth.SessionKey = "expires_at"
	SessionKeyRedirectAfterLogin auth.SessionKey = "redirect_after_login"
)
