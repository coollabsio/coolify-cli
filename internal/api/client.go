package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	defaultRetries = 3
	apiV1Path      = "/api/v1/"
)

// Client is the HTTP client for Coolify API
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	debug      bool
	retries    int
	timeout    time.Duration
}

// NewClient creates a new API client
func NewClient(baseURL, token string, opts ...Option) *Client {
	c := &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{},
		timeout:    defaultTimeout,
		retries:    defaultRetries,
		debug:      false,
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	// Set timeout on HTTP client
	c.httpClient.Timeout = c.timeout

	return c
}

// Get makes a GET request to the API
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.doRequest(ctx, "GET", path, nil, result)
}

// Post makes a POST request to the API
func (c *Client) Post(ctx context.Context, path string, body, result interface{}) error {
	return c.doRequest(ctx, "POST", path, body, result)
}

// Delete makes a DELETE request to the API
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.doRequest(ctx, "DELETE", path, nil, nil)
}

// Patch makes a PATCH request to the API
func (c *Client) Patch(ctx context.Context, path string, body, result interface{}) error {
	return c.doRequest(ctx, "PATCH", path, body, result)
}

// GetVersion fetches the API version
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	var version string
	err := c.Get(ctx, "version", &version)
	return version, err
}

// doRequest executes an HTTP request with retry logic
func (c *Client) doRequest(ctx context.Context, method, path string, body, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			// Always log retries so users know what's happening
			log.Printf("Request failed, retrying (attempt %d/%d) after %v...", attempt, c.retries, backoff)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := c.doRequestOnce(ctx, method, path, body, result)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on client errors (4xx) except 429 (rate limit)
		if apiErr, ok := err.(*Error); ok {
			if apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 && apiErr.StatusCode != 429 {
				return err
			}
		}

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	return lastErr
}

// doRequestOnce executes a single HTTP request
func (c *Client) doRequestOnce(ctx context.Context, method, path string, body, result interface{}) error {
	url := c.baseURL + apiV1Path + path

	if c.debug {
		log.Printf("%s %s", method, url)
	}

	// Prepare request body
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)

		if c.debug {
			log.Printf("Request body: %s", string(jsonBody))
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if c.debug {
		log.Printf("Response status: %d", resp.StatusCode)
		log.Printf("Response body: %s", string(respBody))
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := string(respBody)
		if message == "" {
			message = "Unknown error"
		} else {
			// Try to extract error message from JSON
			var errResp struct {
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			if err := json.Unmarshal(respBody, &errResp); err == nil {
				if errResp.Message != "" {
					message = errResp.Message
				} else if errResp.Error != "" {
					message = errResp.Error
				}
			}
		}

		return NewError(resp.StatusCode, path, message)
	}

	// Unmarshal response into result
	if result != nil {
		// Handle string responses
		if strResult, ok := result.(*string); ok {
			*strResult = string(respBody)
			return nil
		}

		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
