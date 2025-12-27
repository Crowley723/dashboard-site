package k8s

import (
	"context"
	"fmt"
	"homelab-dashboard/internal/config"
	"log/slog"
	"os"
	"path/filepath"

	certmanagerclientset "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes and cert-manager clients
type Client struct {
	ClientSet         *kubernetes.Clientset
	CertManagerClient *certmanagerclientset.Clientset
	Config            *rest.Config
	Namespace         string
	IssuerName        string
	IssuerKind        string
	Logger            *slog.Logger
}

// NewClient creates a new Kubernetes client based on the configuration
func NewClient(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*Client, error) {
	if cfg.Features == nil || !cfg.Features.MTLSManagement.Enabled {
		return nil, fmt.Errorf("mtls_management is not enabled")
	}

	if cfg.Features.MTLSManagement.Kubernetes == nil {
		return nil, fmt.Errorf("kubernetes configuration is missing")
	}

	k8sCfg := cfg.Features.MTLSManagement.Kubernetes
	issuerCfg := cfg.Features.MTLSManagement.CertificateIssuer

	var restConfig *rest.Config
	var err error

	if k8sCfg.InCluster {
		logger.Info("using in-cluster Kubernetes configuration")
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
		}
	} else {
		kubeconfig := k8sCfg.Kubeconfig
		if kubeconfig == "" {
			// Use default kubeconfig location
			if home := homeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			}
		}

		logger.Info("using kubeconfig file", "path", kubeconfig)
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
		}
	}

	// Create standard Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	// Create cert-manager clientset
	certManagerClient, err := certmanagerclientset.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create cert-manager clientset: %w", err)
	}

	// Verify connection by listing namespaces
	_, err = clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return nil, fmt.Errorf("failed to verify kubernetes connection: %w", err)
	}

	logger.Info("kubernetes client initialized successfully",
		"namespace", k8sCfg.Namespace,
		"issuer", issuerCfg.Name,
		"issuerKind", issuerCfg.Kind)

	return &Client{
		ClientSet:         clientset,
		CertManagerClient: certManagerClient,
		Config:            restConfig,
		Namespace:         k8sCfg.Namespace,
		IssuerName:        issuerCfg.Name,
		IssuerKind:        issuerCfg.Kind,
		Logger:            logger,
	}, nil
}

// homeDir returns the user's home directory
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // Windows
}
