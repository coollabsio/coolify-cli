package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
)

var docsCmd = &cobra.Command{
	Use:    "docs",
	Short:  "Generate documentation",
	Hidden: true,
}

var manCmd = &cobra.Command{
	Use:   "man",
	Short: "Generate man pages",
	Long: `Generate man pages for all Coolify CLI commands.

The man pages will be written to the specified directory (default: ./man).`,
	Example: `  coolify docs man
  coolify docs man --output-dir=/usr/local/share/man/man1`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		outputDir, _ := cmd.Flags().GetString("output-dir")

		// Create output directory if it doesn't exist
		if err := os.MkdirAll(outputDir, 0750); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Generate man pages
		header := &doc.GenManHeader{
			Title:   "COOLIFY",
			Section: "1",
			Source:  "Coolify CLI",
		}

		if err := doc.GenManTree(rootCmd, header, outputDir); err != nil {
			return fmt.Errorf("failed to generate man pages: %w", err)
		}

		absPath, _ := filepath.Abs(outputDir)
		fmt.Printf("Man pages generated successfully in: %s\n", absPath)
		fmt.Println("\nTo install the man pages system-wide:")
		fmt.Println("  sudo cp man/*.1 /usr/local/share/man/man1/")
		fmt.Println("  sudo mandb")
		fmt.Println("\nTo view a man page:")
		fmt.Println("  man coolify")
		fmt.Println("  man coolify-servers")

		return nil
	},
}

var markdownCmd = &cobra.Command{
	Use:     "markdown",
	Aliases: []string{"md"},

	Short: "Generate markdown documentation",
	Long: `Generate markdown documentation for all Coolify CLI commands.

The markdown files will be written to the specified directory (default: ./docs).`,
	Example: `  coolify docs markdown
  coolify docs markdown --output-dir=./documentation`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		outputDir, _ := cmd.Flags().GetString("output-dir")

		// Create output directory if it doesn't exist
		if err := os.MkdirAll(outputDir, 0750); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Generate markdown docs
		if err := doc.GenMarkdownTree(rootCmd, outputDir); err != nil {
			return fmt.Errorf("failed to generate markdown docs: %w", err)
		}

		absPath, _ := filepath.Abs(outputDir)
		fmt.Printf("Markdown documentation generated successfully in: %s\n", absPath)

		return nil
	},
}

var llmsCmd = &cobra.Command{
	Use:   "llms",
	Short: "Generate llms.txt for AI agent command specification",
	Long: `Generate a machine-readable llms.txt file that defines all CLI commands and their parameters.

This file is intended to enable AI agents to understand and interact with the CLI.
The output file will be written to the specified path (default: ./llms.txt).`,
	Example: `  coolify docs llms
  coolify docs llms --output=./llms.txt`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		outputFile, _ := cmd.Flags().GetString("output")

		var sb strings.Builder
		sb.WriteString(llmsHeader)
		writeLLMsCommand(&sb, rootCmd, "coolify")

		if err := os.WriteFile(outputFile, []byte(sb.String()), 0600); err != nil {
			return fmt.Errorf("failed to write llms.txt: %w", err)
		}

		absPath, _ := filepath.Abs(outputFile)
		fmt.Printf("llms.txt generated successfully: %s\n", absPath)

		return nil
	},
}

