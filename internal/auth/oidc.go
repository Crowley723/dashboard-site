package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"homelab-dashboard/internal/config"
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/models"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// NewRealOIDCProvider creates a new instance of the real OIDC provider with initialized config
func NewRealOIDCProvider(ctx context.Context, cfg config.OIDCConfig) (middlewares.OIDCProvider, error) {
	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		Scopes:       cfg.Scopes,
		RedirectURL:  cfg.RedirectURI,
	}

	return &RealOIDCProvider{
		provider:     provider,
		oauth2Config: oauth2Config,
	}, nil
}

type RealOIDCProvider struct {
	provider     *oidc.Provider
	oauth2Config *oauth2.Config
}

func (r *RealOIDCProvider) GetProvider() *oidc.Provider {
	return r.provider
}

func (r *RealOIDCProvider) GetOAuth2Config() *oauth2.Config {
	return r.oauth2Config
}

func (r *RealOIDCProvider) GenerateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func (r *RealOIDCProvider) StartLogin(ctx *middlewares.AppContext) (string, error) {
	state, err := r.GenerateState()
	if err != nil {
		return "", err
	}

	ctx.Logger.Info("Storing OAuth state in session")
	ctx.SessionManager.SetOauthState(ctx, state)

	authURL := ctx.OIDCProvider.GetOAuth2Config().AuthCodeURL(state,
		oauth2.SetAuthURLParam("prompt", "login"),
	)
	return authURL, nil
}

func (r *RealOIDCProvider) HandleCallback(ctx *middlewares.AppContext) (*models.User, error) {
	// Check for error parameters first
	if errorParam := ctx.Request.URL.Query().Get("error"); errorParam != "" {
		errorDescription := ctx.Request.URL.Query().Get("error_description")
		errorURI := ctx.Request.URL.Query().Get("error_uri")
		state := ctx.Request.URL.Query().Get("state")

		// Redirect to error page with parameters
		errorURL := fmt.Sprintf("/error?error=%s", url.QueryEscape(errorParam))
		if errorDescription != "" {
			errorURL += "&error_description=" + url.QueryEscape(errorDescription)
		}
		if errorURI != "" {
			errorURL += "&error_uri=" + url.QueryEscape(errorURI)
		}
		if state != "" {
			errorURL += "&state=" + url.QueryEscape(state)
		}

		// You'll need to handle this redirect in your handler
		return nil, &OIDCError{RedirectURL: errorURL, Message: errorParam}
	}

	storedState := ctx.SessionManager.GetOauthState(ctx)
	if storedState == "" {
		return nil, &OIDCError{
			RedirectURL: "/error?error=invalid_request&error_description=" + url.QueryEscape("No oauth state found in session"),
			Message:     "no oauth state found in session",
		}
	}

	receivedState := ctx.Request.URL.Query().Get("state")
	if receivedState != storedState {
		return nil, &OIDCError{
			RedirectURL: "/error?error=invalid_request&error_description=" + url.QueryEscape("Invalid state parameter"),
			Message:     "invalid state parameter",
		}
	}

	ctx.SessionManager.ClearOauthState(ctx)

	code := ctx.Request.URL.Query().Get("code")
	if code == "" {
		return nil, &OIDCError{
			RedirectURL: "/error?error=invalid_request&error_description=" + url.QueryEscape("No authorization code received"),
			Message:     "no authorization code received",
		}
	}

	token, err := ctx.OIDCProvider.GetOAuth2Config().Exchange(ctx.Request.Context(), code)
	if err != nil {
		return nil, &OIDCError{
			RedirectURL: "/error?error=invalid_grant&error_description=" + url.QueryEscape("Failed to exchange code for token"),
			Message:     fmt.Sprintf("failed to exchange code for token: %v", err),
		}
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, &OIDCError{
			RedirectURL: "/error?error=invalid_token&error_description=" + url.QueryEscape("No id_token found in oauth2 token"),
			Message:     "no id_token found in oauth2 token",
		}
	}

	verifier := ctx.OIDCProvider.GetProvider().Verifier(&oidc.Config{ClientID: ctx.OIDCProvider.GetOAuth2Config().ClientID})

	idToken, err := verifier.Verify(ctx.Request.Context(), rawIDToken)
	if err != nil {
		return nil, &OIDCError{
			RedirectURL: "/error?error=invalid_token&error_description=" + url.QueryEscape("Failed to verify ID Token"),
			Message:     fmt.Sprintf("failed to verify ID Token: %v", err),
		}
	}

	user, err := extractUserClaimsFromToken(idToken)
	if err != nil {
		return nil, &OIDCError{
			RedirectURL: "/error?error=server_error&error_description=" + url.QueryEscape("Failed to extract user from ID Token"),
			Message:     fmt.Sprintf("failed to extract user from ID Token: %v", err),
		}
	}

	enhancedUser, err := fetchUserInfo(ctx, token, user)
	if err != nil {
		ctx.Logger.Warn("Failed to fetch user info, using ID token data only", "error", err)
		enhancedUser = user
	}

	err = ctx.SessionManager.CreateSessionWithTokenExpiry(ctx, idToken, enhancedUser)
	if err != nil {
		return nil, &OIDCError{
			RedirectURL: "/error?error=server_error&error_description=" + url.QueryEscape("Failed to create user session"),
			Message:     fmt.Sprintf("failed to create user session: %v", err),
		}
	}

	return enhancedUser, nil
}

// fetchUserInfo retrieves additional user information from the UserInfo endpoint
func fetchUserInfo(ctx *middlewares.AppContext, token *oauth2.Token, baseUser *models.User) (*models.User, error) {
	userInfo, err := ctx.OIDCProvider.GetProvider().UserInfo(context.Background(), oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	enhancedUser := &models.User{
		Sub: baseUser.Sub,
		Iss: baseUser.Iss,
	}

	var claims struct {
		Username    string   `json:"preferred_username"`
		Name        string   `json:"name"`
		DisplayName string   `json:"display_name"`
		Email       string   `json:"email"`
		Groups      []string `json:"groups"`
	}

	// Extract claims from UserInfo
	if err := userInfo.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse user info claims: %w", err)
	}

	enhancedUser.Username = getPreferredValue(claims.Username, baseUser.Username)
	enhancedUser.DisplayName = getPreferredValue(claims.DisplayName, claims.Name, baseUser.DisplayName)
	enhancedUser.Email = getPreferredValue(claims.Email, baseUser.Email)
	enhancedUser.Groups = claims.Groups

	if len(enhancedUser.Groups) == 0 {
		enhancedUser.Groups = baseUser.Groups
	}

	return enhancedUser, nil
}

// getPreferredValue returns the first non-empty string from the provided values
func getPreferredValue(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
