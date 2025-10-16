# CLI for [Coolify](https://coolify.io) API

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/coollabsio/coolify-cli/main/scripts/install.sh | bash
```

It will install the CLI in `/usr/local/bin/coolify` and the configuration file in `~/.config/coolify/config.json`

> If you are a windows or mac user, please test the installation script and let us know if it works for you.

## Configuration
1. Get a `<token>` from your Coolify dashboard (Cloud or self-hosted) at `/security/api-tokens`

### Cloud

2. Add the token with `coolify instances set token cloud <token>`

### Self-hosted

2. Add the token with `coolify instances add -d <name> <fqdn> <token>`
   
> Replace `<name>` with the name you want to give to the instance.
>
> Replace `<fqdn>` with the fully qualified domain name of your Coolify instance.

Now you can use the CLI with the token you just added.

## Change default instance
You can change the default instance with `coolify instances set default <name>`
## Currently Supported Commands

### Update
- `coolify update` - Update the CLI to the latest version

### Instances
- `coolify instances list` - List all instances
- `coolify instances add` - Create a new instance configuration
- `coolify instances remove` - Remove an instance configuration
- `coolify instances get` - Get an instance configuration
- `coolify instances set <default>|<token>` - Set an instance as default or set a token for an instance
- `coolify instances version` - Get the version of the Coolify API for an instance

### Servers
- `coolify servers list` - List all servers
- `coolify servers get <uuid>` - Get a server by UUID
  - `--resources` - Get the resources and their status of a server
- `coolify servers add <name> <ip> <private_key_uuid>` - Add a new server
  - `--port <port>` - SSH port (default: 22)
  - `--user <user>` - SSH user (default: root)
  - `--validate` - Validate server immediately after adding
- `coolify servers remove <uuid>` - Remove a server
- `coolify servers validate <uuid>` - Validate a server connection

### Projects
- `coolify projects list` - List all projects
- `coolify projects get <uuid>` - Get project environments

### Resources
- `coolify resources list` - List all resources

### Applications
- `coolify app list` - List all applications
- `coolify app get <uuid>` - Get application details
- `coolify app update <uuid>` - Update application configuration
  - `--name <name>` - Application name
  - `--description <description>` - Application description
- `coolify app delete <uuid>` - Delete an application
  - `--force` - Skip confirmation prompt
- `coolify app start <uuid>` - Start an application
- `coolify app stop <uuid>` - Stop an application
- `coolify app restart <uuid>` - Restart an application
- `coolify app logs <uuid>` - Get application logs

#### Application Environment Variables
- `coolify app env list <app_uuid>` - List all environment variables
- `coolify app env get <app_uuid> <env_uuid_or_key>` - Get a specific environment variable
- `coolify app env create <app_uuid>` - Create a new environment variable
  - `--key <key>` - Variable key (required)
  - `--value <value>` - Variable value (required)
  - `--is-preview` - Set variable for preview environments
  - `--is-build-time` - Set variable as build-time variable
  - `--is-literal` - Treat value as literal (no variable expansion)
  - `--is-multiline` - Allow multiline values
  - `--is-shown-once` - Show value only once (for secrets)
- `coolify app env update <app_uuid> <env_uuid>` - Update an environment variable
- `coolify app env delete <app_uuid> <env_uuid>` - Delete an environment variable
- `coolify app env sync <app_uuid>` - Sync environment variables from a .env file
  - `--file <path>` - Path to .env file (required)

### Databases
- `coolify database list` - List all databases
- `coolify database get <uuid>` - Get database details
- `coolify database create <type>` - Create a new database
  - Supported types: `postgresql`, `mysql`, `mariadb`, `mongodb`, `redis`, `keydb`, `clickhouse`, `dragonfly`
  - `--server-uuid <uuid>` - Server UUID (required)
  - `--project-uuid <uuid>` - Project UUID (required)
  - `--name <name>` - Database name
  - `--description <description>` - Database description
  - `--image <image>` - Docker image
  - `--instant-deploy` - Deploy immediately after creation
  - `--is-public` - Make database publicly accessible
  - `--public-port <port>` - Public port number
  - Database-specific flags (postgres-user, mysql-root-password, etc.)
- `coolify database update <uuid>` - Update database configuration
- `coolify database delete <uuid>` - Delete a database
  - `--delete-configurations` - Delete configurations (default: true)
  - `--delete-volumes` - Delete volumes (default: true)
  - `--docker-cleanup` - Run docker cleanup (default: true)
- `coolify database start <uuid>` - Start a database
- `coolify database stop <uuid>` - Stop a database
- `coolify database restart <uuid>` - Restart a database

#### Database Backups
- `coolify database backup list <database_uuid>` - List all backup configurations
- `coolify database backup create <database_uuid>` - Create a new backup configuration
  - `--frequency <cron>` - Backup frequency (cron expression)
  - `--enabled` - Enable backup schedule
  - `--save-s3` - Save backups to S3
  - `--s3-storage-uuid <uuid>` - S3 storage UUID
  - `--retention-amount-locally <n>` - Number of backups to retain locally
  - `--retention-days-locally <n>` - Days to retain backups locally
  - `--timeout <seconds>` - Backup timeout
- `coolify database backup update <database_uuid> <backup_uuid>` - Update a backup configuration
- `coolify database backup delete <database_uuid> <backup_uuid>` - Delete a backup configuration
- `coolify database backup trigger <database_uuid> <backup_uuid>` - Trigger an immediate backup
- `coolify database backup executions <database_uuid> <backup_uuid>` - List backup executions
- `coolify database backup delete-execution <database_uuid> <backup_uuid> <execution_uuid>` - Delete a backup execution

### Services
- `coolify service list` - List all services
- `coolify service get <uuid>` - Get service details
- `coolify service start <uuid>` - Start a service
- `coolify service stop <uuid>` - Stop a service
- `coolify service restart <uuid>` - Restart a service
- `coolify service delete <uuid>` - Delete a service

#### Service Environment Variables
- `coolify service env list <service_uuid>` - List all environment variables
- `coolify service env get <service_uuid> <env_uuid_or_key>` - Get a specific environment variable
- `coolify service env create <service_uuid>` - Create a new environment variable
  - Same flags as application environment variables
- `coolify service env update <service_uuid> <env_uuid>` - Update an environment variable
- `coolify service env delete <service_uuid> <env_uuid>` - Delete an environment variable
- `coolify service env sync <service_uuid>` - Sync environment variables from a .env file
  - `--file <path>` - Path to .env file (required)

### Deployments
- `coolify deploy uuid <uuid>` - Deploy a resource by UUID
  - `--force` - Force deployment
- `coolify deploy name <name>` - Deploy a resource by name
  - `--force` - Force deployment
- `coolify deploy batch <name1,name2,...>` - Deploy multiple resources at once
  - `--force` - Force all deployments
- `coolify deploy list` - List all deployments
- `coolify deploy get <uuid>` - Get deployment details
- `coolify deploy cancel <uuid>` - Cancel a deployment
  - `--force` - Skip confirmation prompt

### GitHub Apps
- `coolify github list` - List all GitHub App integrations
- `coolify github get <app_uuid>` - Get GitHub App details
- `coolify github create` - Create a new GitHub App integration
  - `--name <name>` - GitHub App name (required)
  - `--api-url <url>` - GitHub API URL (required)
  - `--html-url <url>` - GitHub HTML URL (required)
  - `--app-id <id>` - GitHub App ID (required)
  - `--installation-id <id>` - Installation ID (required)
  - `--client-id <id>` - OAuth Client ID (required)
  - `--client-secret <secret>` - OAuth Client Secret (required)
  - `--private-key-uuid <uuid>` - Private key UUID (required)
  - `--organization <org>` - GitHub organization
  - `--custom-user <user>` - Custom SSH user
  - `--custom-port <port>` - Custom SSH port
  - `--webhook-secret <secret>` - Webhook secret
  - `--system-wide` - System-wide installation
- `coolify github update <app_uuid>` - Update a GitHub App
- `coolify github delete <app_uuid>` - Delete a GitHub App
  - `--force` - Skip confirmation prompt
- `coolify github repos <app_uuid>` - List repositories accessible by a GitHub App
- `coolify github branches <app_uuid> <owner/repo>` - List branches for a repository

### Teams
- `coolify team list` - List all teams
- `coolify team get <id>` - Get team details
- `coolify team current` - Get current team
- `coolify team members list [team_id]` - List team members

### Domains
- `coolify domains list` - List all domains

### Private Keys
- `coolify privatekeys list` - List all private keys
- `coolify privatekeys create <name> <private-key>` - Create a new private key
  - Use `@filename` to read from file: `coolify privatekeys create mykey @~/.ssh/id_rsa`
- `coolify privatekeys delete <uuid>` - Delete a private key

## Global Flags

All commands support these global flags:

- `--instance <name>` - Use a specific instance profile instead of default (NEW)
- `--host <fqdn>` - Override the Coolify instance hostname
- `--token <token>` - Override the authentication token
- `--format <format>` - Output format: `table` (default), `json`, or `pretty`
- `--show-sensitive` / `-s` - Show sensitive information (tokens, IPs, etc.)
- `--force` / `-f` - Force operation (skip confirmations)
- `--debug` - Enable debug mode

## Examples

### Multi-Environment Workflows

```bash
# Add multiple instances
coolify instances add prod https://prod.coolify.io <prod-token>
coolify instances add staging https://staging.coolify.io <staging-token>
coolify instances add dev https://dev.coolify.io <dev-token>