// llmsHeader contains the static overview section prepended to the generated command reference.
const llmsHeader = `# Coolify CLI - llms.txt

> A CLI tool for interacting with the Coolify API, built with Go.
> Manage Coolify instances (cloud and self-hosted), servers, projects, applications, databases, services, deployments, domains, and private keys.
> Source: https://github.com/coollabsio/coolify-cli
> API Spec: https://github.com/coollabsio/coolify/blob/v4.x/openapi.json

## Installation

` + "```bash" + `
# Linux/macOS (recommended)
curl -fsSL https://raw.githubusercontent.com/coollabsio/coolify-cli/main/scripts/install.sh | bash

# Homebrew (macOS/Linux)
brew install coollabsio/coolify-cli/coolify-cli

# Windows (PowerShell)
irm https://raw.githubusercontent.com/coollabsio/coolify-cli/main/scripts/install.ps1 | iex

# Go install
go install github.com/coollabsio/coolify-cli/coolify@latest
` + "```" + `

## Authentication

1. Get an API token from your Coolify dashboard at ` + "`/security/api-tokens`" + `
2. For Coolify Cloud: ` + "`coolify context set-token cloud <token>`" + `
3. For self-hosted: ` + "`coolify context add -d <context_name> <url> <token>`" + `

## Configuration

Config file location:
- Linux/macOS: ` + "`~/.config/coolify/config.json`" + `
- Windows: ` + "`%APPDATA%\\coolify\\config.json`" + `

Supports multiple contexts (instances) with ` + "`coolify context`" + ` commands.

## Output Formats

All commands support ` + "`--format`" + ` flag:
- ` + "`table`" + ` (default) - human-readable tabular output
- ` + "`json`" + ` - compact JSON for scripting
- ` + "`pretty`" + ` - indented JSON for debugging

## Command Aliases

Commands accept multiple aliases for convenience:
- ` + "`app` | `apps` | `application` | `applications`" + `
- ` + "`database` | `databases` | `db` | `dbs`" + `
- ` + "`service` | `services` | `svc`" + `
- ` + "`server` | `servers`" + `
- ` + "`project` | `projects`" + `
- ` + "`resource` | `resources`" + `
- ` + "`private-key` | `private-keys` | `key` | `keys`" + `
- ` + "`teams` | `team`" + `
- ` + "`teams members` | `teams member`" + `
- ` + "`github` | `gh` | `github-app` | `github-apps`" + `
- ` + "`app start`" + ` also aliased as ` + "`app deploy`" + `
- ` + "`server domains`" + ` also aliased as ` + "`server domain`" + `

## Supported Database Types

When using ` + "`coolify database create <type>`" + `:
- ` + "`postgresql`" + `
- ` + "`mysql`" + `
- ` + "`mariadb`" + `
- ` + "`mongodb`" + `
- ` + "`redis`" + `
- ` + "`keydb`" + `
- ` + "`clickhouse`" + `
- ` + "`dragonfly`" + `

## Usage Examples

` + "```bash" + `
# Multi-context workflow
coolify context add prod https://prod.coolify.io <token>
coolify context add staging https://staging.coolify.io <token>
coolify context use prod
coolify --context=staging server list

# Application lifecycle
coolify app list
coolify app get <uuid>
coolify app start <uuid>
coolify app stop <uuid>
coolify app restart <uuid>
coolify app logs <uuid> --follow

# Environment variable management
coolify app env list <uuid>
coolify app env create <uuid> --key API_KEY --value secret123
coolify app env sync <uuid> --file .env.production --build-time --preview

# Deploy workflows
coolify deploy name my-application
coolify deploy batch api,worker,frontend --force
coolify deploy list
coolify deploy cancel <uuid>

# Database backup
coolify database backup create <db-uuid> --frequency "0 2 * * *" --enabled --save-s3
coolify database backup trigger <db-uuid> <backup-uuid>

# Application creation
coolify app create public --project-uuid <uuid> --server-uuid <uuid> --git-repository https://github.com/user/repo --git-branch main --build-pack nixpacks --ports-exposes 3000
coolify app create dockerfile --project-uuid <uuid> --server-uuid <uuid> --dockerfile "FROM node:18\nCOPY . .\nRUN npm install\nCMD [\"node\", \"index.js\"]"
coolify app create dockerimage --project-uuid <uuid> --server-uuid <uuid> --docker-registry-image-name nginx --ports-exposes 80

# Service creation (one-click services)
coolify service create <type> --project-uuid <uuid> --server-uuid <uuid> --instant-deploy
coolify service create --list-types  # list all available service types

# Storage management
coolify app storage create <app-uuid> --type persistent --mount-path /data --name my-volume
coolify app storage create <app-uuid> --type file --mount-path /app/config.yml --content "key: value"

# GitHub App integration
coolify github list
coolify github repos <app-uuid>
coolify github branches <app-uuid> owner/repo

# Team management
coolify team list
coolify team current
coolify team members list
` + "```" + `

## API Notes

- All resource identifiers use UUIDs (not internal database IDs)
- API base path: ` + "`/api/v1/`" + `
- Authentication: Bearer token via ` + "`--token`" + ` flag or context configuration
- ` + "`app env sync`" + ` behavior: updates existing variables, creates missing ones, does NOT delete variables not in the file
- ` + "`app start`" + ` aliases to ` + "`app deploy`" + ` and also accepts ` + "`--force`" + ` and ` + "`--instant-deploy`" + ` flags
- Deployment logs support ` + "`--follow`" + ` for real-time streaming and ` + "`--debuglogs`" + ` for internal operations
- ` + "`app logs`" + ` defaults to 100 lines; ` + "`app deployments logs`" + ` defaults to 0 (all lines)
- Short flag ` + "`-n`" + ` can be used instead of ` + "`--lines`" + ` for log commands
- ` + "`completion`" + ` command supports shells: ` + "`bash`" + `, ` + "`zsh`" + `, ` + "`fish`" + `, ` + "`powershell`" + `
- Resource statuses: ` + "`running`" + `, ` + "`stopped`" + `, ` + "`error`" + `
- Teams use numeric IDs (not UUIDs) - this is the only resource that uses IDs
- Fields marked ` + "`sensitive:\"true\"`" + ` (tokens, passwords, IPs, emails) are hidden by default; use ` + "`--show-sensitive`" + ` to reveal

## Data Models (JSON Response Fields)

### Application
Table columns: uuid, name, description, status, fqdn, git_repository, git_branch, build_pack, ports_exposes
JSON-only fields: git_commit_sha, git_full_url, install_command, build_command, start_command, base_directory, publish_directory, static_image, dockerfile, dockerfile_location, docker_registry_image_name, docker_registry_image_tag, docker_compose, ports_mappings, domains, redirect, preview_url_template, health_check_enabled, health_check_path, health_check_port, health_check_host, health_check_method, health_check_scheme, health_check_return_code, health_check_response_text, health_check_interval, health_check_timeout, health_check_retries, health_check_start_period, limits_cpus, limits_cpu_shares, limits_cpuset, limits_memory, limits_memory_reservation, limits_memory_swap, limits_memory_swappiness, pre_deployment_command, post_deployment_command, watch_paths, swarm_replicas, config_hash, settings (nested: is_static, is_build_server_enabled, is_auto_deploy_enabled, is_force_https_enabled, is_debug_enabled, is_preview_deployments_enabled, is_git_submodules_enabled, is_git_lfs_enabled)

### Database
Table columns: uuid, name, description, image, status, type, is_public, public_port
Supported types: postgresql, mysql, mariadb, mongodb, redis, keydb, clickhouse, dragonfly
JSON-only fields: limits_memory, limits_cpus, and database-specific fields (postgres_user, postgres_password, postgres_db, mysql_root_password, mysql_user, mysql_database, mariadb_root_password, mariadb_user, mariadb_database, mongo_initdb_root_username, mongo_initdb_root_password, etc.)

### Service
Table columns: uuid, name, description, status
JSON-only fields: docker_compose, docker_compose_raw
Nested resources: applications (uuid, name, status, fqdn), databases (uuid, name, type, status)

### Server
Table columns: uuid, name, ip (sensitive), user (sensitive), port (sensitive)
JSON-only fields: settings (is_reachable, is_usable)

### Project
Table columns: uuid, name, description
Nested: environments (uuid, name, description, applications)

### Deployment
Table columns: deployment_uuid, application_name, server_name, status, commit
JSON-only fields: commit_message, deployment_url, finished_at, logs, created_at

### Resource
Table columns: uuid, name, type, status

### Private Key
Table columns: uuid, name, public_key (sensitive), private_key (sensitive)

### Team
Table columns: id, name, description, personal_team, show_boarding
JSON-only fields: custom_server_limit, created_at

### Team Member
Table columns: id, name, email (sensitive), role, force_password_reset, marketing_emails

### GitHub App
Table columns: uuid, name, organization, api_url, html_url, custom_user, custom_port
JSON-only fields: app_id, installation_id, client_id, private_key_id, is_system_wide, team_id

### Environment Variable (Application)
Fields: uuid, key, value (sensitive), is_buildtime, is_preview, is_literal, is_shown_once, is_runtime, is_shared, comment, real_value (sensitive)

### Environment Variable (Service/Database)
Same as application but without is_preview field

### Storage (Persistent Volume)
Fields: uuid, name, mount_path, host_path, is_preview_suffix_enabled, is_readonly

### Storage (File)
Fields: uuid, fs_path, mount_path, content, is_directory, is_based_on_git, is_preview_suffix_enabled, chown, chmod

---

## Command Reference

`

