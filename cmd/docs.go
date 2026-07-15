package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
	Short: "Generate llms.txt and llms-full.txt for AI agents",
	Long: `Generate AI-friendly documentation files for the Coolify CLI.

This creates a concise llms.txt quick reference plus a complete llms-full.txt command catalog.
The output files will be written to the specified paths (defaults: ./llms.txt and ./llms-full.txt).`,
	Example: `  coolify docs llms
  coolify docs llms --output=./llms.txt --full-output=./llms-full.txt`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		outputFile, _ := cmd.Flags().GetString("output")
		fullOutputFile, _ := cmd.Flags().GetString("full-output")

		return writeLLMsArtifacts(outputFile, fullOutputFile)
	},
}

const llmsQuickTemplate = `# Coolify CLI - llms.txt

> Quick AI/LLM instructions for the Coolify CLI.
> Source: https://github.com/coollabsio/coolify-cli
> API Spec: https://github.com/coollabsio/coolify/blob/v4.x/openapi.json

## Operating Rules

- Prefer ` + "`--format json`" + ` for automation and parsing.
- Use Coolify UUIDs for resources; do not use internal numeric IDs.
- Team commands are the exception: they use numeric team IDs.
- Authenticate with a saved context when possible; use ` + "`--token`" + ` only for overrides.
- Use ` + "`llms-full.txt`" + ` for the exhaustive command/flag catalog.

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
4. Switch contexts with ` + "`coolify context use <context_name>`" + `

## Configuration

Config file location:
- Linux/macOS: ` + "`~/.config/coolify/config.json`" + `
- Windows: ` + "`%%APPDATA%%\\coolify\\config.json`" + `

Supports multiple contexts (instances) with ` + "`coolify context`" + ` commands.

## Output Formats

All commands support ` + "`--format`" + ` flag:
- ` + "`table`" + ` (default) - human-readable tabular output
- ` + "`json`" + ` - compact JSON for scripting
- ` + "`pretty`" + ` - indented JSON for debugging

## Global Flags

- ` + "`--context <name>`" + ` - use a specific saved context
- ` + "`--token <token>`" + ` - override token from config
- ` + "`--format table|json|pretty`" + ` - choose output format
- ` + "`--show-sensitive`" + ` - reveal sensitive values
- ` + "`--debug`" + ` - enable debug output

## Common Workflows

### Contexts

` + "```bash" + `
coolify context list
coolify context verify
coolify context version
coolify context use prod
` + "```" + `

### Inventory

` + "```bash" + `
coolify server list
coolify project list
coolify resource list
coolify app list
coolify service list
coolify database list
` + "```" + `

### Applications

` + "```bash" + `
coolify app get <uuid>
coolify app start <uuid>
coolify app stop <uuid>
coolify app restart <uuid>
coolify app logs <uuid> --show-timestamps
coolify app move <uuid> --environment-uuid <uuid>
coolify app tag add <uuid> production
coolify app deployments list <app-uuid>
coolify app deployments logs <app-uuid> --follow
` + "```" + `

### Environment Variables

` + "```bash" + `
coolify app env list <app-uuid>
coolify app env create <app-uuid> --key API_KEY --value secret123
coolify app env update <app-uuid> <env-uuid-or-key> --value new-secret
coolify app env sync <app-uuid> --file .env.production --build-time --preview
` + "```" + `

### Deployments

` + "```bash" + `
coolify deploy list
coolify deploy name my-application
coolify deploy batch api,worker,frontend --force
coolify deploy cancel <deployment-uuid>
` + "```" + `

### Databases and Services

` + "```bash" + `
coolify database get <uuid>
coolify database create postgresql --server-uuid <uuid> --project-uuid <uuid> --environment-name production
coolify database logs <uuid> --show-timestamps
coolify database move <uuid> --environment-uuid <uuid>
coolify database backup list <database-uuid>
coolify service get <uuid>
coolify service create <type> --project-uuid <uuid> --server-uuid <uuid> --instant-deploy
coolify service logs <uuid> --sub-service-name <name> --show-timestamps
coolify service application list <service-uuid>
coolify service application restart <service-uuid> <application-uuid>
coolify service database list <service-uuid>
coolify service database restart <service-uuid> <database-uuid>
` + "```" + `

### Tags, Destinations, and Cloud Providers

` + "```bash" + `
coolify tag list
coolify destination list --server <server-uuid>
coolify cloud-token create --provider hetzner --name production --provider-token <token>
coolify cloud-token validate <uuid>
coolify server hetzner locations <cloud-token-uuid>
coolify server hetzner create --cloud-token <uuid> --location <location> --server-type <type> --image <id>
` + "```" + `

## Common Aliases

- ` + "`coolify app`" + ` | ` + "`coolify apps`" + ` | ` + "`coolify application`" + ` | ` + "`coolify applications`" + `
- ` + "`coolify service`" + ` | ` + "`coolify services`" + ` | ` + "`coolify svc`" + `
- ` + "`coolify database`" + ` | ` + "`coolify databases`" + ` | ` + "`coolify db`" + ` | ` + "`coolify dbs`" + `
- ` + "`coolify teams`" + ` | ` + "`coolify team`" + `

## Full Reference

- Full command and parameter catalog: %s
- Regenerate docs: ` + "`go run ./coolify docs llms`" + `
`

