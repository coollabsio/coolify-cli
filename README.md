# CLI for [Coolify](https://coolify.io) API

> [!WARNING]
> Until version 1.0.0, the CLI should be considered unstable. Any minor or patch release may introduce breaking changes. Please read the release notes carefully before updating.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/coollabsio/cli-coolify/main/scripts/install.sh | bash
```

This will install the CLI in `/usr/local/bin/coolify`.

> If you are a Windows or macOS user, please test the installation script and let us know if it works for you.

## Initial Setup

Before using any commands, you need to initialize the CLI by creating a configuration file:

```bash
coolify init
```

This interactive wizard will guide you through setting up your Coolify instance(s). You can choose to:
- Connect to Coolify Cloud using your API token
- Add self-hosted Coolify instance(s) with their FQDN and token

Alternatively, you can generate a default configuration non-interactively:

```bash
coolify init --default
```

The configuration will be stored in `~/.config/coolify/config.json`.

## Getting Your API Token

To use the CLI, you'll need an API token:
1. Log in to your Coolify dashboard (Cloud or self-hosted)
2. Navigate to `/security/api-tokens`
3. Create a new token with appropriate permissions
4. Use this token when initializing the CLI or adding a new instance

## Managing Instances

After initialization, you can manage your Coolify instances:

### Add a New Instance

```bash
coolify instances add MyInstance https://my.instance.tld mytoken
```

Or use the interactive mode:

```bash
coolify instances add
```

### List All Instances

```bash
coolify instances list
```

### Set Default Instance

```bash
coolify instances set default MyInstance
```

### Remove an Instance

```bash
coolify instances remove MyInstance
```

### Update Instance Token

```bash
coolify instances set token MyInstance newtoken
```

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

### Deployments
- `coolify deploy uuid <uuid>` - Deploy a resource by UUID
  - `--force` - Force deployment
- `coolify deploy name <name>` - Deploy a resource by name (NEW)
  - `--force` - Force deployment
- `coolify deploy batch <name1,name2,...>` - Deploy multiple resources at once (NEW)
  - `--force` - Force all deployments

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
