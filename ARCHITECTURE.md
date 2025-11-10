# Coolify CLI Architecture

This document describes the architecture and design principles of the Coolify CLI.

## Overview

The Coolify CLI is a command-line interface for managing Coolify instances, servers, projects, and deployments. It follows a layered architecture pattern that separates concerns and promotes maintainability.

## Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│                    User Interface                       │
│                   (Terminal/Shell)                      │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                  Command Layer (cmd/)                   │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐             │
│  │ servers  │  │  deploy  │  │ projects │  ...        │
│  └──────────┘  └──────────┘  └──────────┘             │
│  • CLI parsing & validation                             │
│  • Flag handling                                        │
│  • Output formatting                                    │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                 Service Layer (internal/service/)       │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ServerService│  │DeployService │  │ProjectService│  │
│  └─────────────┘  └──────────────┘  └──────────────┘  │
│  • Business logic                                       │
│  • Request validation                                   │
│  • Response transformation                              │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                API Client Layer (internal/api/)         │
│  ┌───────────────────────────────────────────────────┐ │
│  │              HTTP Client (api.Client)             │ │
│  └───────────────────────────────────────────────────┘ │
│  • HTTP requests/responses                              │
│  • Authentication (Bearer tokens)                       │
│  • Retry logic                                          │
│  • Error handling                                       │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                 Coolify API (External)                  │
│              https://instance.coolify.io/api/v1/        │
└─────────────────────────────────────────────────────────┘
```

## Supporting Components

```
┌─────────────────────────────────────────────────────────┐
│         Configuration (internal/config/)                │
│  • Multi-instance management                            │
│  • Default instance selection                           │
│  • Token storage                                        │
│  • ~/.config/coolify/config.json                        │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│          Output Formatters (internal/output/)           │
│  ┌─────────┐  ┌────────┐  ┌─────────┐                 │
│  │  Table  │  │  JSON  │  │ Pretty  │                 │
│  └─────────┘  └────────┘  └─────────┘                 │
│  • Flexible output formats                              │
│  • Sensitive data masking                               │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│           Data Models (internal/models/)                │
│  • Server, Project, Resource, Deployment               │
│  • Request/Response structures                          │
│  • JSON marshaling/unmarshaling                         │
└─────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### 1. Command Layer (`cmd/`)

**Purpose**: Handle CLI user interface and interaction

**Responsibilities**:
- Parse command-line arguments and flags
- Validate user input
- Coordinate with service layer
- Format and display output
- Handle errors gracefully

**Key Files**:
- `root.go` - Root command, global flags, initialization
- `servers.go` - Server management commands
- `deploy.go` - Deployment commands
- `context.go` - Context (instance) configuration commands
- `projects.go` - Project listing and inspection
- etc.

**Example**:
```go
var serversListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all servers",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Get API client
        client, err := getAPIClient(cmd)
        if err != nil {
            return err
        }

        // Use service layer
        service := service.NewServerService(client)
        servers, err := service.List(cmd.Context())
        if err != nil {
            return err
        }

        // Format and display output
        formatter, _ := getFormatter(cmd)
        return formatter.Format(servers)
    },
}
```

### 2. Service Layer (`internal/service/`)

**Purpose**: Implement business logic and coordinate API calls

**Responsibilities**:
- Validate business rules
- Coordinate multiple API calls if needed
- Transform API responses to CLI-friendly format
- Handle service-specific error cases

**Key Files**:
- `server.go` - Server operations
- `deployment.go` - Deployment operations
- `project.go` - Project operations
- `resource.go` - Resource operations
- `privatekey.go` - SSH key operations
- `domain.go` - Domain operations

**Example**:
```go
type ServerService struct {
    client *api.Client
}

func (s *ServerService) List(ctx context.Context) ([]models.Server, error) {
    var servers []models.Server
    err := s.client.Get(ctx, "servers", &servers)
    return servers, err
}
```

### 3. API Client Layer (`internal/api/`)

**Purpose**: Handle all HTTP communication with Coolify API

**Responsibilities**:
- Construct HTTP requests
- Add authentication headers
- Retry failed requests with exponential backoff
- Parse HTTP responses
- Convert HTTP errors to meaningful error messages

