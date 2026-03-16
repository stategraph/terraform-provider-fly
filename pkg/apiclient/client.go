package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const (
	DefaultBaseURL = "https://api.machines.dev/v1"
)

type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
	userAgent  string
	limiter    *rate.Limiter
}

type ClientOption func(*Client)

func WithBaseURL(url string) ClientOption {
	return func(c *Client) { c.baseURL = url }
}

func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) { c.httpClient = hc }
}

func NewClient(token, version string, opts ...ClientOption) *Client {
	c := &Client{
		token:      token,
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: 2 * time.Minute},
		userAgent:  fmt.Sprintf("terraform-provider-fly/%s", version),
		limiter:    rate.NewLimiter(rate.Limit(10), 10), // 10 req/s with burst of 10
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) doRequest(ctx context.Context, method, url string, body any) (*http.Response, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter: %w", err)
	}

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	return resp, nil
}

func (c *Client) doJSON(ctx context.Context, method, url string, body any, result any) error {
	resp, err := c.doRequest(ctx, method, url, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return parseAPIError(resp.StatusCode, respBody)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("decoding response (status %d): %w", resp.StatusCode, err)
		}
	}

	return nil
}

func (c *Client) doJSONWithRetry(ctx context.Context, method, url string, body any, result any) error {
	var lastErr error
	for attempt := range 3 {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}
		lastErr = c.doJSON(ctx, method, url, body, result)
		if lastErr == nil {
			return nil
		}
		if apiErr, ok := lastErr.(*APIError); ok {
			if apiErr.StatusCode == 429 || apiErr.StatusCode >= 500 {
				continue
			}
			return lastErr
		}
		// Network errors are retryable
	}
	return lastErr
}

func (c *Client) restURL(path string) string {
	return c.baseURL + path
}
