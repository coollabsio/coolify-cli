package api

import (
	"errors"
	"fmt"
)

// Error represents an API error response
type Error struct {
	StatusCode int
	Message    string
	Path       string
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error %d on %s: %s", e.StatusCode, e.Path, e.Message)
	}
	return fmt.Sprintf("API error %d on %s", e.StatusCode, e.Path)
}

// NewError creates a new API error
func NewError(statusCode int, path, message string) *Error {
	return &Error{
		StatusCode: statusCode,
		Path:       path,
		Message:    message,
	}
}

// IsNotFound checks if the error is a 404 Not Found error
func IsNotFound(err error) bool {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsUnauthorized checks if the error is a 401 or 403 error
func IsUnauthorized(err error) bool {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401 || apiErr.StatusCode == 403
	}
	return false
}

// IsBadRequest checks if the error is a 400 Bad Request error
func IsBadRequest(err error) bool {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 400
	}
	return false
}

// IsServerError checks if the error is a 5xx server error
func IsServerError(err error) bool {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 500 && apiErr.StatusCode < 600
	}
	return false
}
