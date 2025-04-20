package tui

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrValueEmpty      = errors.New("value cannot be empty")
	ErrFqdnInvalid     = errors.New("fqdn must contain a scheme (http:// or https://)")
	ErrFqdnHostMissing = errors.New("fqdn must contain a host")
)

// ValidateNotEmpty validates that a string is not empty even if it has whitespace
func ValidateNotEmpty(s string) error {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return ErrValueEmpty
	}
	return nil
}

// ValidateFQDN validates that a string is a valid FQDN with scheme and host
func ValidateFQDN(s string) error {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return ErrValueEmpty
	}

	if !strings.Contains(trimmed, "://") {
		return ErrFqdnInvalid
	}

	u, err := url.Parse(trimmed)
	if err != nil {
		return err
	}
	if u.Host == "" {
		return ErrFqdnHostMissing
	}
	return nil
}
