package notification

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

var channels = []string{"email", "discord", "slack", "telegram", "pushover", "webhook"}

func NewNotificationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "notification",
		Aliases: []string{"notifications"},
		Short:   "Manage team notification channel settings",
	}
	cmd.AddCommand(newGet(), newUpdate())
	return cmd
}

func svc(cmd *cobra.Command) (*service.NotificationService, error) {
	client, err := cli.GetAPIClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get API client: %w", err)
	}
	return service.NewNotificationService(client), nil
}

func validateChannel(channel string) error {
	for _, c := range channels {
		if c == channel {
			return nil
		}
	}
	return fmt.Errorf("channel must be one of: %s", strings.Join(channels, ", "))
}

func newGet() *cobra.Command {
	return &cobra.Command{
		Use:   "get <channel>",
		Short: "Get notification settings for a channel",
		Long:  "Channels: email, discord, slack, telegram, pushover, webhook",
		Args:  cli.ExactArgs(1, "<channel>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateChannel(args[0]); err != nil {
				return err
			}
			s, err := svc(cmd)
			if err != nil {
				return err
			}
			settings, err := s.Get(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("failed to get notification settings: %w", err)
			}
			formatName, _ := cmd.Flags().GetString("format")
			if formatName == "table" {
				formatName = "pretty"
			}
			formatter, err := output.NewFormatter(formatName, output.Options{})
			if err != nil {
				return err
			}
			return formatter.Format(settings)
		},
	}
}

func newUpdate() *cobra.Command {
	var jsonBody string
	cmd := &cobra.Command{
		Use:   "update <channel>",
		Short: "Update notification settings (JSON body)",
		Long:  `Pass settings as JSON via --json. Example: coolify notification update webhook --json '{"webhook_enabled":true}'`,
		Args:  cli.ExactArgs(1, "<channel>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateChannel(args[0]); err != nil {
				return err
			}
			if jsonBody == "" {
				return fmt.Errorf("--json is required")
			}
			var body map[string]any
			if err := json.Unmarshal([]byte(jsonBody), &body); err != nil {
				return fmt.Errorf("invalid --json: %w", err)
			}
			s, err := svc(cmd)
			if err != nil {
				return err
			}
			settings, err := s.Update(cmd.Context(), args[0], body)
			if err != nil {
				return fmt.Errorf("failed to update notification settings: %w", err)
			}
			formatName, _ := cmd.Flags().GetString("format")
			if formatName == "table" {
				formatName = "pretty"
			}
			formatter, err := output.NewFormatter(formatName, output.Options{})
			if err != nil {
				return err
			}
			return formatter.Format(settings)
		},
	}
	cmd.Flags().StringVar(&jsonBody, "json", "", "JSON object of fields to update")
	return cmd
}
