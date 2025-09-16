package middlewares

import (
	"homelab-dashboard/internal/models"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

//go:generate mockgen -source=oidc_provider.go -destination=../mocks/oidc.go -package=mocks

type OIDCProvider interface {
	GenerateState() (string, error)
	StartLogin(ctx *AppContext) (string, error)
	HandleCallback(ctx *AppContext) (*models.User, error)
	GetProvider() *oidc.Provider
	GetOAuth2Config() *oauth2.Config
}