const llmsFullIntro = `# Coolify CLI - llms-full.txt

> Full AI/LLM command catalog for the Coolify CLI.
> Manage Coolify instances (cloud and self-hosted), servers, projects, applications, databases, services, deployments, domains, and private keys.
> Source: https://github.com/coollabsio/coolify-cli
> API Spec: https://github.com/coollabsio/coolify/blob/v4.x/openapi.json

## Companion Files

- Quick instructions: %s
- Regenerate docs: ` + "`go run ./coolify docs llms`" + `

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
- Windows: ` + "`%%APPDATA%%\\coolify\\config.json`" + `

Supports multiple contexts (instances) with ` + "`coolify context`" + ` commands.

## Output Formats

All commands support ` + "`--format`" + ` flag:
- ` + "`table`" + ` (default) - human-readable tabular output
- ` + "`json`" + ` - compact JSON for scripting
- ` + "`pretty`" + ` - indented JSON for debugging
`

const llmsFullBody = `

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

---

## Command Reference

`

func buildQuickLLMSText(fullReferencePath string) string {
	return fmt.Sprintf(llmsQuickTemplate, fullReferencePath)
}

func buildFullLLMSText(quickReferencePath string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, llmsFullIntro, quickReferencePath)
	writeLLMsAliases(&sb, rootCmd, "coolify")
	sb.WriteString(llmsFullBody)
	writeLLMsCommand(&sb, rootCmd, "coolify")
	return sb.String()
}

func writeLLMsArtifacts(outputFile, fullOutputFile string) error {
	if filepath.Clean(outputFile) == filepath.Clean(fullOutputFile) {
		return fmt.Errorf("output and full-output must be different files")
	}

	if err := ensureParentDir(outputFile); err != nil {
		return err
	}
	if err := ensureParentDir(fullOutputFile); err != nil {
		return err
	}

	quickReferencePath := llmsReferencePath(fullOutputFile, outputFile)
	fullReferencePath := llmsReferencePath(outputFile, fullOutputFile)

	if err := os.WriteFile(outputFile, []byte(buildQuickLLMSText(fullReferencePath)), 0600); err != nil {
		return fmt.Errorf("failed to write llms.txt: %w", err)
	}
	if err := os.WriteFile(fullOutputFile, []byte(buildFullLLMSText(quickReferencePath)), 0600); err != nil {
		return fmt.Errorf("failed to write llms-full.txt: %w", err)
	}

	absQuickPath, _ := filepath.Abs(outputFile)
	absFullPath, _ := filepath.Abs(fullOutputFile)
	fmt.Printf("llms.txt generated successfully: %s\n", absQuickPath)
	fmt.Printf("llms-full.txt generated successfully: %s\n", absFullPath)

	return nil
}

func ensureParentDir(path string) error {
	parentDir := filepath.Dir(path)
	if err := os.MkdirAll(parentDir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	return nil
}

func llmsReferencePath(fromFile, toFile string) string {
	referencePath, err := filepath.Rel(filepath.Dir(fromFile), toFile)
	if err != nil {
		return filepath.ToSlash(toFile)
	}
	referencePath = filepath.ToSlash(referencePath)
	if strings.HasPrefix(referencePath, ".") || strings.HasPrefix(referencePath, "/") {
		return referencePath
	}
	return "./" + referencePath
}

// writeLLMsAliases writes aliases derived from the Cobra command tree.
func writeLLMsAliases(sb *strings.Builder, cmd *cobra.Command, parentPath string) {
	aliases := collectLLMsAliases(cmd, parentPath)
	if len(aliases) == 0 {
		return
	}

	sb.WriteString("\n## Command Aliases\n\n")
	sb.WriteString("Aliases are derived from the CLI command tree:\n")
	for _, aliasLine := range aliases {
		fmt.Fprintf(sb, "- %s\n", aliasLine)
	}
}

func collectLLMsAliases(cmd *cobra.Command, parentPath string) []string {
	var aliases []string
	if cmd.Name() != "docs" && cmd.Name() != "help" {
		if len(cmd.Aliases) > 0 {
			aliasNames := append([]string{cmd.Name()}, cmd.Aliases...)
			for i := range aliasNames {
				aliasNames[i] = fmt.Sprintf("`%s`", commandPathPrefix(parentPath, cmd)+aliasNames[i])
			}
			aliases = append(aliases, strings.Join(aliasNames, " | "))
		}
	}

	for _, child := range cmd.Commands() {
		if child.Hidden || child.Name() == "help" {
			continue
		}
		aliases = append(aliases, collectLLMsAliases(child, llmsCommandName(parentPath, cmd))...)
	}

	slices.Sort(aliases)
	return slices.Compact(aliases)
}

func llmsCommandName(parentPath string, cmd *cobra.Command) string {
	if !cmd.HasParent() {
		return parentPath
	}

	parts := strings.Fields(cmd.Use)
	commandPath := parentPath + " " + parts[0]
	if len(parts) > 1 {
		commandPath += " " + strings.Join(parts[1:], " ")
	}
	return commandPath
}

func commandPathPrefix(parentPath string, cmd *cobra.Command) string {
	if cmd.HasParent() {
		return parentPath + " "
	}
	return ""
}

// writeLLMsCommand recursively writes command documentation in llms.txt format.
func writeLLMsCommand(sb *strings.Builder, cmd *cobra.Command, parentPath string) {
	// Build the full command path including args from Use field
	commandPath := llmsCommandName(parentPath, cmd)

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
				if f.DefValue != "" && f.DefValue != "[]" {
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
			childPath = llmsCommandName(parentPath, cmd)
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
	llmsCmd.Flags().String("full-output", "./llms-full.txt", "Full output file path")

	return docsCmd
}
