package handlers

import (
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/models"
	"net/http"
)

type ConfigResponse struct {
	MTLS     MTLSConfigResponse     `json:"mtls,omitempty"`
	Firewall FirewallConfigResponse `json:"firewall,omitempty"`
}

type MTLSConfigResponse struct {
	Enabled            bool                        `json:"enabled"`
	CertificateSubject *CertificateSubjectResponse `json:"certificate_subject,omitempty"`
	KeyAlgorithm       string                      `json:"key_algorithm,omitempty"`
	ProviderType       string                      `json:"provider_type,omitempty"`
}

type CertificateSubjectResponse struct {
	Organization string `json:"organization,omitempty"`
	Country      string `json:"country,omitempty"`
	Locality     string `json:"locality,omitempty"`
	Province     string `json:"province,omitempty"`
}

type FirewallConfigResponse struct {
	Enabled bool `json:"enabled"`
}

type AuthStatusResponse struct {
	Authenticated bool            `json:"authenticated"`
	User          *models.User    `json:"user,omitempty"`
	Config        *ConfigResponse `json:"config,omitempty"`
}

func GETAuthStatusHandler(ctx *middlewares.AppContext) {
	response := AuthStatusResponse{
		Authenticated: false,
	}

	if !ctx.SessionManager.IsUserAuthenticated(ctx) {
		ctx.WriteJSON(http.StatusUnauthorized, response)
		return
	}

	config := &ConfigResponse{}

	if ctx.Config.Features != nil && ctx.Config.Features.MTLSManagement.Enabled {
		mtlsConfig := MTLSConfigResponse{
			Enabled: true,
		}

		if ctx.Config.Features.MTLSManagement.CertificateSubject != nil {
			mtlsConfig.CertificateSubject = &CertificateSubjectResponse{
				Organization: ctx.Config.Features.MTLSManagement.CertificateSubject.Organization,
				Country:      ctx.Config.Features.MTLSManagement.CertificateSubject.Country,
				Locality:     ctx.Config.Features.MTLSManagement.CertificateSubject.Locality,
				Province:     ctx.Config.Features.MTLSManagement.CertificateSubject.Province,
			}
		}

		if ctx.Config.Features.MTLSManagement.Kubernetes != nil && ctx.Config.Features.MTLSManagement.Kubernetes.Enabled {
			mtlsConfig.ProviderType = "kubernetes"
		} else if ctx.Config.Features.MTLSManagement.Database != nil && ctx.Config.Features.MTLSManagement.Database.Enabled {
			mtlsConfig.ProviderType = "database"
			mtlsConfig.KeyAlgorithm = ctx.Config.Features.MTLSManagement.Database.KeyAlgorithm
		}

		config.MTLS = mtlsConfig
	}

	if ctx.Config.Features != nil && ctx.Config.Features.FirewallManagement.Enabled {
		config.Firewall = FirewallConfigResponse{
			Enabled: true,
		}
	}

	response.Config = config

	user, ok := ctx.SessionManager.GetAuthenticatedUser(ctx)
	if user == nil {
		ctx.WriteJSON(http.StatusUnauthorized, response)
		return
	}

	if ok {
		response.Authenticated = true
		response.User = user
		ctx.WriteJSON(http.StatusOK, response)
		return
	}

	ctx.WriteJSON(http.StatusUnauthorized, response)
}