# Set default
coolify instances set default prod

# Use different profiles
coolify --instance=staging servers list
coolify --instance=prod deploy name api
coolify --instance=dev resources list

# Default profile (prod in this case)
coolify servers list
```

### Application Management

```bash
# List all applications
coolify app list

# Get application details
coolify app get <uuid>

# Manage application lifecycle
coolify app start <uuid>
coolify app stop <uuid>
coolify app restart <uuid>

# View application logs
coolify app logs <uuid>

# Environment variables
coolify app env list <uuid>
coolify app env create <uuid> --key API_KEY --value secret123
coolify app env sync <uuid> --file .env
```

### Database Management

```bash
# List databases
coolify database list

# Create a PostgreSQL database
coolify database create postgresql \
  --server-uuid <server-uuid> \
  --project-uuid <project-uuid> \
  --name mydb \
  --instant-deploy

# Manage database lifecycle
coolify database start <uuid>
coolify database stop <uuid>
coolify database restart <uuid>

# Backup management
coolify database backup list <database-uuid>
coolify database backup create <database-uuid> \
  --frequency "0 2 * * *" \
  --enabled \
  --save-s3 \
  --retention-days-locally 7
coolify database backup trigger <database-uuid> <backup-uuid>
```

### Service Management

```bash
# List services
coolify service list

