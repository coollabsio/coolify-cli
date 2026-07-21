package cloudinit

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewCloudInitCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "cloud-init", Aliases: []string{"cloudinit", "cloud-init-script"}, Short: "Manage cloud-init scripts"}
	cmd.AddCommand(newList(), newGet(), newCreate(), newUpdate(), newDelete())
	return cmd
}

func svc(cmd *cobra.Command) (*service.CloudInitService, error) {
	client, err := cli.GetAPIClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get API client: %w", err)
	}
	return service.NewCloudInitService(client), nil
}

func format(cmd *cobra.Command, value any) error {
	formatName, _ := cmd.Flags().GetString("format")
	showSensitive, _ := cmd.Flags().GetBool("show-sensitive")
	if !showSensitive {
		switch typed := value.(type) {
		case *models.CloudInitScript:
			redacted := *typed
			if typed.Script != nil {
				o := output.SensitiveOverlay
				redacted.Script = &o
			}
			value = &redacted
		case []models.CloudInitScript:
			redacted := append([]models.CloudInitScript(nil), typed...)
			for i := range redacted {
				if redacted[i].Script != nil {
					o := output.SensitiveOverlay
					redacted[i].Script = &o
				}
			}
			value = redacted
		}
	}
	formatter, err := output.NewFormatter(formatName, output.Options{ShowSensitive: showSensitive})
	if err != nil {
		return err
	}
	return formatter.Format(value)
}

func newList() *cobra.Command {
	return &cobra.Command{Use: "list", Short: "List cloud-init scripts", RunE: func(cmd *cobra.Command, _ []string) error {
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		items, err := s.List(cmd.Context())
		if err != nil {
			return err
		}
		return format(cmd, items)
	}}
}

func newGet() *cobra.Command {
	return &cobra.Command{Use: "get <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "Get a cloud-init script", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		item, err := s.Get(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return format(cmd, item)
	}}
}

func newCreate() *cobra.Command {
	var name, script, scriptFile string
	cmd := &cobra.Command{Use: "create", Short: "Create a cloud-init script", RunE: func(cmd *cobra.Command, _ []string) error {
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		body := script
		if scriptFile != "" {
			b, err := os.ReadFile(scriptFile)
			if err != nil {
				return fmt.Errorf("failed to read --script-file: %w", err)
			}
			body = string(b)
		}
		if body == "" {
			return fmt.Errorf("--script or --script-file is required")
		}
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		resp, err := s.Create(cmd.Context(), models.CloudInitScriptCreateRequest{Name: name, Script: body})
		if err != nil {
			return err
		}
		return format(cmd, resp)
	}}
	cmd.Flags().StringVar(&name, "name", "", "Script name")
	cmd.Flags().StringVar(&script, "script", "", "Script contents")
	cmd.Flags().StringVar(&scriptFile, "script-file", "", "Read script contents from file")
	return cmd
}

func newUpdate() *cobra.Command {
	var name, script, scriptFile string
	cmd := &cobra.Command{Use: "update <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "Update a cloud-init script", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		req := models.CloudInitScriptUpdateRequest{}
		if cmd.Flags().Changed("name") {
			req.Name = &name
		}
		if scriptFile != "" {
			b, err := os.ReadFile(scriptFile)
			if err != nil {
				return fmt.Errorf("failed to read --script-file: %w", err)
			}
			body := string(b)
			req.Script = &body
		} else if cmd.Flags().Changed("script") {
			req.Script = &script
		}
		resp, err := s.Update(cmd.Context(), args[0], req)
		if err != nil {
			return err
		}
		return format(cmd, resp)
	}}
	cmd.Flags().StringVar(&name, "name", "", "Script name")
	cmd.Flags().StringVar(&script, "script", "", "Script contents")
	cmd.Flags().StringVar(&scriptFile, "script-file", "", "Read script contents from file")
	return cmd
}

func newDelete() *cobra.Command {
	return &cobra.Command{Use: "delete <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "Delete a cloud-init script", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		if err := s.Delete(cmd.Context(), args[0]); err != nil {
			return err
		}
		fmt.Println("Cloud-init script deleted.")
		return nil
	}}
}
