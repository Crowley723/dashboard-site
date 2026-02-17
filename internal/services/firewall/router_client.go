package firewall

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"homelab-dashboard/internal/config"
	"io"
	"net/http"
	"time"
)

type RouterClient struct {
	config     config.Config
	httpClient *http.Client
}

func NewRouterClient(cfg config.Config) *RouterClient {
	return &RouterClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *RouterClient) normalizeEndpoint() string {
	endpoint := c.config.Features.FirewallManagement.RouterEndpoint
	if len(endpoint) > 0 && endpoint[len(endpoint)-1] == '/' {
		return endpoint[:len(endpoint)-1]
	}
	return endpoint
}

const (
	fmtGETAliasByUUIDPath = "/api/firewall/alias/get_item/%s"
	fmtSETAliasByUUIDPath = "/api/firewall/alias/set_item/%s"
	reconfigurePath       = "/api/firewall/alias/reconfigure"
)

// GetAliasIPs retrieves current IPs from an OPNsense alias
func (c *RouterClient) GetAliasIPs(ctx context.Context, aliasUUID string) ([]string, error) {
	url := c.normalizeEndpoint() + fmt.Sprintf(fmtGETAliasByUUIDPath, aliasUUID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(
		c.config.Features.FirewallManagement.RouterAPIKey,
		c.config.Features.FirewallManagement.RouterAPISecret,
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var aliasResp AliasGetResponse
	if err := json.NewDecoder(resp.Body).Decode(&aliasResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	ips := aliasResp.Alias.GetSelectedIPs()

	return ips, nil
}

// UpdateAlias updates an OPNsense alias with new IPs
func (c *RouterClient) UpdateAlias(ctx context.Context, aliasUUID string, ipsToAdd, ipsToRemove []string) error {
	currentIPs, err := c.GetAliasIPs(ctx, aliasUUID)
	if err != nil {
		return fmt.Errorf("failed to get current alias state: %w", err)
	}

	ipMap := make(map[string]bool)

	for _, ip := range currentIPs {
		ipMap[ip] = true
	}

	for _, ip := range ipsToRemove {
		delete(ipMap, ip)
	}

	for _, ip := range ipsToAdd {
		ipMap[ip] = true
	}

	var newIPs []string
	for ip := range ipMap {
		newIPs = append(newIPs, ip)
	}

	url := c.normalizeEndpoint() + fmt.Sprintf(fmtGETAliasByUUIDPath, aliasUUID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create get request: %w", err)
	}
	req.SetBasicAuth(
		c.config.Features.FirewallManagement.RouterAPIKey,
		c.config.Features.FirewallManagement.RouterAPISecret,
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("get request failed: %w", err)
	}
	defer resp.Body.Close()

	var fullAlias AliasGetResponse
	if err := json.NewDecoder(resp.Body).Decode(&fullAlias); err != nil {
		return fmt.Errorf("failed to decode alias: %w", err)
	}

	// 4. Build set request with newline-separated IPs
	content := ""
	for i, ip := range newIPs {
		if i > 0 {
			content += "\n"
		}
		content += ip
	}

	setReq := AliasSetRequest{
		Alias: AliasSetBody{
			Enabled:        fullAlias.Alias.Enabled,
			Name:           fullAlias.Alias.Name,
			Type:           getSelected(fullAlias.Alias.Type),
			Proto:          getSelected(fullAlias.Alias.Proto),
			Categories:     getSelected(fullAlias.Alias.Categories),
			UpdateFreq:     fullAlias.Alias.UpdateFreq,
			Content:        content,
			PathExpression: fullAlias.Alias.PathExpression,
			AuthType:       getSelected(fullAlias.Alias.AuthType),
			Username:       fullAlias.Alias.Username,
			Password:       fullAlias.Alias.Password,
			Interface:      getSelected(fullAlias.Alias.Interface),
			Counters:       fullAlias.Alias.Counters,
			Description:    fullAlias.Alias.Description,
		},
	}

	setURL := c.normalizeEndpoint() + fmt.Sprintf(fmtSETAliasByUUIDPath, aliasUUID)
	jsonBody, err := json.Marshal(setReq)
	if err != nil {
		return fmt.Errorf("failed to marshal set request: %w", err)
	}

	setHTTPReq, err := http.NewRequestWithContext(ctx, "POST", setURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create set request: %w", err)
	}
	setHTTPReq.SetBasicAuth(
		c.config.Features.FirewallManagement.RouterAPIKey,
		c.config.Features.FirewallManagement.RouterAPISecret,
	)
	setHTTPReq.Header.Set("Content-Type", "application/json")

	setResp, err := c.httpClient.Do(setHTTPReq)
	if err != nil {
		return fmt.Errorf("set request failed: %w", err)
	}
	defer setResp.Body.Close()

	if setResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(setResp.Body)
		return fmt.Errorf("set request returned status %d: %s", setResp.StatusCode, string(body))
	}

	reconfigURL := c.normalizeEndpoint() + reconfigurePath
	reconfigReq, err := http.NewRequestWithContext(ctx, "POST", reconfigURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create reconfigure request: %w", err)
	}
	reconfigReq.SetBasicAuth(
		c.config.Features.FirewallManagement.RouterAPIKey,
		c.config.Features.FirewallManagement.RouterAPISecret,
	)

	reconfigResp, err := c.httpClient.Do(reconfigReq)
	if err != nil {
		return fmt.Errorf("reconfigure request failed: %w", err)
	}
	defer reconfigResp.Body.Close()

	if reconfigResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(reconfigResp.Body)
		return fmt.Errorf("reconfigure returned status %d: %s", reconfigResp.StatusCode, string(body))
	}

	return nil
}