// writeLLMsCommand recursively writes command documentation in llms.txt format.
func writeLLMsCommand(sb *strings.Builder, cmd *cobra.Command, parentPath string) {
	// Build the full command path including args from Use field
	commandPath := parentPath
	if cmd.HasParent() {
		parts := strings.Fields(cmd.Use)
		commandPath = parentPath + " " + parts[0]
		// Append positional args from the Use field (e.g., "<uuid>", "[optional]")
		if len(parts) > 1 {
			commandPath += " " + strings.Join(parts[1:], " ")
		}
	}

	// Skip the docs command itself and help command
	if cmd.Name() == "docs" || cmd.Name() == "help" {
		return
	}

	// Determine if this command should be written
	isRoot := !cmd.HasParent()
	isRunnable := cmd.RunE != nil || cmd.Run != nil
	hasVisibleChildren := false
	for _, child := range cmd.Commands() {
		if !child.Hidden && child.Name() != "help" {
			hasVisibleChildren = true
			break
		}
	}

	// Write the root command, runnable commands, and leaf commands (no children)
	if isRoot || isRunnable || !hasVisibleChildren {
		// Get description - prefer Long if it's a single clean sentence, otherwise use Short
		description := cmd.Short
		if cmd.Long != "" {
			longLines := strings.Split(strings.TrimSpace(cmd.Long), "\n")
			if len(longLines) == 1 && len(longLines[0]) < 200 {
				description = longLines[0]
			}
		}

		fmt.Fprintf(sb, "Command: %s\n", commandPath)
		fmt.Fprintf(sb, "Description: %s\n", description)

		// For root command, show persistent flags; for others, show local flags
		var flags []*pflag.Flag
		if isRoot {
			cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
				if f.Name == "help" {
					return
				}
				flags = append(flags, f)
			})
		} else {
			cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
				if f.Name == "help" {
					return
				}
				flags = append(flags, f)
			})
		}

		if len(flags) == 0 {
			sb.WriteString("Parameters: (None)\n")
		} else {
			sb.WriteString("Parameters:\n")
			for _, f := range flags {
				flagType := f.Value.Type()
				// Normalize type names
				switch flagType {
				case "int", "int32", "int64":
					flagType = "integer"
				case "bool":
					flagType = "boolean"
				}

				// Check if the flag is marked as required via cobra annotation
				// or via "(required)" in the usage string
				required := isFlagRequired(f)

				if f.Shorthand != "" {
					fmt.Fprintf(sb, "  - name: --%s (-%s)\n", f.Name, f.Shorthand)
				} else {
					fmt.Fprintf(sb, "  - name: --%s\n", f.Name)
				}
				fmt.Fprintf(sb, "    type: %s\n", flagType)
				fmt.Fprintf(sb, "    description: %s\n", f.Usage)
				fmt.Fprintf(sb, "    required: %t\n", required)
				if f.DefValue != "" && f.DefValue != "false" && f.DefValue != "0" && f.DefValue != "[]" {
					fmt.Fprintf(sb, "    default: %s\n", f.DefValue)
				}
			}
		}

		sb.WriteString("\n")
	}

	// Recurse into subcommands
	for _, child := range cmd.Commands() {
		if child.Hidden || child.Name() == "help" {
			continue
		}
		childPath := parentPath
		if cmd.HasParent() {
			parts := strings.Fields(cmd.Use)
			childPath = parentPath + " " + parts[0]
		}
		writeLLMsCommand(sb, child, childPath)
	}
}

// isFlagRequired checks if a flag is required by looking at cobra annotations
// and the "(required)" convention in usage strings.
func isFlagRequired(f *pflag.Flag) bool {
	// Check cobra's MarkFlagRequired annotation
	if ann, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok && len(ann) > 0 && ann[0] == "true" {
		return true
	}
	// Check for "(required)" in usage string (convention used in this codebase)
	return strings.Contains(strings.ToLower(f.Usage), "(required)")
}

func NewDocsCommand() *cobra.Command {
	docsCmd.AddCommand(manCmd)
	docsCmd.AddCommand(markdownCmd)
	docsCmd.AddCommand(llmsCmd)

	manCmd.Flags().StringP("output-dir", "o", "./man", "Output directory for man pages")
	markdownCmd.Flags().StringP("output-dir", "o", "./docs", "Output directory for markdown files")
	llmsCmd.Flags().StringP("output", "o", "./llms.txt", "Output file path")

	return docsCmd
}
