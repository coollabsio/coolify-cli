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
- `coolify servers get` - Get a server
  - `--resources` - Get the resources and their status of a server