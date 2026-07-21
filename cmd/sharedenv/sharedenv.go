package sharedenv

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewSharedEnvCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "shared-env",
		Aliases: []string{"shared-envs", "sharedenv"},
		Short:   "Manage hierarchical shared environment variables",
		Long:    "Team, project, environment, and server shared env vars (inherited by resources).",
	}
	cmd.AddCommand(newTeamCmd(), newProjectCmd(), newEnvironmentCmd(), newServerCmd())
	return cmd
}

func svc(cmd *cobra.Command) (*service.SharedEnvService, error) {
	client, err := cli.GetAPIClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get API client: %w", err)
	}
	return service.NewSharedEnvService(client), nil
}

func formatEnvs(cmd *cobra.Command, envs []models.SharedEnvironmentVariable) error {
	showSensitive, _ := cmd.Flags().GetBool("show-sensitive")
	if !showSensitive {
		for i := range envs {
			if envs[i].Value != nil {
				masked := "********"
				envs[i].Value = &masked
			}
		}
	}
	formatName, _ := cmd.Flags().GetString("format")
	formatter, err := output.NewFormatter(formatName, output.Options{ShowSensitive: showSensitive})
	if err != nil {
		return err
	}
	return formatter.Format(envs)
}

func formatOne(cmd *cobra.Command, env *models.SharedEnvironmentVariable) error {
	return formatEnvs(cmd, []models.SharedEnvironmentVariable{*env})
}

func createFlags(cmd *cobra.Command) (key, value, comment *string, isLiteral, isMultiline, isShownOnce *bool) {
	key = cmd.Flags().String("key", "", "Variable key")
	value = cmd.Flags().String("value", "", "Variable value")
	comment = cmd.Flags().String("comment", "", "Comment")
	isLiteral = cmd.Flags().Bool("literal", false, "Treat value as literal")
	isMultiline = cmd.Flags().Bool("multiline", false, "Multiline value")
	isShownOnce = cmd.Flags().Bool("shown-once", false, "Show value only once")
	return
}

func buildCreate(cmd *cobra.Command, key, value, comment string, isLiteral, isMultiline, isShownOnce bool) (models.SharedEnvCreateRequest, error) {
	if key == "" || value == "" {
		return models.SharedEnvCreateRequest{}, fmt.Errorf("--key and --value are required")
	}
	req := models.SharedEnvCreateRequest{Key: key, Value: value}
	if cmd.Flags().Changed("literal") {
		req.IsLiteral = &isLiteral
	}
	if cmd.Flags().Changed("multiline") {
		req.IsMultiline = &isMultiline
	}
	if cmd.Flags().Changed("shown-once") {
		req.IsShownOnce = &isShownOnce
	}
	if cmd.Flags().Changed("comment") {
		req.Comment = &comment
	}
	return req, nil
}

func buildUpdate(cmd *cobra.Command) models.SharedEnvUpdateRequest {
	req := models.SharedEnvUpdateRequest{}
	if cmd.Flags().Changed("key") {
		v, _ := cmd.Flags().GetString("key")
		req.Key = &v
	}
	if cmd.Flags().Changed("value") {
		v, _ := cmd.Flags().GetString("value")
		req.Value = &v
	}
	if cmd.Flags().Changed("comment") {
		v, _ := cmd.Flags().GetString("comment")
		req.Comment = &v
	}
	if cmd.Flags().Changed("literal") {
		v, _ := cmd.Flags().GetBool("literal")
		req.IsLiteral = &v
	}
	if cmd.Flags().Changed("multiline") {
		v, _ := cmd.Flags().GetBool("multiline")
		req.IsMultiline = &v
	}
	if cmd.Flags().Changed("shown-once") {
		v, _ := cmd.Flags().GetBool("shown-once")
		req.IsShownOnce = &v
	}
	return req
}

func newTeamCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "team", Short: "Team-level shared envs (/team/envs)"}
	cmd.AddCommand(
		&cobra.Command{Use: "list", Short: "List team shared envs", RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := svc(cmd)
			if err != nil {
				return err
			}
			items, err := s.ListTeam(cmd.Context())
			if err != nil {
				return err
			}
			return formatEnvs(cmd, items)
		}},
	)
	create := &cobra.Command{Use: "create", Short: "Create team shared env", RunE: func(cmd *cobra.Command, _ []string) error {
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")
		comment, _ := cmd.Flags().GetString("comment")
		isLiteral, _ := cmd.Flags().GetBool("literal")
		isMultiline, _ := cmd.Flags().GetBool("multiline")
		isShownOnce, _ := cmd.Flags().GetBool("shown-once")
		req, err := buildCreate(cmd, key, value, comment, isLiteral, isMultiline, isShownOnce)
		if err != nil {
			return err
		}
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		resp, err := s.CreateTeam(cmd.Context(), req)
		if err != nil {
			return err
		}
		fmt.Printf("Created shared env id=%d\n", resp.ID)
		return nil
	}}
	createFlags(create)
	cmd.AddCommand(create)

	update := &cobra.Command{Use: "update <id>", Args: cli.ExactArgs(1, "<id>"), Short: "Update team shared env", RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("id must be an integer")
		}
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		env, err := s.UpdateTeam(cmd.Context(), id, buildUpdate(cmd))
		if err != nil {
			return err
		}
		return formatOne(cmd, env)
	}}
	createFlags(update)
	cmd.AddCommand(update)

	cmd.AddCommand(&cobra.Command{Use: "delete <id>", Args: cli.ExactArgs(1, "<id>"), Short: "Delete team shared env", RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("id must be an integer")
		}
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		if err := s.DeleteTeam(cmd.Context(), id); err != nil {
			return err
		}
		fmt.Println("Shared env deleted.")
		return nil
	}})
	return cmd
}

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "project", Short: "Project-level shared envs"}
	cmd.AddCommand(&cobra.Command{Use: "list <project_uuid>", Args: cli.ExactArgs(1, "<project_uuid>"), Short: "List project shared envs", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		items, err := s.ListProject(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return formatEnvs(cmd, items)
	}})
	create := &cobra.Command{Use: "create <project_uuid>", Args: cli.ExactArgs(1, "<project_uuid>"), Short: "Create project shared env", RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")
		comment, _ := cmd.Flags().GetString("comment")
		isLiteral, _ := cmd.Flags().GetBool("literal")
		isMultiline, _ := cmd.Flags().GetBool("multiline")
		isShownOnce, _ := cmd.Flags().GetBool("shown-once")
		req, err := buildCreate(cmd, key, value, comment, isLiteral, isMultiline, isShownOnce)
		if err != nil {
			return err
		}
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		resp, err := s.CreateProject(cmd.Context(), args[0], req)
		if err != nil {
			return err
		}
		fmt.Printf("Created shared env id=%d\n", resp.ID)
		return nil
	}}
	createFlags(create)
	cmd.AddCommand(create)
	update := &cobra.Command{Use: "update <project_uuid> <id>", Args: cli.ExactArgs(2, "<project_uuid> <id>"), Short: "Update project shared env", RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("id must be an integer")
		}
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		env, err := s.UpdateProject(cmd.Context(), args[0], id, buildUpdate(cmd))
		if err != nil {
			return err
		}
		return formatOne(cmd, env)
	}}
	createFlags(update)
	cmd.AddCommand(update)
	cmd.AddCommand(&cobra.Command{Use: "delete <project_uuid> <id>", Args: cli.ExactArgs(2, "<project_uuid> <id>"), Short: "Delete project shared env", RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("id must be an integer")
		}
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		if err := s.DeleteProject(cmd.Context(), args[0], id); err != nil {
			return err
		}
		fmt.Println("Shared env deleted.")
		return nil
	}})
	return cmd
}

func newEnvironmentCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "environment", Aliases: []string{"env"}, Short: "Environment-level shared envs"}
	cmd.AddCommand(&cobra.Command{Use: "list <project_uuid> <environment>", Args: cli.ExactArgs(2, "<project_uuid> <environment>"), Short: "List environment shared envs", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		items, err := s.ListEnvironment(cmd.Context(), args[0], args[1])
		if err != nil {
			return err
		}
		return formatEnvs(cmd, items)
	}})
	create := &cobra.Command{Use: "create <project_uuid> <environment>", Args: cli.ExactArgs(2, "<project_uuid> <environment>"), Short: "Create environment shared env", RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")
		comment, _ := cmd.Flags().GetString("comment")
		isLiteral, _ := cmd.Flags().GetBool("literal")
		isMultiline, _ := cmd.Flags().GetBool("multiline")
		isShownOnce, _ := cmd.Flags().GetBool("shown-once")
		req, err := buildCreate(cmd, key, value, comment, isLiteral, isMultiline, isShownOnce)
		if err != nil {
			return err
		}
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		resp, err := s.CreateEnvironment(cmd.Context(), args[0], args[1], req)
		if err != nil {
			return err
		}
		fmt.Printf("Created shared env id=%d\n", resp.ID)
		return nil
	}}
	createFlags(create)
	cmd.AddCommand(create)
	update := &cobra.Command{Use: "update <project_uuid> <environment> <id>", Args: cli.ExactArgs(3, "<project_uuid> <environment> <id>"), Short: "Update environment shared env", RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("id must be an integer")
		}
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		env, err := s.UpdateEnvironment(cmd.Context(), args[0], args[1], id, buildUpdate(cmd))
		if err != nil {
			return err
		}
		return formatOne(cmd, env)
	}}
	createFlags(update)
	cmd.AddCommand(update)
	cmd.AddCommand(&cobra.Command{Use: "delete <project_uuid> <environment> <id>", Args: cli.ExactArgs(3, "<project_uuid> <environment> <id>"), Short: "Delete environment shared env", RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("id must be an integer")
		}
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		if err := s.DeleteEnvironment(cmd.Context(), args[0], args[1], id); err != nil {
			return err
		}
		fmt.Println("Shared env deleted.")
		return nil
	}})
	return cmd
}

func newServerCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "server", Short: "Server-level shared envs"}
	cmd.AddCommand(&cobra.Command{Use: "list <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "List server shared envs", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		items, err := s.ListServer(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return formatEnvs(cmd, items)
	}})
	create := &cobra.Command{Use: "create <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "Create server shared env", RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")
		comment, _ := cmd.Flags().GetString("comment")
		isLiteral, _ := cmd.Flags().GetBool("literal")
		isMultiline, _ := cmd.Flags().GetBool("multiline")
		isShownOnce, _ := cmd.Flags().GetBool("shown-once")
		req, err := buildCreate(cmd, key, value, comment, isLiteral, isMultiline, isShownOnce)
		if err != nil {
			return err
		}
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		resp, err := s.CreateServer(cmd.Context(), args[0], req)
		if err != nil {
			return err
		}
		fmt.Printf("Created shared env id=%d\n", resp.ID)
		return nil
	}}
	createFlags(create)
	cmd.AddCommand(create)
	update := &cobra.Command{Use: "update <server_uuid> <id>", Args: cli.ExactArgs(2, "<server_uuid> <id>"), Short: "Update server shared env", RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("id must be an integer")
		}
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		env, err := s.UpdateServer(cmd.Context(), args[0], id, buildUpdate(cmd))
		if err != nil {
			return err
		}
		return formatOne(cmd, env)
	}}
	createFlags(update)
	cmd.AddCommand(update)
	cmd.AddCommand(&cobra.Command{Use: "delete <server_uuid> <id>", Args: cli.ExactArgs(2, "<server_uuid> <id>"), Short: "Delete server shared env", RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("id must be an integer")
		}
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		if err := s.DeleteServer(cmd.Context(), args[0], id); err != nil {
			return err
		}
		fmt.Println("Shared env deleted.")
		return nil
	}})
	return cmd
}
