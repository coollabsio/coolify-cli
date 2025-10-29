# Contributing to Coolify CLI

Thank you for your interest in contributing to the Coolify CLI! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Architecture](#project-architecture)
- [Adding a New Command](#adding-a-new-command)
- [Testing Requirements](#testing-requirements)
- [Code Style & Conventions](#code-style--conventions)
- [Submitting Changes](#submitting-changes)

## Getting Started

Before you start contributing:

1. **Read the [ARCHITECTURE.md](ARCHITECTURE.md)** for detailed architectural guidance
2. **Review the [OpenAPI specification](https://github.com/coollabsio/coolify/blob/v4.x/openapi.json)** to understand available API endpoints
3. **Check existing issues** to see if your feature/bug is already being worked on
4. **Open an issue** to discuss your proposed changes (for large features)

### Prerequisites

- Go 1.24 or higher
- Git

## Development Setup

### Clone and Build

```bash
# Fork the repository on GitHub
# Clone your fork
git clone https://github.com/YOUR_USERNAME/coolify-cli.git
cd coolify-cli

# Build the CLI
go build -o coolify ./coolify

# Install locally
go install
```

### Running the CLI

```bash
# Run without installing
go run ./coolify [command]

# Example commands
go run ./coolify context list
go run ./coolify server list --debug

# With flags
go run ./coolify server list --format json --debug
```

### Project Structure

```
cmd/                 # CLI commands (organized by feature)
├── root.go          # Root command and global flags
├── application/     # Application management commands
├── context/         # Manage Coolify instances
├── server/          # Server management
├── project/         # Project management
├── database/        # Database management
├── deployment/      # Deployment operations
├── service/         # Service management
└── ...

internal/            # Internal packages
├── api/             # API client (HTTP communication)
├── cli/             # CLI utilities (GetAPIClient helper)
├── config/          # Configuration management
├── models/          # Data models and structs
├── output/          # Output formatters (table, json, pretty)
├── parser/          # Input parsing utilities
├── service/         # Business logic layer
└── version/         # Version management

test/                # Test utilities and fixtures
└── fixtures/        # Mock API response data
```

## Project Architecture

The Coolify CLI follows a **layered architecture**:

```
User → Commands (cmd/) → Services (internal/service/) → API Client (internal/api/) → Coolify API
```

### Layer Responsibilities

1. **Command Layer** (`cmd/`)
   - Parse CLI arguments and flags
   - Call service layer methods
   - Format output using output formatters

2. **Service Layer** (`internal/service/`)
   - Business logic
   - Coordinate API calls
   - Transform data

3. **API Client Layer** (`internal/api/`)
   - HTTP communication
   - Retry logic with exponential backoff
   - Authentication (Bearer tokens)
   - Error handling

### Key Dependencies

- **cobra**: CLI framework
- **viper**: Configuration management
- **stretchr/testify**: Testing assertions

## Adding a New Command

Follow these steps to add a new command:

### 1. Create Command Directory Structure

```bash
# Create directory for your command
mkdir -p cmd/myfeature
```

### 2. Create Parent Command

Create `cmd/myfeature/myfeature.go`:

```go
package myfeature

import "github.com/spf13/cobra"

// NewMyFeatureCommand creates the myfeature parent command
func NewMyFeatureCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:     "myfeature",
        Aliases: []string{"mf"},
        Short:   "MyFeature related commands",
        Long:    `Manage MyFeature resources.`,
    }

    // Add subcommands
    cmd.AddCommand(NewListCommand())
    cmd.AddCommand(NewGetCommand())
    // ... more subcommands

    return cmd
}
```

### 3. Create Subcommand

Create `cmd/myfeature/list.go`:

```go
package myfeature

import (
    "context"
    "fmt"

    "github.com/coollabsio/coolify-cli/internal/cli"
    "github.com/coollabsio/coolify-cli/internal/output"
    "github.com/coollabsio/coolify-cli/internal/service"
    "github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "list",
        Short: "List all myfeature resources",
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()

            // Get API client
            client, err := cli.GetAPIClient(cmd)
            if err != nil {
                return fmt.Errorf("failed to get API client: %w", err)
            }

            // Use service layer
            svc := service.NewMyFeatureService(client)
            items, err := svc.List(ctx)
            if err != nil {
                return fmt.Errorf("failed to list items: %w", err)
            }

            // Format output
            format, _ := cmd.Flags().GetString("format")
            showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

            formatter, err := output.NewFormatter(format, output.Options{
                ShowSensitive: showSensitive,
            })
            if err != nil {
                return err
            }

            return formatter.Format(items)
        },
    }
}
```

### 4. Create Service Layer

Create `internal/service/myfeature.go`:

```go
package service

import (
    "context"

    "github.com/coollabsio/coolify-cli/internal/api"
    "github.com/coollabsio/coolify-cli/internal/models"
)

type MyFeatureService struct {
    client *api.Client
}

func NewMyFeatureService(client *api.Client) *MyFeatureService {
    return &MyFeatureService{client: client}
}

func (s *MyFeatureService) List(ctx context.Context) ([]models.MyFeature, error) {
    var items []models.MyFeature
    err := s.client.Get(ctx, "myfeature", &items)
    return items, err
}

func (s *MyFeatureService) Get(ctx context.Context, uuid string) (*models.MyFeature, error) {
    var item models.MyFeature
    err := s.client.Get(ctx, "myfeature/"+uuid, &item)
    return &item, err
}

func (s *MyFeatureService) Create(ctx context.Context, req models.MyFeatureCreateRequest) (*models.Response, error) {
    var response models.Response
    err := s.client.Post(ctx, "myfeature", req, &response)
    return &response, err
}

func (s *MyFeatureService) Delete(ctx context.Context, uuid string) error {
    return s.client.Delete(ctx, "myfeature/"+uuid)
}
```

### 5. Create Models

Create `internal/models/myfeature.go`:

```go
package models

type MyFeature struct {
    ID          int    `json:"id" table:"-"`           // Hidden from table output
    UUID        string `json:"uuid"`                   // Shown to users
    Name        string `json:"name"`
    Description string `json:"description"`
    Status      string `json:"status"`
    // Add more fields...
}

type MyFeatureCreateRequest struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}
```

**Important**: Always use `UUID` for user-facing identifiers, not database `ID`. Hide `ID` field from table output using `table:"-"` tag.

### 6. Register Command

Add your command to `cmd/root.go`:

```go
import (
    // ... existing imports
    "github.com/coollabsio/coolify-cli/cmd/myfeature"
)

func init() {
    // ... existing code
    rootCmd.AddCommand(myfeature.NewMyFeatureCommand())
}
```

### 7. Create Tests

Create `internal/service/myfeature_test.go`:

```go
package service

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/coollabsio/coolify-cli/internal/api"
    "github.com/coollabsio/coolify-cli/internal/models"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMyFeatureService_List(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/api/v1/myfeature", r.URL.Path)
        assert.Equal(t, "GET", r.Method)

        items := []models.MyFeature{
            {UUID: "uuid-1", Name: "item-1"},
            {UUID: "uuid-2", Name: "item-2"},
        }
        json.NewEncoder(w).Encode(items)
    }))
    defer server.Close()

    client := api.NewClient(server.URL, "test-token")
    svc := NewMyFeatureService(client)

    items, err := svc.List(cmd.Context())

    require.NoError(t, err)
    assert.Len(t, items, 2)
    assert.Equal(t, "uuid-1", items[0].UUID)
}
```

### 8. Update Documentation

- Add command documentation to `README.md`
- Include usage examples and flag descriptions

## Testing Requirements

**All code changes MUST include tests.** This is non-negotiable.

### Coverage Requirements

- **Minimum coverage**: 70% for all packages
- **New features**: 80%+ coverage required
- **Bug fixes**: Must include regression tests
- **Refactoring**: Must maintain or improve existing coverage

### Running Tests

```bash
# Run all tests
go test ./internal/...

# Run with coverage
go test ./internal/... -cover

# Run specific package
go test ./internal/service/... -v

# Run specific test
go test ./internal/service -run TestServerService_List -v

# Generate coverage report
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Writing Tests

#### Use Table-Driven Tests

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "successful case",
            input:   "test",
            want:    "expected",
            wantErr: false,
        },
        {
            name:    "error case",
            input:   "",
            want:    "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

#### Mock HTTP Requests

**IMPORTANT**: Never call real APIs in tests. Use `httptest.NewServer()`:

```go
func TestServiceMethod(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request
        assert.Equal(t, "/api/v1/endpoint", r.URL.Path)
        assert.Equal(t, "GET", r.Method)

        // Return mock response
        response := models.MyResponse{Data: "test"}
        json.NewEncoder(w).Encode(response)
    }))
    defer server.Close()

    client := api.NewClient(server.URL, "test-token")
    // ... test your service
}
```

### Test Guidelines

- **Test naming**: `TestFunctionName_Scenario_ExpectedBehavior`
- **Use subtests**: `t.Run()` for related test cases
- **Use testify**: `require.NoError()` for must-pass assertions, `assert.Equal()` for comparisons
- **Mock HTTP**: Use `httptest.NewServer()` for all API tests
- **Test contexts**: Always pass `context.Background()` in tests
- **Test errors**: Verify error messages and types

## Code Style & Conventions

### Go Standards

- Follow standard Go idioms and conventions
- Use `gofmt` for code formatting
- Run `go vet` to catch common issues
- Prefer standard library over external dependencies

### Project Conventions

#### API Client Usage

```go
// Create client (usually done via cli.GetAPIClient())
client := api.NewClient(baseURL, token, api.WithDebug(true))

// GET request
var result MyStruct
err := client.Get(ctx, "endpoint", &result)

// POST request
err := client.Post(ctx, "endpoint", requestBody, &result)

// DELETE request
err := client.Delete(ctx, "endpoint")

// PATCH request
err := client.Patch(ctx, "endpoint", requestBody, &result)
```

#### Service Layer Pattern

```go
type MyService struct {
    client *api.Client
}

func NewMyService(client *api.Client) *MyService {
    return &MyService{client: client}
}

func (s *MyService) List(ctx context.Context) ([]models.Item, error) {
    var items []models.Item
    err := s.client.Get(ctx, "items", &items)
    return items, err
}
```

#### Error Handling

```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to fetch data: %w", err)
}

// Check and handle specific error types
if apiErr, ok := err.(*api.Error); ok {
    if apiErr.StatusCode == 404 {
        return fmt.Errorf("resource not found")
    }
}
```

#### Global Flags

All commands automatically inherit these global flags:

- `--format` (table|json|pretty) - Output format
- `--show-sensitive` - Show sensitive information
- `--debug` - Enable debug mode
- `--context` - Use specific context by name
- `--token` - Override context token

Access flags in commands:

```go
format, _ := cmd.Flags().GetString("format")
showSensitive, _ := cmd.Flags().GetBool("show-sensitive")
debug, _ := cmd.Flags().GetBool("debug")
```

## Submitting Changes

### Before Committing

```bash
# 1. Format code
go fmt ./...

# 2. Run tests
go test ./internal/...

# 3. Check coverage
go test ./internal/... -cover

# 4. Run vet
go vet ./...
```

### Commit Messages

Write clear, descriptive commit messages following conventional commits format:

```
<type>: <short summary>

<detailed description>

<footer>
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

Example:

```
feat: add server domains list command

- Implement GET /servers/{uuid}/domains endpoint
- Add server domains subcommand
- Include tests for domain listing
- Update README with new command documentation
```

### Pull Requests

1. **Fork** the repository
2. **Create a branch** from `v4.x`: `git checkout -b feature/my-feature v4.x`
3. **Make your changes** with tests
4. **Push** to your fork: `git push origin feature/my-feature`
5. **Open a pull request** against the `v4.x` branch
6. **Describe your changes** clearly in the PR description
7. **Link related issues** using "Fixes #123" or "Closes #123"

### PR Checklist

- [ ] Tests pass locally (`go test ./internal/...`)
- [ ] Code coverage meets requirements (70%+ minimum)
- [ ] Code is formatted (`go fmt ./...`)
- [ ] README.md updated (if adding new commands)
- [ ] CLAUDE.md updated (if changing architecture)
- [ ] Commit messages are descriptive
- [ ] PR description explains the changes
- [ ] All global flags are supported (format, show-sensitive, debug)
- [ ] Used UUIDs (not IDs) for resource identifiers

## Release Process (not for contributors :) ) 

Releases are automated using GoReleaser:

1. Tag a new version: `git tag v1.2.3`
2. Push the tag: `git push origin v1.2.3`
3. Create a GitHub release
4. GoReleaser builds binaries for all platforms automatically

## Getting Help

- **Discord**: https://coolify.io/discord
- **Issues**: [Open an issue](https://github.com/coollabsio/coolify-cli/issues) for bugs or feature requests
- **Architecture**: Read [ARCHITECTURE.md](ARCHITECTURE.md) for detailed design documentation
- **API Reference**: See the [OpenAPI specification](https://github.com/coollabsio/coolify/blob/v4.x/openapi.json)
- **Code Guidance**: See [CLAUDE.md](CLAUDE.md) for AI assistant guidance

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to Coolify CLI! 🚀
