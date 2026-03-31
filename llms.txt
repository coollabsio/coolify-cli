# Coolify CLI - llms.txt

> Quick AI/LLM instructions for the Coolify CLI.
> Source: https://github.com/coollabsio/coolify-cli
> API Spec: https://github.com/coollabsio/coolify/blob/v4.x/openapi.json

## Operating Rules

- Prefer `--format json` for automation and parsing.
- Use Coolify UUIDs for resources; do not use internal numeric IDs.
- Team commands are the exception: they use numeric team IDs.
- Authenticate with a saved context when possible; use `--token` only for overrides.
- Use `llms-full.txt` for the exhaustive command/flag catalog.

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
4. Switch contexts with `coolify context use <context_name>`

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

## Global Flags

- `--context <name>` - use a specific saved context
- `--token <token>` - override token from config
- `--format table|json|pretty` - choose output format
- `--show-sensitive` - reveal sensitive values
- `--debug` - enable debug output

## Common Workflows

### Contexts

```bash
coolify context list
coolify context verify
coolify context version
coolify context use prod
```

### Inventory

```bash
coolify server list
coolify project list
coolify resource list
coolify app list
coolify service list
coolify database list
```

### Applications

```bash
coolify app get <uuid>
coolify app start <uuid>
coolify app stop <uuid>
coolify app restart <uuid>
coolify app logs <uuid> --follow
coolify app deployments list <app-uuid>
coolify app deployments logs <app-uuid> --follow
```

### Environment Variables

```bash
coolify app env list <app-uuid>
coolify app env create <app-uuid> --key API_KEY --value secret123
coolify app env update <app-uuid> <env-uuid-or-key> --value new-secret
coolify app env sync <app-uuid> --file .env.production --build-time --preview
```

### Deployments

```bash
coolify deploy list
coolify deploy name my-application
coolify deploy batch api,worker,frontend --force
coolify deploy cancel <deployment-uuid>
```

### Databases and Services

```bash
coolify database get <uuid>
coolify database create postgresql --server-uuid <uuid> --project-uuid <uuid> --environment-name production
coolify database backup list <database-uuid>
coolify service get <uuid>
coolify service create <type> --project-uuid <uuid> --server-uuid <uuid> --instant-deploy
```

## Common Aliases

- `coolify app` | `coolify apps` | `coolify application` | `coolify applications`
- `coolify service` | `coolify services` | `coolify svc`
- `coolify database` | `coolify databases` | `coolify db` | `coolify dbs`
- `coolify teams` | `coolify team`

## Full Reference

- Full command and parameter catalog: ./llms-full.txt
- Regenerate docs: `go run ./coolify docs llms`
