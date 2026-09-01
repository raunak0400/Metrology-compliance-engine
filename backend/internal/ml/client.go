package ml

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

// Client is the HTTP client for the Python ML service. Every field the
// backend needs from ML flows through this type.
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

func (c *Client) post(ctx context.Context, path string, ref ImageRef, out interface{}) error {
	body, err := json.Marshal(ref)
	if err != nil {
		return fmt.Errorf("marshal ml request: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build ml request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("x-api-key", c.APIKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return apperror.ServiceUnavailable(fmt.Sprintf("Request to %s failed: %s", path, err.Error()))
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		text, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return apperror.ServiceUnavailable(fmt.Sprintf("Upstream %s returned %d: %s", path, resp.StatusCode, string(text)))
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return apperror.ServiceUnavailable(fmt.Sprintf("Invalid response from %s: %s", path, err.Error()))
	}
	return nil
}

func (c *Client) OCR(ctx context.Context, ref ImageRef) (*OcrResult, error) {
	var out OcrResult
	if err := c.post(ctx, "/ml/ocr", ref, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DetectLabel(ctx context.Context, ref ImageRef) (*DetectLabelResult, error) {
	var out DetectLabelResult
	if err := c.post(ctx, "/ml/detect-label", ref, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ExtractFields(ctx context.Context, ref ImageRef) (*ExtractFieldsResult, error) {
	var out ExtractFieldsResult
	if err := c.post(ctx, "/ml/extract-fields", ref, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Analyze(ctx context.Context, ref ImageRef) (*AnalyzeResult, error) {
	var out AnalyzeResult
	if err := c.post(ctx, "/ml/analyze", ref, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Health(ctx context.Context) (*HealthResult, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, c.BaseURL+"/health", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, apperror.ServiceUnavailable("ML service health check failed: " + err.Error())
	}
	defer resp.Body.Close()

	var out HealthResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, apperror.ServiceUnavailable("Invalid ML service health response")
	}
	return &out, nil
}
