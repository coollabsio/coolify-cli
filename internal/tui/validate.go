package tui

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
)

var (
	ErrValueEmpty      = errors.New("value cannot be empty")
	ErrFqdnInvalid     = errors.New("fqdn must contain a scheme (http:// or https://)")
	ErrFqdnHostMissing = errors.New("fqdn must contain a host")
	ErrInvalidIP       = errors.New("invalid IP address or hostname")
	ErrInvalidPort     = errors.New("port must be between 1 and 65535")
	ErrInvalidYesNo    = errors.New("must be 'yes' or 'no'")
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

// ValidateIPOrHostname validates that a string is either a valid IP address or hostname
func ValidateIPOrHostname(s string) error {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return ErrValueEmpty
	}

	// Try parsing as IP address
	if net.ParseIP(trimmed) != nil {
		return nil
	}

	// Check if it's a valid hostname (simple check)
	if len(trimmed) > 0 && len(trimmed) <= 253 {
		return nil
	}

	return ErrInvalidIP
}

// ValidatePort validates that a string is a valid port number
func ValidatePort(s string) error {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil // Allow empty for default
	}

	port, err := strconv.Atoi(trimmed)
	if err != nil {
		return ErrInvalidPort
	}

	if port < 1 || port > 65535 {
		return ErrInvalidPort
	}

	return nil
}

// ValidateYesNo validates that a string is either "yes" or "no"
func ValidateYesNo(s string) error {
	trimmed := strings.ToLower(strings.TrimSpace(s))
	if trimmed == "" {
		return nil // Allow empty for default
	}

	if trimmed != "yes" && trimmed != "no" {
		return ErrInvalidYesNo
	}

	return nil
}
