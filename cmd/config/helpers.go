package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

const defaultFile = "coolify.json"

func configService(cmd *cobra.Command) (*service.ProjectConfigService, error) {
	client, err := cli.GetAPIClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get API client: %w", err)
	}
	return service.NewProjectConfigService(client), nil
}

func formatOf(cmd *cobra.Command) string {
	format, _ := cmd.Flags().GetString("format")
	return format
}

func showSensitive(cmd *cobra.Command) bool {
	show, _ := cmd.Flags().GetBool("show-sensitive")
	return show
}

func printData(cmd *cobra.Command, data any) error {
	formatter, err := output.NewFormatter(formatOf(cmd), output.Options{Writer: cmd.OutOrStdout(), ShowSensitive: showSensitive(cmd)})
	if err != nil {
		return err
	}
	return formatter.Format(data)
}

// readConfigFile reads a coolify.json file. "-" reads standard input.
func readConfigFile(cmd *cobra.Command, path string) (string, error) {
	var (
		data []byte
		err  error
	)
	if path == "-" {
		data, err = io.ReadAll(cmd.InOrStdin())
	} else {
		data, err = os.ReadFile(path) //nolint:gosec // G304: the user names the file to read
	}
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", path, err)
	}
	return string(data), nil
}

// confirm asks a yes/no question on the command input.
func confirm(cmd *cobra.Command, question string) (bool, error) {
	fmt.Fprintf(cmd.OutOrStdout(), "%s (yes/no): ", question)
	answer, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && answer == "" {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "yes" || answer == "y", nil
}

// planRow is one change of a plan in table output.
type planRow struct {
	Resource    string `json:"resource"`
	Change      string `json:"change"`
	Setting     string `json:"setting"`
	Destructive bool   `json:"destructive"`
}

// printPlan shows a plan: in table format as a change table followed by the
// errors, warnings, and missing required values; otherwise as JSON.
func printPlan(cmd *cobra.Command, plan *models.ConfigPlan) error {
	if formatOf(cmd) != output.FormatTable {
		return printData(cmd, plan)
	}
	out := cmd.OutOrStdout()
	if len(plan.Changes) == 0 {
		fmt.Fprintln(out, "No changes.")
	} else {
		rows := make([]planRow, 0, len(plan.Changes))
		for _, change := range plan.Changes {
			rows = append(rows, planRow{Resource: change.Key, Change: change.Kind, Setting: change.Path, Destructive: change.Destructive})
		}
		if err := printData(cmd, rows); err != nil {
			return err
		}
	}
	if plan.RemovesAll {
		fmt.Fprintln(out, "\nThis plan removes all resources of the environment.")
	}
	if plan.Destructive {
		fmt.Fprintln(out, "\nThis plan has destructive changes and waits for approval.")
	}
	if len(plan.Errors) > 0 {
		fmt.Fprintln(out, "\nErrors:")
		for _, e := range plan.Errors {
			if e.Path != "" {
				fmt.Fprintf(out, "  - %s: %s\n", e.Path, e.Message)
			} else {
				fmt.Fprintf(out, "  - %s\n", e.Message)
			}
		}
	}
	if len(plan.Warnings) > 0 {
		fmt.Fprintln(out, "\nWarnings:")
		for _, w := range plan.Warnings {
			fmt.Fprintf(out, "  - %s\n", w)
		}
	}
	if len(plan.MissingRequired) > 0 {
		fmt.Fprintln(out, "\nMissing required values:")
		keys := make([]string, 0, len(plan.MissingRequired))
		for key := range plan.MissingRequired {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(out, "  - %s: %s\n", key, strings.Join(plan.MissingRequired[key], ", "))
		}
	}
	return nil
}

// planErr returns an error when the plan has errors, so the command exits non-zero.
func planErr(plan *models.ConfigPlan) error {
	if plan == nil || len(plan.Errors) == 0 {
		return nil
	}
	return fmt.Errorf("the file has %d error(s)", len(plan.Errors))
}

// printRevision shows a revision and, in table format, its plan.
func printRevision(cmd *cobra.Command, revision *models.ConfigRevision) error {
	if err := printData(cmd, revision); err != nil {
		return err
	}
	if formatOf(cmd) != output.FormatTable {
		return nil
	}
	if revision.Plan != nil {
		fmt.Fprintln(cmd.OutOrStdout())
		if err := printPlan(cmd, revision.Plan); err != nil {
			return err
		}
	}
	if revision.Status == models.ConfigRevisionWaitingApproval {
		if revision.ApprovedAt == nil {
			fmt.Fprintf(cmd.OutOrStdout(), "\nApprove with: coolify config approve %s\n", revision.UUID)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "\nApproved. Coolify applies the revision in the background. Check with: coolify config revisions <link_uuid>")
		}
	}
	return nil
}

// revisionErr returns an error for a failed revision, so the command exits non-zero.
func revisionErr(revision *models.ConfigRevision) error {
	if revision == nil || revision.Status != models.ConfigRevisionFailed {
		return nil
	}
	if revision.Error != nil && *revision.Error != "" {
		return fmt.Errorf("revision %s failed: %s", revision.UUID, *revision.Error)
	}
	return fmt.Errorf("revision %s failed", revision.UUID)
}

// printLink shows a link and, in table format, the webhook details of a new git link.
func printLink(cmd *cobra.Command, link *models.ConfigLink) error {
	if err := printData(cmd, link); err != nil {
		return err
	}
	if formatOf(cmd) != output.FormatTable || len(link.WebhookURLs) == 0 {
		return nil
	}
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "\nWebhook URLs:")
	providers := make([]string, 0, len(link.WebhookURLs))
	for provider := range link.WebhookURLs {
		providers = append(providers, provider)
	}
	sort.Strings(providers)
	for _, provider := range providers {
		fmt.Fprintf(out, "  %s: %s\n", provider, link.WebhookURLs[provider])
	}
	if link.WebhookSecret != nil {
		secret := output.SensitiveOverlay + " (use --show-sensitive to show it)"
		if showSensitive(cmd) {
			secret = *link.WebhookSecret
		}
		fmt.Fprintf(out, "Webhook secret: %s\n", secret)
	}
	return nil
}
