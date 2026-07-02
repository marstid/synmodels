// Package api provides HTTP client functionality for fetching models from the API.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/marstid/synmodels/internal/types"
)

const (
	defaultTimeout = 30 * time.Second
	defaultBaseURL = "https://api.synthetic.new/openai/v1"
)

// Client is an HTTP client for fetching models from the API.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new API client with default configuration.
func NewClient() *Client {
	baseURL := defaultBaseURL
	if envURL := os.Getenv("SYN_API"); envURL != "" {
		baseURL = envURL
	}
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		baseURL: baseURL,
	}
}

// NewClientWithURL creates a new API client with a custom base URL (useful for testing).
func NewClientWithURL(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		baseURL: baseURL,
	}
}

// FetchModels retrieves the list of models from the API.
func (c *Client) FetchModels(ctx context.Context) ([]types.Model, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to match the expected API format
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		if len(body) > 0 {
			return nil, fmt.Errorf("unexpected status code: %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var modelsResp types.ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return modelsResp.Data, nil
}
