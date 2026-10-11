package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"sort"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	defaultRetries = 3
	apiV1Path      = "/api/v1/"

	// minUploadTimeout is the minimum per-request timeout for multipart
	// uploads, which carry far more data than regular JSON calls.
	minUploadTimeout = 10 * time.Minute
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

// Put makes a PUT request to the API
func (c *Client) Put(ctx context.Context, path string, body, result interface{}) error {
	return c.doRequest(ctx, "PUT", path, body, result)
}

// MultipartFile is the file part of a multipart/form-data request.
type MultipartFile struct {
	// FieldName is the form field name of the file part (e.g. "file").
	FieldName string
	// FileName is the client file name sent with the part.
	FileName string
	// Content is streamed into the request body. When it implements
	// io.Seeker, failed requests are retried from the start of Content.
	Content io.Reader
	// Size is the number of bytes Content yields. When >= 0 the request
	// carries an exact Content-Length; when < 0 it is sent chunked.
	Size int64
}

// PostMultipart makes a multipart/form-data POST request to the API. Fields
// are written as plain form fields before the file part. The file content is
// streamed, never buffered in memory.
func (c *Client) PostMultipart(ctx context.Context, path string, fields map[string]string, file MultipartFile, result interface{}) error {
	attempt := func() error {
		if seeker, ok := file.Content.(io.Seeker); ok {
			if _, err := seeker.Seek(0, io.SeekStart); err != nil {
				return fmt.Errorf("failed to rewind upload content: %w", err)
			}
		}
		return c.doMultipartOnce(ctx, path, fields, file, result)
	}
	if _, ok := file.Content.(io.Seeker); !ok {
		return attempt()
	}
	return c.withRetry(ctx, attempt)
}

func (c *Client) doMultipartOnce(ctx context.Context, path string, fields map[string]string, file MultipartFile, result interface{}) error {
	url := c.baseURL + apiV1Path + path

	// Render everything except the file content up front so the body can be
	// streamed as prefix + content + suffix with a known length.
	var head bytes.Buffer
	writer := multipart.NewWriter(&head)
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := writer.WriteField(key, fields[key]); err != nil {
			return fmt.Errorf("failed to write form field %q: %w", key, err)
		}
	}
	if _, err := writer.CreateFormFile(file.FieldName, file.FileName); err != nil {
		return fmt.Errorf("failed to write file part: %w", err)
	}
	prefix := append([]byte(nil), head.Bytes()...)
	head.Reset()
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to finish multipart body: %w", err)
	}
	suffix := append([]byte(nil), head.Bytes()...)

	if c.debug {
		log.Printf("POST %s", url)
		log.Printf("Request body: multipart/form-data fields=%v file=%s:%q (%d bytes)", keys, file.FieldName, file.FileName, file.Size)
	}

	body := io.MultiReader(bytes.NewReader(prefix), file.Content, bytes.NewReader(suffix))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	if file.Size >= 0 {
		req.ContentLength = int64(len(prefix)) + file.Size + int64(len(suffix))
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	uploadClient := *c.httpClient
	if uploadClient.Timeout > 0 && uploadClient.Timeout < minUploadTimeout {
		uploadClient.Timeout = minUploadTimeout
	}

	return c.execute(&uploadClient, req, path, result)
}

// GetVersion fetches the API version
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	var version string
	err := c.Get(ctx, "version", &version)
	return version, err
}

// doRequest executes an HTTP request with retry logic
func (c *Client) doRequest(ctx context.Context, method, path string, body, result interface{}) error {
	return c.withRetry(ctx, func() error {
		return c.doRequestOnce(ctx, method, path, body, result)
	})
}

// withRetry runs attempt up to c.retries+1 times with exponential backoff.
func (c *Client) withRetry(ctx context.Context, attemptFn func() error) error {
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

		err := attemptFn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on client errors (4xx) except 429 (rate limit)
		var apiErr *Error
		if errors.As(err, &apiErr) {
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
			log.Printf("Request body: %s", redactJSONForLog(jsonBody))
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.execute(c.httpClient, req, path, result)
}

// execute sends a prepared request with authentication, then handles debug
// logging, error mapping and response decoding.
func (c *Client) execute(httpClient *http.Client, req *http.Request, path string, result interface{}) error {
	req.Header.Set("Authorization", "Bearer "+c.token)
	// Without this, Laravel answers validation errors with a 302 redirect to an HTML page.
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := httpClient.Do(req) //nolint:gosec // G704: URL is the user-configured Coolify instance
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
		//nolint:gosec // G706: debug-only output; the response body is redacted
		log.Printf("Response status: %d", resp.StatusCode)
		log.Printf("Response body: %s", redactJSONForLog(respBody)) //nolint:gosec // G706: see above
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

func redactJSONForLog(jsonBody []byte) string {
	var value any
	if err := json.Unmarshal(jsonBody, &value); err != nil {
		return "<unavailable>"
	}
	redactSensitiveFields(value)
	redacted, err := json.Marshal(value)
	if err != nil {
		return "<unavailable>"
	}
	return string(redacted)
}

func redactSensitiveFields(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if key == "token" || key == "password" {
				typed[key] = "********"
				continue
			}
			redactSensitiveFields(child)
		}
	case []any:
		for _, child := range typed {
			redactSensitiveFields(child)
		}
	}
}
