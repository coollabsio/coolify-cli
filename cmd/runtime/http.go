package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APIError represents an error response from the API
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s (rc: %d)", e.Message, e.StatusCode)
}

// NewRequest creates a new HTTP request with authorization headers
func (c *Coolify) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s/api/v1/%s", c.Config.FQDN, path)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	// Set authorization headers
	req.Header.Set("Authorization", "Bearer "+c.Config.Token)
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		// If method is sending data, set content type to json
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// DoRequest performs the HTTP request and handles common response processing
func (c *Coolify) DoRequest(req *http.Request) ([]byte, error) {
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
		}

		// Try to parse error message from JSON response
		var errResp map[string]string
		if err := json.Unmarshal(body, &errResp); err == nil {
			if msg, ok := errResp["message"]; ok {
				apiErr.Message = msg
			}
		}

		// If no structured error message, use raw body
		if apiErr.Message == "" {
			apiErr.Message = string(body)
		}

		return nil, apiErr
	}

	return body, nil
}
