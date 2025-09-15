package data

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type MimirClient struct {
	api v1.API
}

func NewMimirClient(baseUrl, username, password string) (*MimirClient, error) {
	cfg := api.Config{
		Address: baseUrl,
	}

	if username != "" && password != "" {
		cfg.RoundTripper = &BasicAuthTransport{
			Username: username,
			Password: password,
			Proxied:  api.DefaultRoundTripper,
		}
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	v1api := v1.NewAPI(client)

	return &MimirClient{api: v1api}, nil
}

func (m *MimirClient) Query(ctx context.Context, query string, timestamp time.Time) (model.Value, error) {
	result, warnings, err := m.api.Query(ctx, query, timestamp)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	if len(warnings) > 0 {
		fmt.Printf("query warnings: %v\n", warnings)
	}

	return result, nil
}

func (m *MimirClient) QueryRange(ctx context.Context, query string, r v1.Range) (model.Value, error) {
	result, warnings, err := m.api.QueryRange(ctx, query, r)
	if err != nil {
		return nil, fmt.Errorf("range query failed: %w", err)
	}

	if len(warnings) > 0 {
		fmt.Printf("range query warnings: %v\n", warnings)
	}

	return result, nil
}

type BasicAuthTransport struct {
	Username string
	Password string
	Proxied  http.RoundTripper
}

func (b *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if b.Username != "" && b.Password != "" {
		req.SetBasicAuth(b.Username, b.Password)
	}
	return b.Proxied.RoundTrip(req)
}
