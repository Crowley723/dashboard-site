package auth

var (
	SessionKeyUserData           SessionKey = "user_data"
	SessionKeyAuthenticated      SessionKey = "authenticated"
	SessionKeyTokenExpiry        SessionKey = "token_expiry"
	SessionKeyCreatedAt          SessionKey = "created_at"
	SessionKeyExpiresAt          SessionKey = "expires_at"
	SessionKeyRedirectAfterLogin SessionKey = "redirect_after_login"
	SessionKeyOauthState         SessionKey = "oauth_state"
)
