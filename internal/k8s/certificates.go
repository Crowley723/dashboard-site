package k8s

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"homelab-dashboard/internal/models"
	"strings"
	"time"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// labels for tracking certificates created by this system
const (
	LabelManagedBy   = "app.kubernetes.io/managed-by"
	LabelOwnerSub    = "conduit.homelab.dev/owner-sub"
	LabelOwnerIss    = "conduit.homelab.dev/owner-iss"
	LabelRequestID   = "conduit.homelab.dev/request-id"
	ManagedByConduit = "conduit"
)

// GenerateCertificateName generates a unique certificate name based on owner and timestamp
// Format: hash of "sub:iss:timestamp" truncated to fit Kubernetes naming constraints
func GenerateCertificateName(sub, iss string, timestamp time.Time) string {
	input := fmt.Sprintf("%s:%s:%d", sub, iss, timestamp.Unix())
	hash := sha256.Sum256([]byte(input))
	hashStr := hex.EncodeToString(hash[:])

	// Kubernetes DNS-1123 subdomain naming requirements:
	// - lowercase alphanumeric characters, '-' or '.'
	// - must start and end with alphanumeric
	// - max 253 characters
	// We'll use "cert-" prefix + first 32 chars of hash
	name := fmt.Sprintf("cert-%s", hashStr[:32])
	return strings.ToLower(name)
}

// CreateCertificateFromRequest creates a cert-manager Certificate resource from a CertificateRequest
func (c *Client) CreateCertificateFromRequest(ctx context.Context, request *models.CertificateRequest) (*certmanagerv1.Certificate, error) {
	// Generate certificate name
	certName := GenerateCertificateName(request.OwnerSub, request.OwnerIss, request.RequestedAt)
	secretName := fmt.Sprintf("%s-tls", certName)

	duration := time.Duration(request.ValidityDays) * 24 * time.Hour

	issuerRef := cmmeta.IssuerReference{
		Name: c.IssuerName,
		Kind: c.IssuerKind,
	}

	// If using Issuer (not ClusterIssuer), set the group
	if c.IssuerKind == "Issuer" {
		group := "cert-manager.io"
		issuerRef.Group = group
	}

	subject := &certmanagerv1.X509Subject{}
	if len(request.OrganizationalUnits) > 0 {
		subject.OrganizationalUnits = request.OrganizationalUnits
	}

	cert := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certName,
			Namespace: c.Namespace,
			Labels: map[string]string{
				LabelManagedBy: ManagedByConduit,
				LabelOwnerSub:  sanitizeLabelValue(request.OwnerSub),
				LabelOwnerIss:  sanitizeLabelValue(removeURLSchemeAndSlashes(request.OwnerIss)),
				LabelRequestID: fmt.Sprintf("%d", request.ID),
			},
		},
		Spec: certmanagerv1.CertificateSpec{
			SecretName: secretName,
			Duration: &metav1.Duration{
				Duration: duration,
			},
			IssuerRef:  issuerRef,
			CommonName: request.CommonName,
			DNSNames:   request.DNSNames,
			Subject:    subject,
			Usages: []certmanagerv1.KeyUsage{
				certmanagerv1.UsageDigitalSignature,
				certmanagerv1.UsageKeyEncipherment,
				certmanagerv1.UsageClientAuth,
			},
		},
	}

	c.Logger.Info("creating certificate",
		"name", certName,
		"namespace", c.Namespace,
		"commonName", request.CommonName,
		"requestID", request.ID)

	created, err := c.CertManagerClient.CertmanagerV1().Certificates(c.Namespace).Create(ctx, cert, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	c.Logger.Info("certificate created successfully",
		"name", created.Name,
		"namespace", created.Namespace)

	return created, nil
}

// GetCertificate retrieves a Certificate resource
func (c *Client) GetCertificate(ctx context.Context, namespace, name string) (*certmanagerv1.Certificate, error) {
	cert, err := c.CertManagerClient.CertmanagerV1().Certificates(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("certificate not found: %s/%s", namespace, name)
		}
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}
	return cert, nil
}

// GetCertificatePEM retrieves the PEM-encoded certificate data from the secret
func (c *Client) GetCertificatePEM(ctx context.Context, namespace, name string) (certPEM, keyPEM, caPEM []byte, err error) {
	cert, err := c.GetCertificate(ctx, namespace, name)
	if err != nil {
		return nil, nil, nil, err
	}

	secret, err := c.GetCertificateSecret(ctx, namespace, cert.Spec.SecretName)
	if err != nil {
		return nil, nil, nil, err
	}

	return secret.Data["tls.crt"], secret.Data["tls.key"], secret.Data["ca.crt"], nil
}

// IsCertificateReady checks if a Certificate is ready
func (c *Client) IsCertificateReady(ctx context.Context, namespace, name string) (bool, error) {
	cert, err := c.GetCertificate(ctx, namespace, name)
	if err != nil {
		return false, err
	}

	for _, condition := range cert.Status.Conditions {
		if condition.Type == certmanagerv1.CertificateConditionReady {
			return condition.Status == cmmeta.ConditionTrue, nil
		}
	}

	return false, nil
}

// GetCertificateSecret retrieves the Secret containing the issued certificate
func (c *Client) GetCertificateSecret(ctx context.Context, namespace, secretName string) (*corev1.Secret, error) {
	secret, err := c.ClientSet.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("certificate secret not found: %s/%s", namespace, secretName)
		}
		return nil, fmt.Errorf("failed to get certificate secret: %w", err)
	}
	return secret, nil
}

// DeleteCertificate deletes a Certificate resource
func (c *Client) DeleteCertificate(ctx context.Context, namespace, name string) error {
	c.Logger.Info("deleting certificate", "name", name, "namespace", namespace)

	err := c.CertManagerClient.CertmanagerV1().Certificates(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			c.Logger.Warn("certificate already deleted", "name", name, "namespace", namespace)
			return nil
		}
		return fmt.Errorf("failed to delete certificate: %w", err)
	}

	c.Logger.Info("certificate deleted successfully", "name", name, "namespace", namespace)
	return nil
}

// removeURLSchemeAndSlashes removes https://, http://, and all slashes from a string
func removeURLSchemeAndSlashes(value string) string {
	value = strings.TrimPrefix(value, "https://")
	value = strings.TrimPrefix(value, "http://")

	value = strings.ReplaceAll(value, "/", "")

	return value
}

// sanitizeLabelValue ensures the label value meets Kubernetes requirements
// Label values must be 63 characters or less and match regex: [a-z0-9A-Z]([-_.a-z0-9A-Z]*[a-z0-9A-Z])?
func sanitizeLabelValue(value string) string {
	// Remove URL scheme and slashes first
	value = removeURLSchemeAndSlashes(value)

	// Remove any characters not in the allowed set
	var result strings.Builder
	for _, char := range value {
		if (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' || char == '.' {
			result.WriteRune(char)
		}
	}

	sanitized := result.String()

	// Ensure it starts and ends with alphanumeric
	sanitized = strings.TrimLeft(sanitized, "-_.")
	sanitized = strings.TrimRight(sanitized, "-_.")

	// Truncate to 63 characters
	if len(sanitized) > 63 {
		sanitized = sanitized[:63]
		sanitized = strings.TrimRight(sanitized, "-_.")
	}

	if sanitized == "" {
		hash := sha256.Sum256([]byte(value))
		sanitized = hex.EncodeToString(hash[:8])
	}

	return sanitized
}
