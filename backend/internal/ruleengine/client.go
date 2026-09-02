package ruleengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/raunak0400/metrology-backend/internal/apperror"
)

// Client is the HTTP client for the Python Legal Metrology Rule Engine
// service. Every compliance decision the backend needs flows through this
// type — see package doc.
type Client struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
	http    *http.Client
}

func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Timeout: timeout,
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal rule engine request: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	reqCtx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, method, c.BaseURL+path, reader)
	if err != nil {
		return fmt.Errorf("build rule engine request: %w", err)
	}
	if body != nil {
		req.Header.Set("content-type", "application/json")
	}
	if c.APIKey != "" {
		req.Header.Set("x-api-key", c.APIKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return apperror.ServiceUnavailable(fmt.Sprintf("Request to %s failed: %s", path, err.Error()))
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		text, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		return apperror.ServiceUnavailable(fmt.Sprintf("Upstream %s returned %d: %s", path, resp.StatusCode, string(text)))
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return apperror.ServiceUnavailable(fmt.Sprintf("Invalid response from %s: %s", path, err.Error()))
	}
	return nil
}

// Validate calls POST /validate — the deterministic compliance decision. The
// rule engine performs no OCR/ML; it only evaluates the structured input
// already extracted upstream.
func (c *Client) Validate(ctx context.Context, req ValidationRequest) (*ValidationResponse, error) {
	var out ValidationResponse
	if err := c.do(ctx, http.MethodPost, "/validate", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Rules calls GET /rules — the authoritative rule registry.
func (c *Client) Rules(ctx context.Context) (*RuleRegistryResponse, error) {
	var out RuleRegistryResponse
	if err := c.do(ctx, http.MethodGet, "/rules", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Health calls GET /health.
func (c *Client) Health(ctx context.Context) (*HealthResult, error) {
	var out HealthResult
	if err := c.do(ctx, http.MethodGet, "/health", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
