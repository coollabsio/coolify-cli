package firewall

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/ssh"
)

// newContainersCommand builds `coolify firewall containers`.
func newContainersCommand(flags *FirewallFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "containers",
		Short: "List containers on coolify-mesh across all servers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runContainers(cmd.Context(), cmd, flags)
		},
	}
}

func runContainers(ctx context.Context, cmd *cobra.Command, flags *FirewallFlags) error {
	if err := flags.SSHMeshFlags.Validate(); err != nil {
		return err
	}
	runner, err := flags.BuildSSHClient()
	if err != nil {
		return fmt.Errorf("SSH client: %w", err)
	}
	return emitContainers(ctx, cmd, flags, runner)
}

// emitContainers is factored out so tests can pass a fake ssh.Runner.
func emitContainers(
	ctx context.Context,
	cmd *cobra.Command,
	flags *FirewallFlags,
	runner ssh.Runner,
) error {
	all, results := discoverAllViaPkg(ctx, runner, flags)

	rows := make([]models.ContainerRow, 0, len(all))
	for _, c := range all {
		rows = append(rows, models.ContainerRow{
			Host: c.Host, ID: c.ID, Name: c.Name, IP: c.IP.String(),
		})
	}

	var errs []string
	for _, r := range results {
		if r.Err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", r.Host, r.Err))
		}
	}
	for _, e := range errs {
		fmt.Fprintln(os.Stderr, "Warning:", e)
	}

	format, _ := cmd.Root().PersistentFlags().GetString("format")
	if format == "" {
		format = output.FormatTable
	}
	formatter, err := output.NewFormatter(format, output.Options{Writer: os.Stdout})
	if err != nil {
		return err
	}
	if format == output.FormatJSON || format == output.FormatPretty {
		return formatter.Format(models.FirewallContainersOutput{
			Containers: rows, Errors: errs,
		})
	}
	if len(rows) == 0 {
		fmt.Fprintln(os.Stderr, "No containers found on coolify-mesh network.")
		return nil
	}
	return formatter.Format(rows)
}
