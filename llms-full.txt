# Coolify CLI - llms-full.txt

> Full AI/LLM command catalog for the Coolify CLI.
> Manage Coolify instances (cloud and self-hosted), servers, projects, applications, databases, services, deployments, domains, and private keys.
> Source: https://github.com/coollabsio/coolify-cli
> API Spec: https://github.com/coollabsio/coolify/blob/v4.x/openapi.json

## Companion Files

- Quick instructions: ./llms.txt
- Regenerate docs: `go run ./coolify docs llms`

## Installation

```bash
# Linux/macOS (recommended)
curl -fsSL https://raw.githubusercontent.com/coollabsio/coolify-cli/main/scripts/install.sh | bash

# Homebrew (macOS/Linux)
brew install coollabsio/coolify-cli/coolify-cli

# Windows (PowerShell)
irm https://raw.githubusercontent.com/coollabsio/coolify-cli/main/scripts/install.ps1 | iex

# Go install
go install github.com/coollabsio/coolify-cli/coolify@latest
```

## Authentication

1. Get an API token from your Coolify dashboard at `/security/api-tokens`
2. For Coolify Cloud: `coolify context set-token cloud <token>`
3. For self-hosted: `coolify context add -d <context_name> <url> <token>`

## Configuration

Config file location:
- Linux/macOS: `~/.config/coolify/config.json`
- Windows: `%APPDATA%\coolify\config.json`

Supports multiple contexts (instances) with `coolify context` commands.

## Output Formats

All commands support `--format` flag:
- `table` (default) - human-readable tabular output
- `json` - compact JSON for scripting
- `pretty` - indented JSON for debugging

## Command Aliases

Aliases are derived from the CLI command tree:
- `coolify app env` | `coolify app envs` | `coolify app environment`
- `coolify app previews` | `coolify app preview`
- `coolify app start` | `coolify app deploy`
- `coolify app storage` | `coolify app storages`
- `coolify app` | `coolify apps` | `coolify application` | `coolify applications`
- `coolify database storage` | `coolify database storages`
- `coolify database` | `coolify databases` | `coolify db` | `coolify dbs`
- `coolify github` | `coolify gh` | `coolify github-app` | `coolify github-apps`
- `coolify private-key` | `coolify private-keys` | `coolify key` | `coolify keys`
- `coolify project` | `coolify projects`
- `coolify resource` | `coolify resources`
- `coolify server domains` | `coolify server domain`
- `coolify server` | `coolify servers`
- `coolify service storage` | `coolify service storages`
- `coolify service` | `coolify services` | `coolify svc`
- `coolify teams members` | `coolify teams member`
- `coolify teams` | `coolify team`


## Supported Database Types

When using `coolify database create <type>`:
- `postgresql`
- `mysql`
- `mariadb`
- `mongodb`
- `redis`
- `keydb`
- `clickhouse`
- `dragonfly`

## Usage Examples

```bash
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
```

## API Notes

- All resource identifiers use UUIDs (not internal database IDs)
- API base path: `/api/v1/`
- Authentication: Bearer token via `--token` flag or context configuration
- `app env sync` behavior: updates existing variables, creates missing ones, does NOT delete variables not in the file
- `app start` aliases to `app deploy` and also accepts `--force` and `--instant-deploy` flags
- Deployment logs support `--follow` for real-time streaming and `--debuglogs` for internal operations
- `app logs` defaults to 100 lines; `app deployments logs` defaults to 0 (all lines)
- Short flag `-n` can be used instead of `--lines` for log commands
- `completion` command supports shells: `bash`, `zsh`, `fish`, `powershell`
- Resource statuses: `running`, `stopped`, `error`
- Teams use numeric IDs (not UUIDs) - this is the only resource that uses IDs
- Fields marked `sensitive:"true"` (tokens, passwords, IPs, emails) are hidden by default; use `--show-sensitive` to reveal

---

## Command Reference

Command: coolify
Description: Coolify CLI
Parameters:
  - name: --context
    type: string
    description: Use specific context by name
    required: false
  - name: --debug
    type: boolean
    description: Debug mode
    required: false
    default: false
  - name: --format
    type: string
    description: Format output (table|json|pretty)
    required: false
    default: table
  - name: --show-sensitive (-s)
    type: boolean
    description: Show sensitive information
    required: false
    default: false
  - name: --token
    type: string
    description: Token for authentication (override context token)
    required: false

Command: coolify app create deploy-key
Description: Create an application from a private repository using SSH deploy key
Parameters:
  - name: --base-directory
    type: string
    description: Base directory for the application
    required: false
  - name: --build-command
    type: string
    description: Custom build command
    required: false
  - name: --build-pack
    type: string
    description: Build pack: nixpacks, static, dockerfile, dockercompose (required)
    required: true
  - name: --description
    type: string
    description: Application description
    required: false
  - name: --destination-uuid
    type: string
    description: Destination UUID if server has multiple destinations
    required: false
  - name: --dockerfile-target-build
    type: string
    description: Dockerfile target build stage
    required: false
  - name: --domains
    type: string
    description: Domain(s) for the application
    required: false
  - name: --environment-name
    type: string
    description: Environment name
    required: false
  - name: --environment-uuid
    type: string
    description: Environment UUID
    required: false
  - name: --git-branch
    type: string
    description: Git branch (required)
    required: true
  - name: --git-commit-sha
    type: string
    description: Specific commit SHA to deploy
    required: false
  - name: --git-repository
    type: string
    description: Git repository SSH URL, e.g., 'git@github.com:owner/repo.git' (required)
    required: true
  - name: --health-check-enabled
    type: boolean
    description: Enable health checks
    required: false
    default: false
  - name: --health-check-path
    type: string
    description: Health check path
    required: false
  - name: --install-command
    type: string
    description: Custom install command
    required: false
  - name: --instant-deploy
    type: boolean
    description: Deploy immediately after creation
    required: false
    default: false
  - name: --limits-cpus
    type: string
    description: CPU limit
    required: false
  - name: --limits-memory
    type: string
    description: Memory limit
    required: false
  - name: --name
    type: string
    description: Application name
    required: false
  - name: --ports-exposes
    type: string
    description: Exposed ports, e.g., '3000' or '3000,8080' (required)
    required: true
  - name: --ports-mappings
    type: string
    description: Port mappings (host:container)
    required: false
  - name: --private-key-uuid
    type: string
    description: Private key UUID (required)
    required: true
  - name: --project-uuid
    type: string
    description: Project UUID (required)
    required: true
  - name: --publish-directory
    type: string
    description: Publish directory for static builds
    required: false
  - name: --server-uuid
    type: string
    description: Server UUID (required)
    required: true
  - name: --start-command
    type: string
    description: Custom start command
    required: false

