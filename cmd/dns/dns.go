// Package dns has the dns commands of applications and services: Cloudflare DNS records for their domains.
package dns

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// Resource describes the resource type a dns command tree works on.
type Resource struct {
	// APIType is the first path segment of the API route: applications or services.
	APIType string
	// Command is the CLI command of the resource (app, service), used in examples.
	Command string
	// Arg is the usage name of the UUID argument, for example "<app_uuid>".
	Arg string
}

// statusRow is one table row of the DNS status: one hostname and provider.
type statusRow struct {
	Hostname             string `json:"hostname"`
	Zone                 string `json:"zone"`
	IntegrationTokenUUID string `json:"integration_token_uuid"`
	AutomaticDNS         string `json:"automatic_dns"`
	Record               string `json:"record"`
}

// NewCommand creates the dns parent command for the resource type.
func NewCommand(r Resource) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dns",
		Short: "Manage Cloudflare DNS records for the domains of a resource",
		Long: `Create, replace, and release Cloudflare DNS records that point the domains of a resource to its server.
Add a Cloudflare token first with: coolify integration-token create --provider cloudflare`,
	}
	cmd.AddCommand(newStatusCommand(r), newCreateCommand(r), newReplaceCommand(r), newReleaseCommand(r))
	return cmd
}

func newStatusCommand(r Resource) *cobra.Command {
	return &cobra.Command{
		Use:   "status " + r.Arg,
		Short: "Show the DNS providers and managed records of each domain",
		Args:  cli.ExactArgs(1, r.Arg),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := newService(cmd)
			if err != nil {
				return err
			}
			resp, err := svc.Status(cmd.Context(), r.APIType, args[0])
			if err != nil {
				return err
			}
			return format(cmd, resp, statusRows(resp))
		},
	}
}

func newCreateCommand(r Resource) *cobra.Command {
	var req models.DNSRecordCreateRequest
	cmd := &cobra.Command{
		Use:   "create " + r.Arg,
		Short: "Create the DNS record of a domain",
		Long: `Create a DNS record at Cloudflare that points a domain of the resource to a public address of its server.
An existing record with the same address is adopted. When another record uses the hostname, the command
fails and prints the replace command; ask before you run it.`,
		Example: fmt.Sprintf("  coolify %s dns create %s --hostname app.example.com", r.Command, r.Arg),
		Args:    cli.ExactArgs(1, r.Arg),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := newService(cmd)
			if err != nil {
				return err
			}
			record, err := svc.Create(cmd.Context(), r.APIType, args[0], trimCreate(req))
			if conflict := (*service.DNSConflictError)(nil); errors.As(err, &conflict) {
				return fmt.Errorf("%w\nTo overwrite it: coolify %s dns replace %s --hostname %s --record-id %s --current %s",
					err, r.Command, args[0], strings.ToLower(strings.TrimSpace(req.Hostname)), conflict.Conflict.RecordID, conflict.Conflict.Current)
			}
			if err != nil {
				return err
			}
			return format(cmd, record, *record)
		},
	}
	cmd.Flags().StringVar(&req.Hostname, "hostname", "", "A saved domain of the resource (required)")
	cmd.Flags().StringVar(&req.IntegrationTokenUUID, "integration-token", "", "Cloudflare token to use when several manage the zone")
	cmd.Flags().StringVar(&req.Address, "address", "", "A public address of the server (default: the first one)")
	return cmd
}

