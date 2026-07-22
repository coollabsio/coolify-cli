package task

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	svcservice "github.com/coollabsio/coolify-cli/internal/service"
)

// NewCommand creates the service scheduled-task command tree.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "task",
		Aliases: []string{"tasks", "scheduled-task", "scheduled-tasks"},
		Short:   "Manage service scheduled tasks",
		Long:    `List, create, update, delete, and execute scheduled tasks for services.`,
	}
	cmd.AddCommand(
		newListCommand(),
		newCreateCommand(),
		newUpdateCommand(),
		newDeleteCommand(),
		newExecutionsCommand(),
		newExecuteCommand(),
	)
	return cmd
}

// NewExecuteTaskCommand is the top-level backward-compatible execute-task command.
func NewExecuteTaskCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "execute-task <service_uuid> <task_uuid>",
		Args:  cli.ExactArgs(2, "<service_uuid> <task_uuid>"),
		Short: "Execute a service scheduled task now",
		RunE:  runExecute,
	}
}

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <service_uuid>",
		Short: "List scheduled tasks for a service",
		Args:  cli.ExactArgs(1, "<service_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			tasks, err := svcservice.NewService(client).ListScheduledTasks(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return format(cmd, tasks)
		},
	}
}

func newCreateCommand() *cobra.Command {
	var name, command, frequency, container string
	var timeout int
	var enabled bool

	cmd := &cobra.Command{
		Use:   "create <service_uuid>",
		Short: "Create a scheduled task for a service",
		Long: `Create a scheduled task for a service.

Examples:
  coolify service task create <service_uuid> --name backup --command "pg_dump" --frequency "0 2 * * *"
  coolify service task create <service_uuid> --name cleanup --command "date" --frequency hourly --container app --enabled=false`,
		Args: cli.ExactArgs(1, "<service_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if command == "" {
				return fmt.Errorf("--command is required")
			}
			if frequency == "" {
				return fmt.Errorf("--frequency is required")
			}

			req := models.ScheduledTaskCreateRequest{
				Name:      name,
				Command:   command,
				Frequency: frequency,
			}
			if cmd.Flags().Changed("container") {
				req.Container = &container
			}
			if cmd.Flags().Changed("timeout") {
				req.Timeout = &timeout
			}
			if cmd.Flags().Changed("enabled") {
				req.Enabled = &enabled
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			task, err := svcservice.NewService(client).CreateScheduledTask(cmd.Context(), args[0], req)
			if err != nil {
				return err
			}
			return format(cmd, task)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Task name (required)")
	cmd.Flags().StringVar(&command, "command", "", "Command to run (required)")
	cmd.Flags().StringVar(&frequency, "frequency", "", "Cron expression or frequency (required)")
	cmd.Flags().StringVar(&container, "container", "", "Container name to run the command in")
	cmd.Flags().IntVar(&timeout, "timeout", 300, "Timeout in seconds")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Whether the task is enabled")

	return cmd
}

func newUpdateCommand() *cobra.Command {
	var name, command, frequency, container string
	var timeout int
	var enabled bool

	cmd := &cobra.Command{
		Use:   "update <service_uuid> <task_uuid>",
		Short: "Update a service scheduled task",
		Long: `Update a scheduled task. At least one field flag is required.

Examples:
  coolify service task update <service_uuid> <task_uuid> --name renamed
  coolify service task update <service_uuid> <task_uuid> --frequency "0 3 * * *" --enabled=false`,
		Args: cli.ExactArgs(2, "<service_uuid> <task_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			req := models.ScheduledTaskUpdateRequest{}
			hasUpdates := false

			if cmd.Flags().Changed("name") {
				req.Name = &name
				hasUpdates = true
			}
			if cmd.Flags().Changed("command") {
				req.Command = &command
				hasUpdates = true
			}
			if cmd.Flags().Changed("frequency") {
				req.Frequency = &frequency
				hasUpdates = true
			}
			if cmd.Flags().Changed("container") {
				req.Container = &container
				hasUpdates = true
			}
			if cmd.Flags().Changed("timeout") {
				req.Timeout = &timeout
				hasUpdates = true
			}
			if cmd.Flags().Changed("enabled") {
				req.Enabled = &enabled
				hasUpdates = true
			}
			if !hasUpdates {
				return fmt.Errorf("at least one field flag is required (--name, --command, --frequency, --container, --timeout, --enabled)")
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			task, err := svcservice.NewService(client).UpdateScheduledTask(cmd.Context(), args[0], args[1], req)
			if err != nil {
				return err
			}
			return format(cmd, task)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Task name")
	cmd.Flags().StringVar(&command, "command", "", "Command to run")
	cmd.Flags().StringVar(&frequency, "frequency", "", "Cron expression or frequency")
	cmd.Flags().StringVar(&container, "container", "", "Container name to run the command in")
	cmd.Flags().IntVar(&timeout, "timeout", 300, "Timeout in seconds")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Whether the task is enabled")

	return cmd
}

func newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <service_uuid> <task_uuid>",
		Short: "Delete a service scheduled task",
		Args:  cli.ExactArgs(2, "<service_uuid> <task_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			if err := svcservice.NewService(client).DeleteScheduledTask(cmd.Context(), args[0], args[1]); err != nil {
				return err
			}
			fmt.Println("Scheduled task deleted.")
			return nil
		},
	}
}

func newExecutionsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "executions <service_uuid> <task_uuid>",
		Short: "List executions for a service scheduled task",
		Args:  cli.ExactArgs(2, "<service_uuid> <task_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			executions, err := svcservice.NewService(client).ListScheduledTaskExecutions(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}
			return format(cmd, executions)
		},
	}
}

func newExecuteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "execute <service_uuid> <task_uuid>",
		Short: "Execute a service scheduled task now",
		Args:  cli.ExactArgs(2, "<service_uuid> <task_uuid>"),
		RunE:  runExecute,
	}
}

func runExecute(cmd *cobra.Command, args []string) error {
	client, err := cli.GetAPIClient(cmd)
	if err != nil {
		return fmt.Errorf("failed to get API client: %w", err)
	}
	resp, err := svcservice.NewService(client).ExecuteScheduledTask(cmd.Context(), args[0], args[1])
	if err != nil {
		return err
	}
	return format(cmd, resp)
}

func format(cmd *cobra.Command, value any) error {
	formatName, _ := cmd.Flags().GetString("format")
	if formatName == "table" {
		formatName = "pretty"
	}
	formatter, err := output.NewFormatter(formatName, output.Options{})
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}
	return formatter.Format(value)
}