Command: coolify app create dockerfile
Description: Create an application from a custom Dockerfile
Parameters:
  - name: --description
    type: string
    description: Application description
    required: false
  - name: --destination-uuid
    type: string
    description: Destination UUID if server has multiple destinations
    required: false
  - name: --dockerfile
    type: string
    description: Dockerfile content (required)
    required: true
  - name: --dockerfile-target-build
    type: string
    description: Dockerfile target build stage
    required: false
  - name: --domains
    type: string
    description: Domain(s) for the application
    required: false
  - name: --environment-name
    type: string
    description: Environment name
    required: false
  - name: --environment-uuid
    type: string
    description: Environment UUID
    required: false
  - name: --health-check-enabled
    type: boolean
    description: Enable health checks
    required: false
    default: false
  - name: --health-check-path
    type: string
    description: Health check path
    required: false
  - name: --instant-deploy
    type: boolean
    description: Deploy immediately after creation
    required: false
    default: false
  - name: --limits-cpus
    type: string
    description: CPU limit
    required: false
  - name: --limits-memory
    type: string
    description: Memory limit
    required: false
  - name: --name
    type: string
    description: Application name
    required: false
  - name: --ports-exposes
    type: string
    description: Exposed ports, e.g., '3000' or '3000,8080'
    required: false
  - name: --ports-mappings
    type: string
    description: Port mappings (host:container)
    required: false
  - name: --project-uuid
    type: string
    description: Project UUID (required)
    required: true
  - name: --server-uuid
    type: string
    description: Server UUID (required)
    required: true

Command: coolify app create dockerimage
Description: Create an application from a pre-built Docker image
Parameters:
  - name: --description
    type: string
    description: Application description
    required: false
  - name: --destination-uuid
    type: string
    description: Destination UUID if server has multiple destinations
    required: false
  - name: --docker-registry-image-name
    type: string
    description: Docker image name from registry (required)
    required: true
  - name: --docker-registry-image-tag
    type: string
    description: Docker image tag (defaults to 'latest')
    required: false
  - name: --dockerfile-target-build
    type: string
    description: Dockerfile target build stage
    required: false
  - name: --domains
    type: string
    description: Domain(s) for the application
    required: false
  - name: --environment-name
    type: string
    description: Environment name
    required: false
  - name: --environment-uuid
    type: string
    description: Environment UUID
    required: false
  - name: --health-check-enabled
    type: boolean
    description: Enable health checks
    required: false
    default: false
  - name: --health-check-path
    type: string
    description: Health check path
    required: false
  - name: --instant-deploy
    type: boolean
    description: Deploy immediately after creation
    required: false
    default: false
  - name: --limits-cpus
    type: string
    description: CPU limit
    required: false
  - name: --limits-memory
    type: string
    description: Memory limit
    required: false
  - name: --name
    type: string
    description: Application name
    required: false
  - name: --ports-exposes
    type: string
    description: Exposed ports, e.g., '80' or '80,443' (required)
    required: true
  - name: --ports-mappings
    type: string
    description: Port mappings (host:container)
    required: false
  - name: --project-uuid
    type: string
    description: Project UUID (required)
    required: true
  - name: --server-uuid
    type: string
    description: Server UUID (required)
    required: true

Command: coolify app create github
Description: Create an application from a private repository using GitHub App
Parameters:
  - name: --base-directory
    type: string
    description: Base directory for the application
    required: false
  - name: --build-command
    type: string
    description: Custom build command
    required: false
  - name: --build-pack
    type: string
    description: Build pack: nixpacks, static, dockerfile, dockercompose (required)
    required: true
  - name: --description
    type: string
    description: Application description
    required: false
  - name: --destination-uuid
    type: string
    description: Destination UUID if server has multiple destinations
    required: false
  - name: --dockerfile-target-build
    type: string
    description: Dockerfile target build stage
    required: false
  - name: --domains
    type: string
    description: Domain(s) for the application
    required: false
  - name: --environment-name
    type: string
    description: Environment name
    required: false
  - name: --environment-uuid
    type: string
    description: Environment UUID
    required: false
  - name: --git-branch
    type: string
    description: Git branch (required)
    required: true
  - name: --git-commit-sha
    type: string
    description: Specific commit SHA to deploy
    required: false
  - name: --git-repository
    type: string
    description: Git repository in format 'owner/repo' (required)
    required: true
  - name: --github-app-uuid
    type: string
    description: GitHub App UUID (required)
    required: true
  - name: --health-check-enabled
    type: boolean
    description: Enable health checks
    required: false
    default: false
  - name: --health-check-path
    type: string
    description: Health check path
    required: false
  - name: --install-command
    type: string
    description: Custom install command
    required: false
  - name: --instant-deploy
    type: boolean
    description: Deploy immediately after creation
    required: false
    default: false
  - name: --limits-cpus
    type: string
    description: CPU limit
    required: false
  - name: --limits-memory
    type: string
    description: Memory limit
    required: false
  - name: --name
    type: string
    description: Application name
    required: false
  - name: --ports-exposes
    type: string
    description: Exposed ports, e.g., '3000' or '3000,8080' (required)
    required: true
  - name: --ports-mappings
    type: string
    description: Port mappings (host:container)
    required: false
  - name: --project-uuid
    type: string
    description: Project UUID (required)
    required: true
  - name: --publish-directory
    type: string
    description: Publish directory for static builds
    required: false
  - name: --server-uuid
    type: string
    description: Server UUID (required)
    required: true
  - name: --start-command
    type: string
    description: Custom start command
    required: false

Command: coolify app create public
Description: Create an application from a public git repository
Parameters:
  - name: --base-directory
    type: string
    description: Base directory for the application
    required: false
  - name: --build-command
    type: string
    description: Custom build command
    required: false
  - name: --build-pack
    type: string
    description: Build pack: nixpacks, static, dockerfile, dockercompose (required)
    required: true
  - name: --description
    type: string
    description: Application description
    required: false
  - name: --destination-uuid
    type: string
    description: Destination UUID if server has multiple destinations
    required: false
  - name: --dockerfile-target-build
    type: string
    description: Dockerfile target build stage
    required: false
  - name: --domains
    type: string
    description: Domain(s) for the application
    required: false
  - name: --environment-name
    type: string
    description: Environment name
    required: false
  - name: --environment-uuid
    type: string
    description: Environment UUID
    required: false
  - name: --git-branch
    type: string
    description: Git branch (required)
    required: true
  - name: --git-commit-sha
    type: string
    description: Specific commit SHA to deploy
    required: false
  - name: --git-repository
    type: string
    description: Git repository URL (required)
    required: true
  - name: --health-check-enabled
    type: boolean
    description: Enable health checks
    required: false
    default: false
  - name: --health-check-path
    type: string
    description: Health check path
    required: false
  - name: --install-command
    type: string
    description: Custom install command
    required: false
  - name: --instant-deploy
    type: boolean
    description: Deploy immediately after creation
    required: false
    default: false
  - name: --limits-cpus
    type: string
    description: CPU limit
    required: false
  - name: --limits-memory
    type: string
    description: Memory limit
    required: false
  - name: --name
    type: string
    description: Application name
    required: false
  - name: --ports-exposes
    type: string
    description: Exposed ports, e.g., '3000' or '3000,8080' (required)
    required: true
  - name: --ports-mappings
    type: string
    description: Port mappings (host:container)
    required: false
  - name: --project-uuid
    type: string
    description: Project UUID (required)
    required: true
  - name: --publish-directory
    type: string
    description: Publish directory for static builds
    required: false
  - name: --server-uuid
    type: string
    description: Server UUID (required)
    required: true
  - name: --start-command
    type: string
    description: Custom start command
    required: false

Command: coolify app delete <uuid>
Description: Delete an application. This action cannot be undone.
Parameters:
  - name: --force (-f)
    type: boolean
    description: Skip confirmation prompt
    required: false
    default: false

Command: coolify app deployments list <app-uuid>
Description: Retrieve a list of all deployments for a specific application.
Parameters: (None)