func newReplaceCommand(r Resource) *cobra.Command {
	var req models.DNSRecordReplaceRequest
	cmd := &cobra.Command{
		Use:   "replace " + r.Arg,
		Short: "Overwrite the conflicting record reported by create",
		Long: `Point the conflicting record that create reported to the server. The record is changed only when it
still has the ID and value of the conflict.`,
		Args: cli.ExactArgs(1, r.Arg),
		RunE: func(cmd *cobra.Command, args []string) error {
			if force, _ := cmd.Flags().GetBool("force"); !force {
				ok, err := cli.Confirm(cmd, fmt.Sprintf("Overwrite the DNS record of %s that points to %s?", req.Hostname, req.Current))
				if err != nil || !ok {
					return err
				}
			}
			svc, err := newService(cmd)
			if err != nil {
				return err
			}
			req.Hostname = strings.TrimSpace(req.Hostname)
			req.RecordID = strings.TrimSpace(req.RecordID)
			req.Current = strings.TrimSpace(req.Current)
			req.IntegrationTokenUUID = strings.TrimSpace(req.IntegrationTokenUUID)
			req.Address = strings.TrimSpace(req.Address)
			record, err := svc.Replace(cmd.Context(), r.APIType, args[0], req)
			if err != nil {
				return err
			}
			return format(cmd, record, *record)
		},
	}
	cmd.Flags().StringVar(&req.Hostname, "hostname", "", "A saved domain of the resource (required)")
	cmd.Flags().StringVar(&req.RecordID, "record-id", "", "Record ID from the conflict (required)")
	cmd.Flags().StringVar(&req.Current, "current", "", "Current value from the conflict (required)")
	cmd.Flags().StringVar(&req.IntegrationTokenUUID, "integration-token", "", "Cloudflare token to use when several manage the zone")
	cmd.Flags().StringVar(&req.Address, "address", "", "A public address of the server (default: the first one)")
	cmd.Flags().Bool("force", false, "Skip the confirmation")
	return cmd
}

func newReleaseCommand(r Resource) *cobra.Command {
	var keepRecord bool
	cmd := &cobra.Command{
		Use:   "release " + r.Arg + " <hostname>",
		Short: "Release the managed record of a hostname the resource no longer uses",
		Long: `Release the managed DNS record of a hostname that the resource no longer uses. Coolify deletes the record
at Cloudflare only when it created it, the record did not change, and no other resource uses the hostname.
Removing a domain already queues this release; use this command to retry it or, with --keep-record, to keep it.`,
		Args: cli.ExactArgs(2, r.Arg+" <hostname>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if force, _ := cmd.Flags().GetBool("force"); !force && !keepRecord {
				ok, err := cli.Confirm(cmd, "Release and delete the DNS record of "+args[1]+"?")
				if err != nil || !ok {
					return err
				}
			}
			svc, err := newService(cmd)
			if err != nil {
				return err
			}
			resp, err := svc.Release(cmd.Context(), r.APIType, args[0], args[1], !keepRecord)
			if err != nil {
				return err
			}
			return format(cmd, resp, *resp)
		},
	}
	cmd.Flags().BoolVar(&keepRecord, "keep-record", false, "Release the reference but keep the record at Cloudflare")
	cmd.Flags().Bool("force", false, "Skip the confirmation")
	return cmd
}

func trimCreate(req models.DNSRecordCreateRequest) models.DNSRecordCreateRequest {
	return models.DNSRecordCreateRequest{
		Hostname:             strings.TrimSpace(req.Hostname),
		IntegrationTokenUUID: strings.TrimSpace(req.IntegrationTokenUUID),
		Address:              strings.TrimSpace(req.Address),
	}
}

func statusRows(resp *models.ResourceDNSStatus) []statusRow {
	rows := []statusRow{}
	for _, host := range resp.Hostnames {
		record := ""
		if len(host.Records) > 0 {
			r := host.Records[0]
			record = fmt.Sprintf("%s %s", r.Type, r.Address)
			if r.Owned {
				record += " (created by Coolify)"
			}
		}
		if len(host.Providers) == 0 {
			rows = append(rows, statusRow{Hostname: host.Hostname, Zone: "no provider", Record: record})
			continue
		}
		for _, provider := range host.Providers {
			rows = append(rows, statusRow{
				Hostname: host.Hostname, Zone: provider.Zone, IntegrationTokenUUID: provider.IntegrationTokenUUID,
				AutomaticDNS: fmt.Sprintf("%t", provider.AutomaticDNS), Record: record,
			})
		}
	}
	return rows
}

func newService(cmd *cobra.Command) (*service.DNSService, error) {
	client, err := cli.GetAPIClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get API client: %w", err)
	}
	return service.NewDNSService(client), nil
}

// format prints the API response for json and pretty, and the table view for table.
func format(cmd *cobra.Command, resp, table any) error {
	formatName, _ := cmd.Flags().GetString("format")
	formatter, err := output.NewFormatter(formatName, output.Options{Writer: cmd.OutOrStdout()})
	if err != nil {
		return err
	}
	if formatName != output.FormatTable {
		return formatter.Format(resp)
	}
	return formatter.Format(table)
}
