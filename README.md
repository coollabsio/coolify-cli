# CLI for [Coolify](https://coolify.io) API

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/coollabsio/coolify-cli/main/scripts/install.sh | bash
```

It will install the CLI in `/usr/local/bin/coolify` and the configuration file in `~/.config/coolify/config.json`

## Configuration
1. Get a `<token>` from your Coolify dashboard (Cloud or self-hosted) at `/security/api-tokens`

If you are using the Cloud version:
2. Add the token with `coolify instances set token <token> https://app.coolify.io`

If you are using a self-hosted version:
2. Add the token with `coolify instances add -d <fqdn> <token>`
   
> Replace `<fqdn>` with the fully qualified domain name of your Coolify instance

Now you can use the CLI with the token you just added.

## Change default instance
You can change the default instance with `coolify instances set default <fqdn|linenumber>`
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