# Get service details
coolify service get <uuid>

# Manage services
coolify service start <uuid>
coolify service restart <uuid>

# Environment variables (same as applications)
coolify service env sync <uuid> --file .env
```

### Deploy Workflows

```bash
# Deploy single app by name (easier than UUID)
coolify deploy name my-application

# Deploy multiple apps at once
coolify deploy batch api,worker,frontend

# Force deploy with specific profile
coolify --instance=prod deploy batch api,worker --force

# Traditional UUID deployment still works
coolify deploy uuid abc123-def456-...

# Monitor deployments
coolify deploy list
coolify deploy get <deployment-uuid>

# Cancel a deployment
coolify deploy cancel <deployment-uuid>
```

### GitHub Apps Integration

```bash
# List GitHub Apps
coolify github list

# Create a GitHub App integration
coolify github create \
  --name "My GitHub App" \
  --api-url "https://api.github.com" \
  --html-url "https://github.com" \
  --app-id 123456 \
  --installation-id 789012 \
  --client-id "Iv1.abc123" \
  --client-secret "secret" \
  --private-key-uuid <key-uuid>

# List repositories accessible by the app
coolify github repos <app-uuid>

# List branches for a repository
coolify github branches <app-uuid> owner/repo

# Delete a GitHub App
coolify github delete <app-uuid>
```

### Team Management

```bash
# List teams
coolify team list

# Get current team
coolify team current

# List team members
coolify team members list
```

### Server Management

```bash
# List servers in production
coolify --instance=prod servers list

# Add a server with validation
coolify servers add myserver 192.168.1.100 <key-uuid> --validate

# Get server details with resources
coolify servers get <uuid> --resources
```

## Output Formats

The CLI supports three output formats:

```bash
# Table format (default, human-readable)
coolify servers list

# JSON format (for scripts)
coolify servers list --format=json

# Pretty JSON (for debugging)
coolify servers list --format=pretty
```

## Architecture

This CLI follows a clean architecture with:
- **Service Layer**: Business logic and API interactions
- **Output Layer**: Consistent formatting across all commands
- **Config Layer**: Multi-instance configuration management
- **Models Layer**: Type-safe data structures

## Development

```bash
# Build
go build -o coolify .

# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Install locally
go install
```

## Contributing

Contributions are welcome! Please check the [restructure documentation](RESTRUCTURE_PLAN.md) for architecture guidelines.

## License

MIT
