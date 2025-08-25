package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"homelab-dashboard/config"
	"homelab-dashboard/middlewares"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

func NewOIDCProvider(ctx context.Context, cfg config.OIDCConfig) (*oidc.Provider, *oauth2.Config, error) {
	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		Scopes:       cfg.Scopes,
		RedirectURL:  cfg.RedirectURI,
	}

	return provider, oauth2Config, err
}

func GenerateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func StartLogin(ctx *middlewares.AppContext) (string, error) {
	state, err := GenerateState()
	if err != nil {
		return "", err
	}

	ctx.Logger.Info("Storing OAuth state in session", "state", state)
	ctx.SessionManager.Put(ctx.Request.Context(), "oauth_state", state)

	retrievedState := ctx.SessionManager.GetString(ctx.Request.Context(), "oauth_state")
	ctx.Logger.Info("Retrieved OAuth state immediately", "state", retrievedState, "matches", state == retrievedState)

	authURL := ctx.OauthConfig.AuthCodeURL(state)
	return authURL, nil
}

func HandleCallback(ctx *middlewares.AppContext) (*User, error) {
	storedState := ctx.SessionManager.GetString(ctx.Request.Context(), "oauth_state")
	if storedState == "" {
		return nil, fmt.Errorf("no oauth state found in session")
	}

	receivedState := ctx.Request.URL.Query().Get("state")
	if receivedState != storedState {
		return nil, fmt.Errorf("invalid state parameter")
	}

	ctx.SessionManager.Remove(ctx.Request.Context(), "oauth_state")

	code := ctx.Request.URL.Query().Get("code")
	if code == "" {
		return nil, fmt.Errorf("no authorization code received")
	}

	token, err := ctx.OauthConfig.Exchange(ctx.Request.Context(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token found in oauth2 token")
	}

	verifier := ctx.OIDCProvider.Verifier(&oidc.Config{ClientID: ctx.OauthConfig.ClientID})

	idToken, err := verifier.Verify(ctx.Request.Context(), rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID Token: %w", err)
	}

	user, err := extractUserClaimsFromToken(idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to extract user from ID Token: %w", err)
	}

	enhancedUser, err := fetchUserInfo(ctx, token, user)
	if err != nil {
		ctx.Logger.Warn("Failed to fetch user info, using ID token data only", "error", err)
		enhancedUser = user
	}

	err = CreateSessionWithTokenExpiry(ctx, idToken, enhancedUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user session: %w", err)
	}

	return enhancedUser, nil
}

// fetchUserInfo retrieves additional user information from the UserInfo endpoint
func fetchUserInfo(ctx *middlewares.AppContext, token *oauth2.Token, baseUser *User) (*User, error) {
	userInfo, err := ctx.OIDCProvider.UserInfo(context.Background(), oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	enhancedUser := &User{
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