Command: coolify app deployments logs <app-uuid> [deployment-uuid]
Description: Get deployment logs for an application
Parameters:
  - name: --debuglogs
    type: boolean
    description: Show debug logs (includes hidden commands and internal operations)
    required: false
    default: false
  - name: --follow (-f)
    type: boolean
    description: Follow log output (like tail -f)
    required: false
    default: false
  - name: --lines (-n)
    type: integer
    description: Number of log lines to display (0 = all)
    required: false
    default: 0

Command: coolify app env create <app_uuid>
Description: Create a new environment variable for a specific application. Use --key and --value flags to specify the variable.
Parameters:
  - name: --build-time
    type: boolean
    description: Available at build time (default: true)
    required: false
    default: true
  - name: --comment
    type: string
    description: Comment for the environment variable
    required: false
  - name: --is-literal
    type: boolean
    description: Treat value as literal (don't interpolate variables)
    required: false
    default: false
  - name: --is-multiline
    type: boolean
    description: Value is multiline
    required: false
    default: false
  - name: --key
    type: string
    description: Environment variable key (required)
    required: true
  - name: --preview
    type: boolean
    description: Available in preview deployments
    required: false
    default: false
  - name: --runtime
    type: boolean
    description: Available at runtime (default: true)
    required: false
    default: true
  - name: --value
    type: string
    description: Environment variable value (required)
    required: true

Command: coolify app env delete <app_uuid> <env_uuid>
Description: Delete an environment variable from an application. First UUID is the application, second is the specific environment variable to delete.
Parameters:
  - name: --force
    type: boolean
    description: Skip confirmation prompt
    required: false
    default: false

Command: coolify app env get <app_uuid> <env_uuid_or_key>
Description: Get detailed information about a specific environment variable by UUID or key name.
Parameters: (None)

Command: coolify app env list <app_uuid>
Description: List all environment variables for an application
Parameters:
  - name: --all
    type: boolean
    description: Show all environment variables (non-preview first, then preview)
    required: false
    default: false
  - name: --preview
    type: boolean
    description: Show preview environment variables instead of regular ones
    required: false
    default: false

Command: coolify app env sync <app_uuid>
Description: Sync environment variables from a .env file
Parameters:
  - name: --build-time
    type: boolean
    description: Make all variables available at build time (default: true)
    required: false
    default: true
  - name: --file (-f)
    type: string
    description: Path to .env file (required)
    required: true
  - name: --is-literal
    type: boolean
    description: Treat all values as literal (don't interpolate variables)
    required: false
    default: false
  - name: --preview
    type: boolean
    description: Make all variables available in preview deployments
    required: false
    default: false
  - name: --runtime
    type: boolean
    description: Make all variables available at runtime (default: true)
    required: false
    default: true

Command: coolify app env update <app_uuid> <env_uuid_or_key>
Description: Update an existing environment variable. Identify it by UUID or key name.
Parameters:
  - name: --build-time
    type: boolean
    description: Available at build time (default: true)
    required: false
    default: true
  - name: --comment
    type: string
    description: Comment for the environment variable
    required: false
  - name: --is-literal
    type: boolean
    description: Treat value as literal
    required: false
    default: false
  - name: --is-multiline
    type: boolean
    description: Value is multiline
    required: false
    default: false
  - name: --key
    type: string
    description: New environment variable key (rename)
    required: false
  - name: --preview
    type: boolean
    description: Available in preview deployments
    required: false
    default: false
  - name: --runtime
    type: boolean
    description: Available at runtime (default: true)
    required: false
    default: true
  - name: --value
    type: string
    description: New environment variable value (required)
    required: true

Command: coolify app get <uuid>
Description: Retrieve detailed information about a specific application.
Parameters: (None)

Command: coolify app list
Description: List all applications in Coolify.
Parameters: (None)

Command: coolify app logs <uuid>
Description: Retrieve logs for an application. Use --follow to continuously stream new logs.
Parameters:
  - name: --follow (-f)
    type: boolean
    description: Follow log output (like tail -f)
    required: false
    default: false
  - name: --lines (-n)
    type: integer
    description: Number of log lines to retrieve
    required: false
    default: 100

Command: coolify app previews delete <app_uuid> <pr_id>
Description: Delete a preview deployment for an application. First argument is the application UUID, second is the pull request ID.
Parameters:
  - name: --force
    type: boolean
    description: Skip confirmation prompt
    required: false
    default: false

Command: coolify app restart <uuid>
Description: Restart a running application.
Parameters: (None)

Command: coolify app start <uuid>
Description: Start an application (initiates a deployment).
Parameters:
  - name: --force
    type: boolean
    description: Force rebuild
    required: false
    default: false
  - name: --instant-deploy
    type: boolean
    description: Instant deploy (skip queuing)
    required: false
    default: false

Command: coolify app stop <uuid>
Description: Stop a running application.
Parameters: (None)

Command: coolify app storage create <app_uuid>
Description: Create a storage for an application
Parameters:
  - name: --content
    type: string
    description: File content (file only)
    required: false
  - name: --fs-path
    type: string
    description: Host directory path (file only, required when --is-directory is set)
    required: false
  - name: --host-path
    type: string
    description: Host path (persistent only)
    required: false
  - name: --is-directory
    type: boolean
    description: Whether this is a directory mount (file only)
    required: false
    default: false
  - name: --mount-path
    type: string
    description: Mount path inside the container (required)
    required: true
  - name: --name
    type: string
    description: Volume name (persistent only)
    required: false
  - name: --type
    type: string
    description: Storage type: 'persistent' or 'file' (required)
    required: true

Command: coolify app storage delete <app_uuid> <storage_uuid>
Description: Delete a storage from an application
Parameters: (None)

Command: coolify app storage list <app_uuid>
Description: List all persistent volumes and file storages for a specific application.
Parameters: (None)

Command: coolify app storage update <app_uuid>
Description: Update a storage for an application
Parameters:
  - name: --content
    type: string
    description: File content (file only)
    required: false
  - name: --host-path
    type: string
    description: Host path (persistent only)
    required: false
  - name: --id
    type: integer
    description: Storage ID (deprecated, use --uuid instead)
    required: false
    default: 0
  - name: --is-preview-suffix-enabled
    type: boolean
    description: Enable preview suffix for this storage
    required: false
    default: false
  - name: --mount-path
    type: string
    description: Mount path inside the container
    required: false
  - name: --name
    type: string
    description: Storage name (persistent only)
    required: false
  - name: --type
    type: string
    description: Storage type: 'persistent' or 'file' (required)
    required: true
  - name: --uuid
    type: string
    description: Storage UUID (required, use 'storage list' to find)
    required: false

Command: coolify app update <uuid>
Description: Update configuration for a specific application. Only specified fields will be updated.
Parameters:
  - name: --base-directory
    type: string
    description: Base directory
    required: false
  - name: --build-command
    type: string
    description: Build command
    required: false
  - name: --description
    type: string
    description: Application description
    required: false
  - name: --docker-image
    type: string
    description: Docker image name
    required: false
  - name: --docker-tag
    type: string
    description: Docker image tag
    required: false
  - name: --dockerfile
    type: string
    description: Dockerfile content
    required: false
  - name: --dockerfile-target-build
    type: string
    description: Dockerfile target build stage
    required: false
  - name: --domains
    type: string
    description: Domains (comma-separated)
    required: false
  - name: --git-branch
    type: string
    description: Git branch
    required: false
  - name: --git-repository
    type: string
    description: Git repository URL
    required: false
  - name: --health-check-enabled
    type: boolean
    description: Enable health check
    required: false
    default: false
  - name: --health-check-path
    type: string
    description: Health check path
    required: false
  - name: --install-command
    type: string
    description: Install command
    required: false
  - name: --name
    type: string
    description: Application name
    required: false
  - name: --ports-exposes
    type: string
    description: Exposed ports
    required: false
  - name: --ports-mappings
    type: string
    description: Port mappings
    required: false
  - name: --publish-directory
    type: string
    description: Publish directory
    required: false
  - name: --start-command
    type: string
    description: Start command
    required: false

Command: coolify completion <shell>
Description: Output shell completion code for the specified shell
Parameters: (None)

Command: coolify config
Description: Display the path to the Coolify CLI configuration file
Parameters: (None)

Command: coolify context add <context_name> <url> <token>
Description: Add a new context
Parameters:
  - name: --default (-d)
    type: boolean
    description: Set as default context
    required: false
    default: false
  - name: --force (-f)
    type: boolean
    description: Force overwrite if context already exists
    required: false
    default: false

Command: coolify context delete <context_name>
Description: Delete a context
Parameters: (None)

Command: coolify context get <context_name>
Description: Get details of a specific context
Parameters: (None)

Command: coolify context list
Description: List all configured contexts
Parameters: (None)

Command: coolify context set-default <context_name>
Description: Set a context as the default
Parameters: (None)

Command: coolify context set-token <context_name> <token>
Description: Update the API token for a context
Parameters: (None)

Command: coolify context update <context_name>
Description: Update a context's properties (name, URL, token)
Parameters:
  - name: --name (-n)
    type: string
    description: New name for the context
    required: false
  - name: --token (-t)
    type: string
    description: New token for the context
    required: false
  - name: --url (-u)
    type: string
    description: New URL for the context
    required: false

Command: coolify context use <context_name>
Description: Switch to a different context (set as default)
Parameters: (None)

Command: coolify context verify
Description: Verify current context connection and authentication
Parameters: (None)

Command: coolify context version
Description: Get current context's Coolify version
Parameters: (None)

Command: coolify database backup create <database_uuid>
Description: Create a new scheduled backup configuration
Parameters:
  - name: --databases-to-backup
    type: string
    description: Comma-separated list of databases to backup
    required: false
  - name: --disable-local-backup
    type: boolean
    description: Disable local backup storage
    required: false
    default: false
  - name: --dump-all
    type: boolean
    description: Dump all databases
    required: false
    default: false
  - name: --enabled
    type: boolean
    description: Enable backup schedule
    required: false
    default: false
  - name: --frequency
    type: string
    description: Backup frequency (cron expression, e.g., '0 0 * * *' for daily)
    required: false
  - name: --retention-amount-locally
    type: integer
    description: Number of backups to retain locally
    required: false
    default: 0
  - name: --retention-amount-s3
    type: integer
    description: Number of backups to retain in S3
    required: false
    default: 0
  - name: --retention-days-locally
    type: integer
    description: Days to retain backups locally
    required: false
    default: 0
  - name: --retention-days-s3
    type: integer
    description: Days to retain backups in S3
    required: false
    default: 0
  - name: --retention-max-storage-locally
    type: string
    description: Max storage for local backups (e.g., '1GB', '500MB')
    required: false
  - name: --retention-max-storage-s3
    type: string
    description: Max storage for S3 backups (e.g., '1GB', '500MB')
    required: false
  - name: --s3-storage-uuid
    type: string
    description: S3 storage UUID
    required: false
  - name: --save-s3
    type: boolean
    description: Save backups to S3
    required: false
    default: false
  - name: --timeout
    type: integer
    description: Backup timeout in seconds
    required: false
    default: 0

Command: coolify database backup delete <database_uuid> <backup_uuid>
Description: Delete a backup configuration and optionally all its executions from S3. First UUID is the database, second is the specific backup configuration.
Parameters:
  - name: --delete-s3
    type: boolean
    description: Delete backup files from S3
    required: false
    default: false

Command: coolify database backup delete-execution <database_uuid> <backup_uuid> <execution_uuid>
Description: Delete a specific backup execution and optionally from S3. First UUID is the database, second is the backup configuration, third is the specific execution.
Parameters:
  - name: --delete-s3
    type: boolean
    description: Delete backup file from S3
    required: false
    default: false

Command: coolify database backup executions <database_uuid> <backup_uuid>
Description: List all executions for a backup configuration. First UUID is the database, second is the specific backup configuration.
Parameters: (None)

Command: coolify database backup list <database_uuid>
Description: List all backup configurations for a specific database.
Parameters: (None)

Command: coolify database backup trigger <database_uuid> <backup_uuid>
Description: Trigger an immediate backup for a specific backup configuration. First UUID is the database, second is the specific backup configuration to trigger.
Parameters: (None)

Command: coolify database backup update <database_uuid> <backup_uuid>
Description: Update a backup configuration settings (frequency, retention, S3, etc.). First UUID is the database, second is the specific backup configuration.
Parameters:
  - name: --databases-to-backup
    type: string
    description: Comma-separated list of databases to backup
    required: false
  - name: --dump-all
    type: boolean
    description: Dump all databases
    required: false
    default: false
  - name: --enabled
    type: boolean
    description: Enable or disable backup
    required: false
    default: false
  - name: --frequency
    type: string
    description: Backup frequency (cron expression)
    required: false
  - name: --retention-amount-locally
    type: integer
    description: Number of backups to retain locally
    required: false
    default: 0
  - name: --retention-amount-s3
    type: integer
    description: Number of backups to retain in S3
    required: false
    default: 0
  - name: --retention-days-locally
    type: integer
    description: Days to retain backups locally
    required: false
    default: 0
  - name: --retention-days-s3
    type: integer
    description: Days to retain backups in S3
    required: false
    default: 0
  - name: --retention-max-storage-locally
    type: integer
    description: Max storage for local backups (MB)
    required: false
    default: 0
  - name: --retention-max-storage-s3
    type: integer
    description: Max storage for S3 backups (MB)
    required: false
    default: 0
  - name: --s3-storage-uuid
    type: string
    description: S3 storage UUID
    required: false
  - name: --save-s3
    type: boolean
    description: Save backups to S3
    required: false
    default: false

Command: coolify database create <type>
Description: Create a new database
Parameters:
  - name: --clickhouse-admin-password
    type: string
    description: Clickhouse admin password
    required: false
  - name: --clickhouse-admin-user
    type: string
    description: Clickhouse admin user
    required: false
  - name: --description
    type: string
    description: Database description
    required: false
  - name: --destination-uuid
    type: string
    description: Destination UUID if server has multiple destinations
    required: false
  - name: --dragonfly-password
    type: string
    description: Dragonfly password
    required: false
  - name: --environment-name
    type: string
    description: Environment name
    required: false
  - name: --environment-uuid
    type: string
    description: Environment UUID
    required: false
  - name: --image
    type: string
    description: Docker image
    required: false
  - name: --instant-deploy
    type: boolean
    description: Deploy immediately after creation
    required: false
    default: false
  - name: --is-public
    type: boolean
    description: Make database publicly accessible
    required: false
    default: false
  - name: --keydb-password
    type: string
    description: KeyDB password
    required: false
  - name: --limits-cpus
    type: string
    description: CPU limit (e.g., '0.5', '2')
    required: false
  - name: --limits-memory
    type: string
    description: Memory limit (e.g., '512m', '2g')
    required: false
  - name: --mariadb-database
    type: string
    description: MariaDB database name
    required: false
  - name: --mariadb-password
    type: string
    description: MariaDB password
    required: false
  - name: --mariadb-root-password
    type: string
    description: MariaDB root password
    required: false
  - name: --mariadb-user
    type: string
    description: MariaDB user
    required: false
  - name: --mongo-database
    type: string
    description: MongoDB database name
    required: false
  - name: --mongo-root-password
    type: string
    description: MongoDB root password
    required: false
  - name: --mongo-root-username
    type: string
    description: MongoDB root username
    required: false
  - name: --mysql-database
    type: string
    description: MySQL database name
    required: false
  - name: --mysql-password
    type: string
    description: MySQL password
    required: false
  - name: --mysql-root-password
    type: string
    description: MySQL root password
    required: false
  - name: --mysql-user
    type: string
    description: MySQL user
    required: false
  - name: --name
    type: string
    description: Database name
    required: false
  - name: --postgres-db
    type: string
    description: PostgreSQL database name
    required: false
  - name: --postgres-password
    type: string
    description: PostgreSQL password
    required: false
  - name: --postgres-user
    type: string
    description: PostgreSQL user
    required: false
  - name: --project-uuid
    type: string
    description: Project UUID (required)
    required: true
  - name: --public-port
    type: integer
    description: Public port
    required: false
    default: 0
  - name: --redis-password
    type: string
    description: Redis password
    required: false
  - name: --server-uuid
    type: string
    description: Server UUID (required)
    required: true

Command: coolify database delete <uuid>
Description: Delete a database and optionally clean up its configurations, volumes, and networks.
Parameters:
  - name: --delete-configurations
    type: boolean
    description: Delete configurations
    required: false
    default: true
  - name: --delete-connected-networks
    type: boolean
    description: Delete connected networks
    required: false
    default: true
  - name: --delete-volumes
    type: boolean
    description: Delete volumes
    required: false
    default: true
  - name: --docker-cleanup
    type: boolean
    description: Run docker cleanup
    required: false
    default: true

Command: coolify database env create <database_uuid>
Description: Create a new environment variable for a specific database. Use --key and --value flags to specify the variable.
Parameters:
  - name: --comment
    type: string
    description: Comment for the environment variable
    required: false
  - name: --is-literal
    type: boolean
    description: Treat value as literal (don't interpolate variables)
    required: false
    default: false
  - name: --is-multiline
    type: boolean
    description: Value is multiline
    required: false
    default: false
  - name: --is-shown-once
    type: boolean
    description: Only show value once
    required: false
    default: false
  - name: --key
    type: string
    description: Environment variable key (required)
    required: true
  - name: --value
    type: string
    description: Environment variable value (required)
    required: true

Command: coolify database env delete <database_uuid> <env_uuid>
Description: Delete an environment variable from a database. First UUID is the database, second is the specific environment variable to delete.
Parameters:
  - name: --force
    type: boolean
    description: Skip confirmation prompt
    required: false
    default: false

Command: coolify database env get <database_uuid> <env_uuid_or_key>
Description: Get detailed information about a specific environment variable. First UUID is the database, second is the environment variable UUID or key name.
Parameters: (None)

Command: coolify database env list <database_uuid>
Description: List all environment variables for a specific database.
Parameters: (None)

Command: coolify database env sync <database_uuid>
Description: Sync environment variables from a .env file
Parameters:
  - name: --file (-f)
    type: string
    description: Path to .env file (required)
    required: true
  - name: --is-literal
    type: boolean
    description: Treat all values as literal (don't interpolate variables)
    required: false
    default: false

Command: coolify database env update <database_uuid> <env_uuid_or_key>
Description: Update an existing environment variable. Identify it by UUID or key name.
Parameters:
  - name: --comment
    type: string
    description: Comment for the environment variable
    required: false
  - name: --is-literal
    type: boolean
    description: Treat value as literal
    required: false
    default: false
  - name: --is-multiline
    type: boolean
    description: Value is multiline
    required: false
    default: false
  - name: --is-shown-once
    type: boolean
    description: Only show value once
    required: false
    default: false
  - name: --key
    type: string
    description: New environment variable key (rename)
    required: false
  - name: --value
    type: string
    description: New environment variable value (required)
    required: true

Command: coolify database get <uuid>
Description: Get detailed information about a specific database by UUID.
Parameters: (None)

Command: coolify database list
Description: List all databases in Coolify.
Parameters: (None)

Command: coolify database restart <uuid>
Description: Restart a database by UUID.
Parameters: (None)

Command: coolify database start <uuid>
Description: Start a database by UUID.
Parameters: (None)

Command: coolify database stop <uuid>
Description: Stop a database by UUID.
Parameters: (None)

Command: coolify database storage create <db_uuid>
Description: Create a storage for a database
Parameters:
  - name: --content
    type: string
    description: File content (file only)
    required: false
  - name: --fs-path
    type: string
    description: Host directory path (file only, required when --is-directory is set)
    required: false
  - name: --host-path
    type: string
    description: Host path (persistent only)
    required: false
  - name: --is-directory
    type: boolean
    description: Whether this is a directory mount (file only)
    required: false
    default: false
  - name: --mount-path
    type: string
    description: Mount path inside the container (required)
    required: true
  - name: --name
    type: string
    description: Volume name (persistent only)
    required: false
  - name: --type
    type: string
    description: Storage type: 'persistent' or 'file' (required)
    required: true

Command: coolify database storage delete <db_uuid> <storage_uuid>
Description: Delete a storage from a database
Parameters: (None)

Command: coolify database storage list <db_uuid>
Description: List all persistent volumes and file storages for a specific database.
Parameters: (None)

Command: coolify database storage update <db_uuid>
Description: Update a storage for a database
Parameters:
  - name: --content
    type: string
    description: File content (file only)
    required: false
  - name: --host-path
    type: string
    description: Host path (persistent only)
    required: false
  - name: --id
    type: integer
    description: Storage ID (deprecated, use --uuid instead)
    required: false
    default: 0
  - name: --is-preview-suffix-enabled
    type: boolean
    description: Enable preview suffix for this storage
    required: false
    default: false
  - name: --mount-path
    type: string
    description: Mount path inside the container
    required: false
  - name: --name
    type: string
    description: Storage name (persistent only)
    required: false
  - name: --type
    type: string
    description: Storage type: 'persistent' or 'file' (required)
    required: true
  - name: --uuid
    type: string
    description: Storage UUID (required, use 'storage list' to find)
    required: false

Command: coolify database update <uuid>
Description: Update a database's configuration by UUID.
Parameters:
  - name: --description
    type: string
    description: Database description
    required: false
  - name: --image
    type: string
    description: Docker image
    required: false
  - name: --is-public
    type: boolean
    description: Make database publicly accessible
    required: false
    default: false
  - name: --limits-cpus
    type: string
    description: CPU limit
    required: false
  - name: --limits-memory
    type: string
    description: Memory limit
    required: false
  - name: --name
    type: string
    description: Database name
    required: false
  - name: --public-port
    type: integer
    description: Public port
    required: false
    default: 0

Command: coolify deploy batch <name1,name2,...>
Description: Deploy multiple resources by name
Parameters:
  - name: --docker-tag
    type: string
    description: Docker image tag override for the deployment
    required: false
  - name: --force
    type: boolean
    description: Force deployment
    required: false
    default: false
  - name: --pull-request-id
    type: integer
    description: Pull request ID for preview deployments
    required: false
    default: 0

Command: coolify deploy cancel <uuid>
Description: Cancel an in-progress deployment. This will stop the deployment process and clean up any temporary resources.
Parameters:
  - name: --force (-f)
    type: boolean
    description: Skip confirmation prompt
    required: false
    default: false

Command: coolify deploy get <uuid>
Description: Get detailed information about a specific deployment by its UUID.
Parameters: (None)

Command: coolify deploy list
Description: List all currently running deployments across all resources.
Parameters: (None)

Command: coolify deploy name <resource_name>
Description: Deploy by resource name
Parameters:
  - name: --docker-tag
    type: string
    description: Docker image tag override for the deployment
    required: false
  - name: --force
    type: boolean
    description: Force deployment
    required: false
    default: false
  - name: --pull-request-id
    type: integer
    description: Pull request ID for preview deployments
    required: false
    default: 0

Command: coolify deploy uuid <uuid>
Description: Deploy by uuid
Parameters:
  - name: --docker-tag
    type: string
    description: Docker image tag override for the deployment
    required: false
  - name: --force
    type: boolean
    description: Force deployment
    required: false
    default: false
  - name: --pull-request-id
    type: integer
    description: Pull request ID for preview deployments
    required: false
    default: 0

Command: coolify firewall
Description: [ALPHA] Manage cross-host container allow rules (Coolify v5)
Parameters:
  - name: --all-namespaces
    type: boolean
    description: Operate across every mesh namespace on each host (list/containers fan out; allow/revoke still require a specific --namespace)
    required: false
    default: false
  - name: --concurrency
    type: integer
    description: Maximum number of parallel SSH connections
    required: false
    default: 10
  - name: --coold-port
    type: integer
    description: TCP port coold's REST API listens on (bound to the WG mgmt IP)
    required: false
    default: 8443
  - name: --coold-token
    type: string
    description: Bearer token override for coold REST API (also reads COOLIFY_COOLD_TOKEN env). When unset, CLI reads /etc/coolify/api-token over SSH per host.
    required: false
  - name: --namespace
    type: string
    description: Namespace the command operates against (must match a namespace created by `coolify init`)
    required: false
    default: default
  - name: --servers
    type: stringSlice
    description: Comma-separated server IPs (required)
    required: true
  - name: --ssh-key
    type: string
    description: Path to SSH private key used to connect to servers (required)
    required: true
  - name: --ssh-passphrase-prompt
    type: boolean
    description: Prompt for SSH key passphrase (also reads COOLIFY_SSH_PASSPHRASE env var)
    required: false
    default: false
  - name: --ssh-port
    type: integer
    description: SSH port
    required: false
    default: 22
  - name: --ssh-timeout
    type: string
    description: SSH connection timeout (e.g. 30s, 1m)
    required: false
    default: 30s
  - name: --ssh-user
    type: string
    description: SSH username
    required: false
    default: root
  - name: --wg-interface
    type: string
    description: WireGuard interface name on remote hosts (must match --wg-interface at init)
    required: false
    default: wg0

Command: coolify firewall allow
Description: Add an allow rule (from container → to container:port)
Parameters:
  - name: --bidirectional
    type: boolean
    description: Also install the reverse rule on the source host (default: one-way; conntrack handles replies)
    required: false
    default: false
  - name: --from
    type: string
    description: Source container (name, short-id, raw IP, or host:name) — required
    required: false
  - name: --port
    type: integer
    description: Destination port (required unless --proto is empty)
    required: false
    default: 0
  - name: --proto
    type: string
    description: Protocol (tcp, udp, or empty for any)
    required: false
    default: tcp
  - name: --to
    type: string
    description: Destination container (name, short-id, raw IP, or host:name) — required
    required: false

Command: coolify firewall containers
Description: List containers on the Coolify mesh bridge across all servers
Parameters: (None)

Command: coolify firewall list
Description: List installed allow rules across all servers
Parameters: (None)

Command: coolify firewall revoke
Description: Remove an allow rule
Parameters:
  - name: --bidirectional
    type: boolean
    description: Also install the reverse rule on the source host (default: one-way; conntrack handles replies)
    required: false
    default: false
  - name: --from
    type: string
    description: Source container (name, short-id, raw IP, or host:name) — required
    required: false
  - name: --port
    type: integer
    description: Destination port (required unless --proto is empty)
    required: false
    default: 0
  - name: --proto
    type: string
    description: Protocol (tcp, udp, or empty for any)
    required: false
    default: tcp
  - name: --to
    type: string
    description: Destination container (name, short-id, raw IP, or host:name) — required
    required: false

Command: coolify github branches <app_uuid> <owner/repo>
Description: List branches for a repository
Parameters: (None)

Command: coolify github create
Description: Create a GitHub App integration
Parameters:
  - name: --api-url
    type: string
    description: GitHub API URL (required, e.g., https://api.github.com)
    required: true
  - name: --app-id
    type: integer
    description: GitHub App ID (required)
    required: true
    default: 0
  - name: --client-id
    type: string
    description: GitHub OAuth Client ID (required)
    required: true
  - name: --client-secret
    type: string
    description: GitHub OAuth Client Secret (required)
    required: true
  - name: --custom-port
    type: integer
    description: Custom port for SSH (default: 22)
    required: false
    default: 0
  - name: --custom-user
    type: string
    description: Custom user for SSH (default: git)
    required: false
  - name: --html-url
    type: string
    description: GitHub HTML URL (required, e.g., https://github.com)
    required: true
  - name: --installation-id
    type: integer
    description: GitHub Installation ID (required)
    required: true
    default: 0
  - name: --name
    type: string
    description: GitHub App name (required)
    required: true
  - name: --organization
    type: string
    description: GitHub organization
    required: false
  - name: --private-key-uuid
    type: string
    description: UUID of existing private key (required)
    required: true
  - name: --system-wide
    type: boolean
    description: Is this app system-wide (cloud only)
    required: false
    default: false
  - name: --webhook-secret
    type: string
    description: GitHub Webhook Secret
    required: false

Command: coolify github delete <app_uuid>
Description: Delete a GitHub App integration. The app must not be used by any applications.
Parameters:
  - name: --force (-f)
    type: boolean
    description: Skip confirmation prompt
    required: false
    default: false

Command: coolify github get <app_uuid>
Description: Get detailed information about a specific GitHub App integration.
Parameters: (None)

Command: coolify github list
Description: List all GitHub App integrations configured in Coolify.
Parameters: (None)

Command: coolify github repos <app_uuid>
Description: List all repositories that are accessible by the specified GitHub App.
Parameters: (None)

Command: coolify github update <app_uuid>
Description: Update an existing GitHub App integration. Provide the app UUID and the fields you want to update.
Parameters:
  - name: --api-url
    type: string
    description: GitHub API URL
    required: false
  - name: --app-id
    type: integer
    description: GitHub App ID
    required: false
    default: 0
  - name: --client-id
    type: string
    description: GitHub OAuth Client ID
    required: false
  - name: --client-secret
    type: string
    description: GitHub OAuth Client Secret
    required: false
  - name: --custom-port
    type: integer
    description: Custom port for SSH
    required: false
    default: 0
  - name: --custom-user
    type: string
    description: Custom user for SSH
    required: false
  - name: --html-url
    type: string
    description: GitHub HTML URL
    required: false
  - name: --installation-id
    type: integer
    description: GitHub Installation ID
    required: false
    default: 0
  - name: --name
    type: string
    description: GitHub App name
    required: false
  - name: --organization
    type: string
    description: GitHub organization
    required: false
  - name: --private-key-uuid
    type: string
    description: UUID of private key
    required: false
  - name: --system-wide
    type: boolean
    description: Is this app system-wide
    required: false
    default: false
  - name: --webhook-secret
    type: string
    description: GitHub Webhook Secret
    required: false

Command: coolify init
Description: [ALPHA] Initialize WireGuard mesh for Coolify v5
Parameters:
  - name: --builder-capacity
    type: integer
    description: Concurrent builds accepted per host (COOLD_BUILDER_CAPACITY).
    required: false
    default: 2
  - name: --builder-cpu-quota
    type: string
    description: cgroup CPU quota for each build subprocess (COOLD_BUILDER_CPU_QUOTA).
systemd CPUQuota format; "200%" = two full cores.
    required: false
    default: 200%
  - name: --builder-hosts
    type: stringSlice
    description: Explicit subset of --servers to enroll with the builder capability.
Takes precedence over --enable-builder. Empty (default) means fall back to
--enable-builder for the whole cluster.
    required: false
  - name: --builder-memory-max
    type: string
    description: cgroup memory cap for each build subprocess (COOLD_BUILDER_MEMORY_MAX).
systemd MemoryMax format; e.g. "2G", "512M".
    required: false
    default: 2G
  - name: --builder-timeout-secs
    type: integer
    description: Hard wall-clock timeout per build in seconds (COOLD_BUILDER_TIMEOUT_SECS).
    required: false
    default: 1800
  - name: --central
    type: string
    description: SSH address of the central VM that will run the scheduler (and later Laravel).
Must be one of the --servers entries. When set, phases 4+5 install the scheduler on that host
and push a per-host JWT to every other server. Leave empty to skip scheduler setup.
    required: false
  - name: --concurrency
    type: integer
    description: Maximum number of parallel SSH connections
    required: false
    default: 10
  - name: --container-pool
    type: string
    description: Shared container address pool — each (namespace, host) pair gets a /<container-prefix> from here, owned by that namespace's Podman bridge
    required: false
    default: 10.210.0.0/16
  - name: --container-prefix
    type: integer
    description: Prefix length of each per-host, per-namespace container subnet
    required: false
    default: 24
  - name: --coold-version
    type: string
    description: Release tag to download for coold (e.g. "nightly", "v1.2.3"). nightly always re-installs on every apply.
    required: false
    default: nightly
  - name: --corrosion-api-port
    type: integer
    description: Corrosion HTTP API port (bound to 127.0.0.1)
    required: false
    default: 8080
  - name: --corrosion-gossip-port
    type: integer
    description: Corrosion SWIM gossip port (bound to the wg0 mgmt IP)
    required: false
    default: 8787
  - name: --corrosion-version
    type: string
    description: Release tag to download for corrosion (e.g. "nightly", "v1.2.3"). nightly always re-installs on every apply.
    required: false
    default: nightly
  - name: --enable-builder
    type: boolean
    description: Cluster-wide shorthand: enable the builder capability on every host
(requires --central). Ignored when --builder-hosts is set.
    required: false
    default: true
  - name: --namespaces
    type: stringSlice
    description: Comma-separated list of namespaces to create on each host. Each namespace is a separate Podman bridge network (coolify-<ns>-mesh) with its own /<container-prefix> per host
    required: false
    default: [default]
  - name: --scheduler-version
    type: string
    description: Release tag to download for scheduler (e.g. "nightly", "v1.2.3").
    required: false
    default: nightly
  - name: --servers
    type: stringSlice
    description: Comma-separated server IPs (required)
    required: true
  - name: --skip-default-deny
    type: boolean
    description: Skip installing the default-deny firewall scaffold. By default, both cross-host and intra-host (same bridge) container traffic is blocked; coold manages the allow list at runtime
    required: false
    default: false
  - name: --ssh-key
    type: string
    description: Path to SSH private key used to connect to servers (required)
    required: true
  - name: --ssh-passphrase-prompt
    type: boolean
    description: Prompt for SSH key passphrase (also reads COOLIFY_SSH_PASSPHRASE env var)
    required: false
    default: false
  - name: --ssh-port
    type: integer
    description: SSH port
    required: false
    default: 22
  - name: --ssh-timeout
    type: string
    description: SSH connection timeout (e.g. 30s, 1m)
    required: false
    default: 30s
  - name: --ssh-user
    type: string
    description: SSH username
    required: false
    default: root
  - name: --wg-interface
    type: string
    description: WireGuard interface name on the remote hosts
    required: false
    default: wg0
  - name: --wg-listen-port
    type: integer
    description: WireGuard UDP listen port
    required: false
    default: 51820
  - name: --wg-mgmt-pool
    type: string
    description: WireGuard management address pool — each host gets a /32 from here, assigned to wg0
    required: false
    default: 100.64.0.0/16
  - name: --yes (-y)
    type: boolean
    description: Skip the interactive alpha confirmation prompt
    required: false
    default: false

Command: coolify init bootstrap
Description: First-time mesh install (all actions allowed)
Parameters: (None)

Command: coolify init extend
Description: Add new hosts to an existing mesh (existing hosts stay untouched)
Parameters:
  - name: --allow-replace
    type: boolean
    description: Unlock destructive-replace actions on existing hosts (e.g. recreating a drifted podman bridge). Off by default — drifted existing hosts are surfaced as skipped actions instead.
    required: false
    default: false
  - name: --new-hosts
    type: stringSlice
    description: Comma-separated subset of --servers that is brand-new this run (required). Only these hosts receive the full first-time install; all other hosts get peer-refresh only.
    required: true

Command: coolify init plan
Description: Show WireGuard mesh changes without applying them
Parameters:
  - name: --intent
    type: string
    description: Preview filter: "bootstrap" (all actions), "extend" (treat --new-hosts as fresh, existing hosts peer-refresh only), "upgrade" (version bumps only).
    required: false
    default: bootstrap

Command: coolify init upgrade
Description: Bump agent binary versions (coold / corrosion / scheduler / builder) on every host
Parameters:
  - name: --allow-nightly
    type: boolean
    description: Permit --coold-version/--corrosion-version/--scheduler-version=nightly. Off by default because nightly re-installs on every run instead of only when the pinned version changes.
    required: false
    default: false

Command: coolify private-key add <key_name> <private_key_or_file>
Description: Add a private key
Parameters: (None)

Command: coolify private-key list
Description: List all private keys
Parameters: (None)

Command: coolify private-key remove <uuid>
Description: Remove a private key
Parameters: (None)

Command: coolify project create
Description: Create a new project
Parameters:
  - name: --description
    type: string
    description: Project description
    required: false
  - name: --name
    type: string
    description: Project name (required)
    required: true

Command: coolify project get <uuid>
Description: Get a project by uuid
Parameters: (None)

Command: coolify project list
Description: List all projects
Parameters: (None)

Command: coolify resource list
Description: List all resources
Parameters: (None)

Command: coolify server add <server_name> <ip_address> <private_key_uuid>
Description: Add a server
Parameters:
  - name: --port (-p)
    type: integer
    description: Port
    required: false
    default: 22
  - name: --user (-u)
    type: string
    description: User
    required: false
    default: root
  - name: --validate
    type: boolean
    description: Validate the server
    required: false
    default: false

Command: coolify server domains <uuid>
Description: Get server domains by uuid
Parameters: (None)

Command: coolify server get <uuid>
Description: Get server details by uuid
Parameters:
  - name: --resources
    type: boolean
    description: With resources
    required: false
    default: false

Command: coolify server list
Description: List all servers
Parameters: (None)

Command: coolify server remove <uuid>
Description: Remove a server
Parameters: (None)

Command: coolify server validate <uuid>
Description: Validate a server
Parameters: (None)

Command: coolify service create <type>
Description: Create a new one-click service
Parameters:
  - name: --description
    type: string
    description: Service description
    required: false
  - name: --destination-uuid
    type: string
    description: Destination UUID if server has multiple destinations
    required: false
  - name: --docker-compose
    type: string
    description: Custom Docker Compose content (for advanced customization)
    required: false
  - name: --environment-name
    type: string
    description: Environment name
    required: false
  - name: --environment-uuid
    type: string
    description: Environment UUID
    required: false
  - name: --instant-deploy
    type: boolean
    description: Deploy immediately after creation
    required: false
    default: false
  - name: --list-types
    type: boolean
    description: List all available service types
    required: false
    default: false
  - name: --name
    type: string
    description: Service name
    required: false
  - name: --project-uuid
    type: string
    description: Project UUID (required)
    required: true
  - name: --server-uuid
    type: string
    description: Server UUID (required)
    required: true

Command: coolify service delete <uuid>
Description: Delete a service and optionally clean up its configurations, volumes, and networks.
Parameters:
  - name: --delete-configurations
    type: boolean
    description: Delete configurations
    required: false
    default: true
  - name: --delete-connected-networks
    type: boolean
    description: Delete connected networks
    required: false
    default: true
  - name: --delete-volumes
    type: boolean
    description: Delete volumes
    required: false
    default: true
  - name: --docker-cleanup
    type: boolean
    description: Run docker cleanup
    required: false
    default: true
  - name: --force (-f)
    type: boolean
    description: Skip confirmation prompt
    required: false
    default: false

Command: coolify service env create <service_uuid>
Description: Create a new environment variable for a specific service. Use --key and --value flags to specify the variable.
Parameters:
  - name: --build-time
    type: boolean
    description: Available at build time (default: true)
    required: false
    default: true
  - name: --comment
    type: string
    description: Comment for the environment variable
    required: false
  - name: --is-literal
    type: boolean
    description: Treat value as literal (don't interpolate variables)
    required: false
    default: false
  - name: --is-multiline
    type: boolean
    description: Value is multiline
    required: false
    default: false
  - name: --key
    type: string
    description: Environment variable key (required)
    required: true
  - name: --runtime
    type: boolean
    description: Available at runtime (default: true)
    required: false
    default: true
  - name: --value
    type: string
    description: Environment variable value (required)
    required: true

Command: coolify service env delete <service_uuid> <env_uuid>
Description: Delete an environment variable from a service. First UUID is the service, second is the specific environment variable to delete.
Parameters:
  - name: --force
    type: boolean
    description: Skip confirmation prompt
    required: false
    default: false

Command: coolify service env get <service_uuid> <env_uuid_or_key>
Description: Get detailed information about a specific environment variable. First UUID is the service, second is the environment variable UUID or key name.
Parameters: (None)

Command: coolify service env list <service_uuid>
Description: List all environment variables for a specific service.
Parameters: (None)

Command: coolify service env sync <service_uuid>
Description: Sync environment variables from a .env file
Parameters:
  - name: --build-time
    type: boolean
    description: Make all variables available at build time (default: true)
    required: false
    default: true
  - name: --file (-f)
    type: string
    description: Path to .env file (required)
    required: true
  - name: --is-literal
    type: boolean
    description: Treat all values as literal (don't interpolate variables)
    required: false
    default: false
  - name: --runtime
    type: boolean
    description: Make all variables available at runtime (default: true)
    required: false
    default: true

Command: coolify service env update <service_uuid> <env_uuid_or_key>
Description: Update an existing environment variable. Identify it by UUID or key name.
Parameters:
  - name: --build-time
    type: boolean
    description: Available at build time (default: true)
    required: false
    default: true
  - name: --comment
    type: string
    description: Comment for the environment variable
    required: false
  - name: --is-literal
    type: boolean
    description: Treat value as literal (don't interpolate variables)
    required: false
    default: false
  - name: --is-multiline
    type: boolean
    description: Value is multiline
    required: false
    default: false
  - name: --key
    type: string
    description: New environment variable key (rename)
    required: false
  - name: --runtime
    type: boolean
    description: Available at runtime (default: true)
    required: false
    default: true
  - name: --value
    type: string
    description: New environment variable value (required)
    required: true

Command: coolify service get <uuid>
Description: Get detailed information about a specific service.
Parameters: (None)

Command: coolify service list
Description: List all services in Coolify.
Parameters: (None)

Command: coolify service restart <uuid>
Description: Restart a service (restart all containers).
Parameters: (None)

Command: coolify service start <uuid>
Description: Start a service (deploy all containers).
Parameters: (None)

Command: coolify service stop <uuid>
Description: Stop a service (stop all containers).
Parameters: (None)

Command: coolify service storage create <service_uuid>
Description: Create a storage for a service
Parameters:
  - name: --content
    type: string
    description: File content (file only)
    required: false
  - name: --fs-path
    type: string
    description: Host directory path (file only, required when --is-directory is set)
    required: false
  - name: --host-path
    type: string
    description: Host path (persistent only)
    required: false
  - name: --is-directory
    type: boolean
    description: Whether this is a directory mount (file only)
    required: false
    default: false
  - name: --mount-path
    type: string
    description: Mount path inside the container (required)
    required: true
  - name: --name
    type: string
    description: Volume name (persistent only)
    required: false
  - name: --resource-uuid
    type: string
    description: UUID of the service sub-resource (required)
    required: true
  - name: --type
    type: string
    description: Storage type: 'persistent' or 'file' (required)
    required: true

Command: coolify service storage delete <service_uuid> <storage_uuid>
Description: Delete a storage from a service
Parameters: (None)

Command: coolify service storage list <service_uuid>
Description: List all persistent volumes and file storages for a specific service.
Parameters: (None)

Command: coolify service storage update <service_uuid>
Description: Update a storage for a service
Parameters:
  - name: --content
    type: string
    description: File content (file only)
    required: false
  - name: --host-path
    type: string
    description: Host path (persistent only)
    required: false
  - name: --id
    type: integer
    description: Storage ID (deprecated, use --uuid instead)
    required: false
    default: 0
  - name: --is-preview-suffix-enabled
    type: boolean
    description: Enable preview suffix for this storage
    required: false
    default: false
  - name: --mount-path
    type: string
    description: Mount path inside the container
    required: false
  - name: --name
    type: string
    description: Storage name (persistent only)
    required: false
  - name: --type
    type: string
    description: Storage type: 'persistent' or 'file' (required)
    required: true
  - name: --uuid
    type: string
    description: Storage UUID (required, use 'storage list' to find)
    required: false

Command: coolify teams current
Description: Get details of the team associated with the current authentication token.
Parameters: (None)

Command: coolify teams get <team_id>
Description: Get detailed information about a specific team by its ID.
Parameters: (None)

Command: coolify teams list
Description: List all teams you have access to.
Parameters: (None)

Command: coolify teams members list [team_id]
Description: List members of a specific team by ID, or list members of the current team if no ID is provided.
Parameters: (None)

Command: coolify update
Description: Update Coolify CLI
Parameters: (None)

Command: coolify version
Description: Current Coolify CLI version
Parameters: (None)