**Key Files**:
- `client.go` - HTTP client implementation
- `error.go` - API error handling
- `options.go` - Client configuration options

**Example**:
```go
type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
    retries    int
    timeout    time.Duration
}

func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
    return c.doRequest(ctx, "GET", path, nil, result)
}
```

### 4. Configuration Layer (`internal/config/`)

**Purpose**: Manage CLI configuration and multiple instances

**Responsibilities**:
- Load/save configuration from disk
- Manage multiple Coolify instances
- Select default instance
- Store API tokens securely (file permissions)

**Key Files**:
- `config.go` - Configuration structure and methods
- `instance.go` - Instance definition
- `loader.go` - File I/O operations

**Configuration File** (`~/.config/coolify/config.json`):
```json
{
  "instances": [
    {
      "name": "prod",
      "fqdn": "https://coolify.example.com",
      "token": "your-api-token",
      "default": true
    },
    {
      "name": "staging",
      "fqdn": "https://staging.coolify.example.com",
      "token": "staging-token"
    }
  ]
}
```

### 5. Output Layer (`internal/output/`)

**Purpose**: Format data for display to users

**Responsibilities**:
- Format data as tables, JSON, or pretty-printed JSON
- Hide sensitive information unless `--show-sensitive` is used
- Handle different data types (slices, structs, primitives)

**Key Files**:
- `formatter.go` - Formatter interface
- `table.go` - Table formatting
- `json.go` - JSON formatting
- `pretty.go` - Pretty JSON formatting

**Supported Formats**:
- `table` - Default, human-readable tables
- `json` - Compact JSON for scripting
- `pretty` - Indented JSON for debugging

### 6. Models Layer (`internal/models/`)

**Purpose**: Define data structures

**Responsibilities**:
- Define API request/response structures
- JSON tags for marshaling
- Common types and timestamps

**Key Files**:
- `server.go` - Server-related types
- `project.go` - Project-related types
- `resource.go` - Resource types
- `deployment.go` - Deployment types
- `common.go` - Shared types

## Data Flow

### Example: Listing Servers

1. **User Input**: `coolify servers list --format=table`

2. **Command Layer** (`cmd/servers.go`):
   - Cobra parses the command
   - `serversListCmd.RunE` is executed
   - Gets API client using `getAPIClient()`
   - Creates ServerService instance

3. **Service Layer** (`internal/service/server.go`):
   - `ServerService.List()` is called
   - Validates context (if needed)
   - Calls API client

4. **API Client Layer** (`internal/api/client.go`):
   - Constructs GET request to `/api/v1/servers`
   - Adds Bearer token authentication
   - Sends HTTP request
   - Retries on failure (with backoff)
   - Parses JSON response

5. **Response Processing**:
   - JSON unmarshaled to `[]models.Server`
   - Returns to service layer
   - Returns to command layer

6. **Output Layer** (`internal/output/table.go`):
   - Command layer creates table formatter
   - Formatter processes server data
   - Formats as table with columns
   - Writes to stdout

7. **User Output**: Table displayed in terminal

## Design Patterns

### 1. Dependency Injection

Services receive the API client as a constructor parameter:

```go
func NewServerService(client *api.Client) *ServerService {
    return &ServerService{client: client}
}
```

**Benefits**:
- Easy to test (can inject mock client)
- Clear dependencies
- Flexible configuration

### 2. Strategy Pattern (Output Formatters)

Different formatters implement the same interface:

```go
type Formatter interface {
    Format(data interface{}) error
}
```

**Benefits**:
- Easy to add new formats
- Consistent API
- Runtime format selection

### 3. Options Pattern (API Client)

Client configuration uses functional options:

```go
client := api.NewClient(url, token,
    api.WithDebug(true),
    api.WithRetries(5),
    api.WithTimeout(60 * time.Second),
)
```

**Benefits**:
- Optional parameters
- Clear intent
- Backward compatible

### 4. Error Wrapping

Errors are wrapped with context at each layer:

```go
if err != nil {
    return fmt.Errorf("failed to list servers: %w", err)
}
```

**Benefits**:
- Error context preserved
- Stack trace maintained
- Better debugging

## Testing Strategy

### Unit Tests

Each layer has comprehensive unit tests:

