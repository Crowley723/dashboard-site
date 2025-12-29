package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/models"
	"homelab-dashboard/internal/storage"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/avct/uasurfer"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"software.sslmate.com/src/go-pkcs12"
)

type CertificateUnlockBody struct {
	Passphrase string `json:"passphrase"`
}

type CertificateUnlockResponse struct {
	Unlocked      bool   `json:"unlocked"`
	Error         string `json:"error,omitempty"`
	DownloadToken string `json:"download_token,omitempty"`
	ExpiresIn     int    `json:"expires_in,omitempty"`
}

type DownloadToken struct {
	CertificateId int       `json:"certificate_id"`
	UserIss       string    `json:"user_iss"`
	UserSub       string    `json:"user_sub"`
	Passphrase    string    `json:"passphrase"`
	CreatedAt     time.Time `json:"created_at"`
}

var downloadTokenLifetime = 5 * time.Minute
var downloadTokenKeyFmt = "download_token:%s"

// POSTCertificateUnlock is the first step in downloading a certificate. The user posts a passphrase they want the p12 file encrypted with and receive a one-time download token
func POSTCertificateUnlock(ctx *middlewares.AppContext) {
	certificateIdParam := chi.URLParam(ctx.Request, "id")
	if certificateIdParam == "" {
		ctx.SetJSONError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	certificateId, err := strconv.Atoi(strings.TrimSpace(certificateIdParam))
	if err != nil {
		ctx.SetJSONError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	user, ok := ctx.SessionManager.GetAuthenticatedUser(ctx)
	if !ok || user == nil {
		ctx.SetJSONError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	request, err := ctx.Storage.Certificates().GetRequestByID(ctx, certificateId)
	if err != nil {
		if errors.Is(err, storage.CertificateRequestNotFoundError) {
			ctx.SetJSONError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
			return
		}

		ctx.SetJSONError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if !user.MatchesUser(request.OwnerIss, request.OwnerSub) {
		ctx.SetJSONError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	var reqBody CertificateUnlockBody
	if err := json.NewDecoder(ctx.Request.Body).Decode(&reqBody); err != nil {
		ctx.Logger.Error("failed to decode request body", "error", err)
		ctx.SetJSONError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	if reqBody.Passphrase == "" || validatePassphraseComplexity(reqBody.Passphrase) != nil {
		resp := CertificateUnlockResponse{
			Unlocked: false,
			Error:    fmt.Sprintf("passphrase must be at least 12 characters"),
		}
		ctx.WriteJSON(http.StatusBadRequest, resp)
		return
	}

	tokenUUID := uuid.New().String()
	tokenHash := hashToken(tokenUUID, []byte(ctx.Config.Features.MTLSManagement.DownloadTokenHMACKey))
	tokenKey := fmt.Sprintf(downloadTokenKeyFmt, tokenHash)

	downloadToken := DownloadToken{
		CertificateId: request.ID,
		UserIss:       request.OwnerIss,
		UserSub:       request.OwnerSub,
		Passphrase:    reqBody.Passphrase,
		CreatedAt:     time.Now(),
	}

	tokenData, _ := json.Marshal(downloadToken)

	if err := ctx.Cache.SetKey(ctx, tokenKey, tokenData, downloadTokenLifetime); err != nil {
		ctx.Logger.Error("unable to store download token", "error", err)
		ctx.SetJSONError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	resp := CertificateUnlockResponse{
		Unlocked:      true,
		DownloadToken: tokenUUID,
		ExpiresIn:     int(downloadTokenLifetime.Seconds()),
	}

	ctx.WriteJSON(http.StatusOK, resp)

}

func GETCertificateDownload(ctx *middlewares.AppContext) {
	certificateIdParam := chi.URLParam(ctx.Request, "id")
	if certificateIdParam == "" {
		ctx.SetJSONError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	certificateId, err := strconv.Atoi(strings.TrimSpace(certificateIdParam))
	if err != nil {
		ctx.SetJSONError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	downloadTokenParam := strings.TrimSpace(ctx.Request.URL.Query().Get("token"))
	if downloadTokenParam == "" {
		ctx.SetJSONError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	user, ok := ctx.SessionManager.GetAuthenticatedUser(ctx)
	if !ok || user == nil {
		ctx.SetJSONError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	tokenHash := hashToken(downloadTokenParam, []byte(ctx.Config.Features.MTLSManagement.DownloadTokenHMACKey))

	downloadTokenJson, err := ctx.Cache.GetDelKey(ctx, fmt.Sprintf(downloadTokenKeyFmt, tokenHash))
	if err != nil {
		ctx.Logger.Error("unable to fetch download token", "error", err)
		ctx.SetJSONError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	var downloadToken DownloadToken
	err = json.Unmarshal([]byte(downloadTokenJson), &downloadToken)
	if err != nil {
		ctx.Logger.Error("failed to unmarshal download token", "error", err)
		ctx.SetJSONError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if !user.MatchesUser(downloadToken.UserIss, downloadToken.UserSub) {
		ctx.SetJSONError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	if certificateId != downloadToken.CertificateId {
		ctx.SetJSONError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	request, err := ctx.Storage.Certificates().GetRequestByID(ctx, certificateId)
	if err != nil {
		if errors.Is(err, storage.CertificateRequestNotFoundError) {
			ctx.SetJSONError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
			return
		}

		ctx.Logger.Warn("failed to fetch certificate request", "error", err)
		ctx.SetJSONError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if request.Status != models.StatusIssued {
		ctx.SetJSONError(http.StatusBadRequest, "Certificate is not issued yet")
		return
	}

	if !user.MatchesUser(request.OwnerIss, request.OwnerSub) {
		ctx.SetJSONError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	certPEM, keyPEM, caPEM, err := ctx.KubernetesClient.GetCertificateForDownload(
		ctx,
		*request.K8sNamespace,
		*request.K8sCertificateName,
	)

	if err != nil {
		ctx.Logger.Error("failed to get certificate from k8s",
			"error", err,
			"certName", *request.K8sCertificateName)
		ctx.SetJSONError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	host, _, err := net.SplitHostPort(ctx.Request.RemoteAddr)
	if err != nil {
		ctx.SetJSONError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	_, err = ctx.Storage.Audit().LogDownload(ctx, certificateId, user.Sub, user.Iss, host, ctx.Request.UserAgent(), *uasurfer.Parse(ctx.Request.UserAgent()))
	if err != nil {
		ctx.Logger.Error("failed to insert download audit log", "error", err)
		ctx.SetJSONError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	ctx.Logger.Info("retrieved certificate from k8s",
		"certName", *request.K8sCertificateName,
		"certPEMLen", len(certPEM),
		"keyPEMLen", len(keyPEM),
		"caPEMLen", len(caPEM))

	p12Bytes, err := GenerateP12(certPEM, keyPEM, caPEM, downloadToken.Passphrase)
	if err != nil {
		ctx.Logger.Error("failed to generate P12", "error", err)
		ctx.SetJSONError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	ctx.Logger.Info("generated P12 file",
		"p12Size", len(p12Bytes),
		"certificateId", certificateId)

	filename := fmt.Sprintf("certificate-%d.p12", certificateId)

	ctx.Response.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	ctx.Response.Header().Set("Content-Type", "application/x-pkcs12")
	ctx.Response.Header().Set("Content-Length", strconv.Itoa(len(p12Bytes)))

	_, err = ctx.Response.Write(p12Bytes)
	if err != nil {
		ctx.Logger.Error("failed to write certificate p12", "error", err)
		return
	}
}

// GenerateP12 creates a PKCS12 bundle from PEM-encoded cert and key
func GenerateP12(certPEM, keyPEM, caPEM []byte, passphrase string) ([]byte, error) {
	cert, err := parseCertificate(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	privateKey, err := parsePrivateKey(keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	var caCerts []*x509.Certificate
	if len(caPEM) > 0 {
		caCerts, err = parseCACertificates(caPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CA certificates: %w", err)
		}
	}

	p12Data, err := pkcs12.Modern.Encode(privateKey, cert, caCerts, passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to encode PKCS12: %w", err)
	}

	return p12Data, nil
}

// parseCertificate parses a single PEM-encoded certificate
func parseCertificate(certPEM []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}
	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("expected CERTIFICATE block, got %s", block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// parsePrivateKey parses a PEM-encoded private key
func parsePrivateKey(keyPEM []byte) (interface{}, error) {
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key PEM")
	}

	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, fmt.Errorf("failed to parse private key: unsupported format")
}

// parseCACertificates parses multiple PEM-encoded CA certificates
func parseCACertificates(caPEM []byte) ([]*x509.Certificate, error) {
	var caCerts []*x509.Certificate

	rest := caPEM
	for len(rest) > 0 {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
		}

		caCerts = append(caCerts, cert)
	}

	return caCerts, nil
}

func validatePassphraseComplexity(passphrase string) error {
	if len(passphrase) < 12 {
		return errors.New("passphrase must be at least 12 characters")
	}
	return nil
}

// hashToken creates an HMAC-SHA256 hash
func hashToken(token string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

func verifyTokenHash(token string, hash string, secret []byte) bool {
	expectedHash := hashToken(token, secret)
	return hmac.Equal([]byte(expectedHash), []byte(hash))
}
