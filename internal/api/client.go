// Package api provides an HTTP client for the JamesBot control API.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"jamesbot/internal/control"
)

// Client is an HTTP client for the control API.
type Client struct {
	endpoint    string
	statsURL    string
	rulesURL    string
	rulesSetURL string
	httpClient  *http.Client
}

// NewClient creates a new API client.
func NewClient(endpoint string) *Client {
	endpoint = strings.TrimSuffix(endpoint, "/")
	return &Client{
		endpoint:    endpoint,
		statsURL:    endpoint + "/stats",
		rulesURL:    endpoint + "/rules",
		rulesSetURL: endpoint + "/rules/set",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Timeout returns the HTTP client timeout duration.
func (c *Client) Timeout() time.Duration {
	if c == nil || c.httpClient == nil {
		return 0
	}
	return c.httpClient.Timeout
}

// GetStats retrieves bot statistics from the control API.
func (c *Client) GetStats() (*control.Stats, error) {
	if c == nil {
		return nil, fmt.Errorf("client is nil")
	}

	resp, err := c.httpClient.Get(c.statsURL)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var stats control.Stats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return &stats, nil
}

// ListRules retrieves all moderation rules from the control API.
func (c *Client) ListRules() ([]control.Rule, error) {
	if c == nil {
		return nil, fmt.Errorf("client is nil")
	}

	resp, err := c.httpClient.Get(c.rulesURL)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var rules []control.Rule
	if err := json.NewDecoder(resp.Body).Decode(&rules); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return rules, nil
}

// SetRule modifies a rule setting via the control API.
func (c *Client) SetRule(name, key, value string) error {
	if c == nil {
		return fmt.Errorf("client is nil")
	}

	body, err := json.Marshal(map[string]string{
		"name":  name,
		"key":   key,
		"value": value,
	})
	if err != nil {
		return fmt.Errorf("encode failed: %w", err)
	}

	resp, err := c.httpClient.Post(c.rulesSetURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("rule update failed: status %d", resp.StatusCode)
	}

	return nil
}
