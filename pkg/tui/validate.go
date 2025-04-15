package tui

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ValueEmptyError      = errors.New("value cannot be empty")
	FQDNInvalidError     = errors.New("FQDN must contain a scheme (http:// or https://)")
	FQDNHostMissingError = errors.New("FQDN must contain a host")
)

// ValidateNotEmpty validates that a string is not empty even if it has whitespace
func ValidateNotEmpty(s string) error {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return ValueEmptyError
	}
	return nil
}

// ValidateFQDN validates that a string is a valid FQDN with scheme and host
func ValidateFQDN(s string) error {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return ValueEmptyError
	}

	if !strings.Contains(trimmed, "://") {
		return FQDNInvalidError
	}

	u, err := url.Parse(trimmed)
	if err != nil {
		return err
	}
	if u.Host == "" {
		return FQDNHostMissingError
	}
	return nil
}