- **Commands**: Mock services, test flag parsing
- **Services**: Mock API client, test business logic
- **API Client**: Use `httptest.Server`, test HTTP handling
- **Config**: Test file I/O with temp directories
- **Output**: Test formatting with buffers

### Integration Tests

Test multiple layers together:

- Commands + Services + Mock API
- Config + File System
- End-to-end workflows

### Coverage Goals

- Overall: 70%+
- New features: 80%+
- Critical paths: 90%+

## Configuration Files

### CLI Configuration

**Location**: `~/.config/coolify/config.json` (Linux/macOS)
**Location**: `%APPDATA%\coolify\config.json` (Windows)

**Structure**:
```json
{
  "instances": [
    {
      "name": "prod",
      "fqdn": "https://coolify.example.com",
      "token": "your-token",
      "default": true
    }
  ],
  "lastUpdateCheckTime": "2025-01-15T10:30:00Z"
}
```

## API Communication

### Base URL

All API calls use: `{fqdn}/api/v1/{endpoint}`

Example: `https://coolify.example.com/api/v1/servers`

### Authentication

Bearer token authentication:

```
Authorization: Bearer {token}
```

### Request/Response

**Content-Type**: `application/json`

**Request Body** (POST):
```json
{
  "name": "my-server",
  "ip": "192.168.1.100"
}
```

**Response Body**:
```json
{
  "uuid": "abc123",
  "name": "my-server",
  "ip": "192.168.1.100"
}
```

### Error Handling

HTTP errors are converted to CLI-friendly messages:

- `401` → "Unauthenticated. Check your API token."
- `404` → "Resource not found."
- `500` → "Server error. Please try again."

### Retry Logic

Failed requests are retried with exponential backoff:

- Attempt 1: Immediate
- Attempt 2: Wait 1s
- Attempt 3: Wait 2s
- Attempt 4: Wait 4s

Does not retry on 4xx errors (except 429 rate limit).

## Security Considerations

### API Token Storage

- Stored in config file with restricted permissions (0600)
- Never logged (even in debug mode)
- Masked in output by default (use `-s` to show)

### Sensitive Data Handling

- Tokens masked as `********` in output
- Use `--show-sensitive` flag to reveal
- Debug logs sanitize sensitive data

### HTTPS

- All API communication uses HTTPS
- Certificate validation enabled

## Performance Optimizations

### Concurrent Operations

Batch deployments run in parallel:

```go
// Deploy multiple resources concurrently
var wg sync.WaitGroup
for _, name := range names {
    wg.Add(1)
    go func(n string) {
        defer wg.Done()
        deployResource(n)
    }(name)
}
wg.Wait()
```

### Connection Reuse

HTTP client reuses connections:

```go
c.httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:       10,
        IdleConnTimeout:    90 * time.Second,
    },
}
```

### Minimal Dependencies

- Use Go standard library when possible
- Only essential external dependencies
- Keep binary size small

## Extensibility

### Adding a New Command

1. Create `cmd/newfeature.go`
2. Define Cobra command
3. Create service if needed (`internal/service/newfeature.go`)
4. Add models if needed (`internal/models/newfeature.go`)
5. Register command in `init()`
6. Write tests

### Adding a New Output Format

1. Create `internal/output/newformat.go`
2. Implement `Formatter` interface
3. Add format constant
4. Update `NewFormatter()` switch

### Adding API Client Features

1. Add method to `internal/api/client.go`
2. Add tests in `internal/api/client_test.go`
3. Use in service layer

## Build & Release

### Build Process

```bash
# Local build
go build -o coolify ./coolify

# Install locally
go install ./coolify

# Multi-platform release
goreleaser release --clean
```

### Release Artifacts

- Linux: amd64, arm64
- macOS: amd64, arm64 (Apple Silicon)
- Windows: amd64

### Distribution

- GitHub Releases
- Install script: `scripts/install.sh`
- Package managers (planned)

## Future Enhancements

- [ ] Shell completion improvements
- [ ] Interactive mode
- [ ] Configuration wizard
- [ ] Plugin system
- [ ] Telemetry (opt-in)
- [ ] Cache layer for frequent queries

## References

- [Cobra Documentation](https://cobra.dev/)
- [Coolify API Specification](https://github.com/coollabsio/coolify/blob/v4.x/openapi.json)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
