package api

import (
	"net/http"
	"time"
)

// Option configures the API client
type Option func(*Client)

// WithDebug enables debug logging
func WithDebug(debug bool) Option {
	return func(c *Client) {
		c.debug = debug
	}
}

// WithTimeout sets the request timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithRetries sets the number of retries for failed requests
func WithRetries(retries int) Option {
	return func(c *Client) {
		c.retries = retries
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}
