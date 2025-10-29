package cli

import (
	"context"
	"fmt"

	compareVersion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/api"
)

// SensitiveInformationOverlay is the string used to hide sensitive data
const SensitiveInformationOverlay = "********"

// ExactArgs returns a validator that ensures exactly n arguments are provided with a helpful error message
func ExactArgs(n int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			if n == 1 {
				return fmt.Errorf("missing required argument: %s\n\nUsage: %s", usage, cmd.UseLine())
			}
			return fmt.Errorf("expected %d argument(s), got %d\n\nUsage: %s", n, len(args), cmd.UseLine())
		}
		return nil
	}
}

// MinArgs returns a validator that ensures at least n arguments are provided with a helpful error message
func MinArgs(n int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			return fmt.Errorf("missing required arguments: %s\n\nUsage: %s", usage, cmd.UseLine())
		}
		return nil
	}
}

// ParseInt parses a string to int with better error message
func ParseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return 0, fmt.Errorf("'%s' is not a valid integer", s)
	}
	return result, nil
}

// SplitOwnerRepo splits owner/repo string into parts
func SplitOwnerRepo(s string) []string {
	parts := make([]string, 0, 2)
	var current string
	for _, char := range s {
		if char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func StringPtr(s string) *string {
	return &s
}

func BoolPtr(b bool) *bool {
	return &b
}

// CheckMinimumVersion checks if the Coolify API version meets the minimum requirement
func CheckMinimumVersion(ctx context.Context, client *api.Client, minimumVersion string) error {
	currentVersionStr, err := client.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Coolify version: %w", err)
	}

	currentVersion, err := compareVersion.NewVersion(currentVersionStr)
	if err != nil {
		return fmt.Errorf("invalid current version '%s': %w", currentVersionStr, err)
	}

	minVersion, err := compareVersion.NewVersion(minimumVersion)
	if err != nil {
		return fmt.Errorf("invalid minimum version '%s': %w", minimumVersion, err)
	}

	if currentVersion.LessThan(minVersion) {
		return fmt.Errorf("this command requires Coolify version %s or higher, but the current version is %s", minimumVersion, currentVersionStr)
	}

	return nil
}
