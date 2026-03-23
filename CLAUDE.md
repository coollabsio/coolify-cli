# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a CLI tool for interacting with the Coolify API, built with Go using the Cobra framework. The CLI allows users to manage Coolify instances (both cloud and self-hosted), servers, projects, resources, deployments, domains, and private keys.

### API Specification
This CLI is a client for the Coolify API. The API specification is defined in the OpenAPI schema:
- **Source**: https://github.com/coollabsio/coolify/blob/v4.x/openapi.json
- **Raw JSON**: https://raw.githubusercontent.com/coollabsio/coolify/refs/heads/v4.x/openapi.json
- **Base Path**: `/api/v1/`
- **Authentication**: Bearer token (API tokens from Coolify dashboard at `/security/api-tokens`)

All commands in this CLI are wrappers around API endpoints defined in the OpenAPI specification. When adding new features or endpoints:
1. Check the OpenAPI spec for available endpoints and their request/response schemas
2. Ensure the CLI command structure follows the API resource hierarchy
3. Match the API's data types and validation rules

## Architecture

### Command Structure
The codebase follows Cobra's command pattern with a root command and subcommands:
- Entry point: `coolify/main.go` calls `cmd.Execute()`
- Root command: `cmd/root.go` - contains core utilities (HTTP client, authentication, version checking, config management)
- Subcommands: Each command is in its own file in `cmd/`:
  - `context.go` - manage Coolify context (add, remove, list, set default/token)
  - `servers.go` - list and get server information
  - `projects.go` - list projects with environments and applications
  - `resources.go` - list resources
  - `deploy.go` - deploy resources
  - `domains.go` - manage domains
  - `privatekeys.go` - manage SSH keys
  - `update.go` - self-update CLI
  - `version.go` - show CLI version

### Configuration Management
- Uses Viper for configuration management
- Config file location: `~/.config/coolify/config.json` (via xdg package)
- Config stores multiple instances with tokens, default instance selection
- Global flags available: `--token`, `--host`, `--format`, `--show-sensitive`, `--force`, `--debug`

### API Communication
Core API functions in `cmd/root.go`:
- `Fetch(url string)` - GET requests
- `Post(url, input)` - POST requests
- `Delete(url)` - DELETE requests
All API calls use `Fqdn + "/api/v1/" + url` pattern with Bearer token authentication

### Version Management
- CLI version tracking with auto-update check (10 minute interval)
- API version checking and minimum version enforcement via `CheckMinimumVersion()`
- Self-update capability using `go-selfupdate` library

### Output Formatting
Three output modes supported via `--format` flag:
- `table` (default) - tabwriter formatted output
- `json` - compact JSON
- `pretty` - indented JSON

## Development Commands

### Build
```bash
go build -o coolify ./coolify
```

### Run locally
```bash
go run ./coolify [command]
```

### Test a command
```bash
go run ./coolify context list
go run ./coolify servers list --debug
```

### Install locally
```bash
go install ./coolify
```

### Run tests
```bash
# Run all tests (tests are in internal/ directory)
go test ./internal/...

# Run with coverage
go test ./internal/... -cover

# Run with verbose output
go test ./internal/... -v

# Run specific package
go test ./internal/api/... -v
go test ./internal/service/... -v

# Run specific test
go test ./internal/api -run TestClient_Get_Success -v
```

### Before committing
```bash
# 1. Run tests
go test ./internal/...

# 2. Check coverage
go test ./internal/... -cover

# 3. Run linter (if available)
golangci-lint run

# 4. Format code
go fmt ./...
```

## Release Process

- Uses GoReleaser for multi-platform builds (Linux, Darwin, Windows on amd64/arm64)
- Release workflow: `.github/workflows/release-cli.yml` triggers on GitHub releases
- GoReleaser config: `.goreleaser.yml`
- Install script: `scripts/install.sh` downloads from GitHub releases

## Key Patterns

### Adding a New Command
1. Create new file in `cmd/` (e.g., `cmd/newfeature.go`)
2. Define command struct with cobra.Command
3. Implement Run function with:
   - Call `CheckDefaultThings(nil)` to validate version and format
   - Use `Fetch()`, `Post()`, or `Delete()` helpers
   - Handle JSON unmarshaling into typed structs
   - Support all three output formats
4. Register command in `init()` function: `rootCmd.AddCommand(yourCmd)`

### API Version Requirements
If a command requires a specific Coolify API version, pass it to `CheckDefaultThings()`:
```go
minimumVersion := "4.0.0"
CheckDefaultThings(&minimumVersion)
```

### Handling Sensitive Data
- Use `ShowSensitive` flag to control display of tokens/secrets
- Default overlay: `SensitiveInformationOverlay = "********"`

### UUID vs ID Pattern
**CRITICAL: Always use UUIDs for user-facing interactions, never internal database IDs.**

When adding new commands or models:
1. **Command Arguments**: Always accept UUIDs as string arguments (e.g., `<resource_uuid>`), never integer IDs
2. **API Endpoints**: Construct API paths using UUIDs (e.g., `resources/{uuid}`), not IDs
3. **Service Layer**: Methods should accept `uuid string` parameters, not `id int`
4. **Table Output**: Hide internal IDs from table output using `table:"-"` struct tags
5. **Model Fields**:
   - Keep `ID int` field with `json:"id" table:"-"` (for API responses, hidden from users)
   - Always include `UUID string` field with `json:"uuid"` (visible to users)

**Example model:**
```go
type Resource struct {
    ID   int    `json:"id" table:"-"`     // Hidden from table output
    UUID string `json:"uuid"`              // Shown in table output
    Name string `json:"name"`
    // ... other fields
}
```

**Why UUIDs?**
- UUIDs are stable across environments (dev, staging, prod)
- IDs are internal implementation details that can change
- UUIDs are more secure (don't expose database sequencing)
- Coolify API uses UUIDs as the primary resource identifier

## Testing Requirements

**CRITICAL: All code changes MUST include tests. This is non-negotiable.**

### Test Coverage Requirements
- **Minimum coverage**: 70% for all packages
- **New features**: Must have 80%+ coverage
- **Bug fixes**: Must include regression tests
- **Refactoring**: Must maintain or improve existing coverage

### Testing Structure
```
test/
├── fixtures/           # Test data, mock API responses
├── mocks/             # Mock implementations of interfaces
└── integration/       # Integration tests with test server
```

### Test Requirements by Package Type

#### 1. Command Tests (`cmd/*_test.go`)
- Test command parsing and flag handling
- Test output formatting (table, json, pretty)
- Use mock API client to avoid real API calls
- Test error handling and validation
- Example:
```go
func TestServersListCmd(t *testing.T) {
    // Test with mock client
    // Verify output format
    // Test error cases
}
```

#### 2. API Client Tests (`internal/api/*_test.go`)
- Test request building
- Test response parsing
- Test error handling (4xx, 5xx status codes)
- Test retry logic
- Test timeout behavior
- **IMPORTANT**: Use `httptest.NewServer()` for mock HTTP responses (NOT real APIs)
- All API tests must use local mock servers, never call real Coolify cloud or external APIs

#### 3. Service Tests (`internal/service/*_test.go`)
- Test business logic
- Mock API client
- Test complex workflows
- Test error propagation

#### 4. Model Tests (`internal/models/*_test.go`)
- Test JSON marshaling/unmarshaling
- Test validation logic
- Test helper methods

#### 5. Integration Tests (`test/integration/*_test.go`)
- Test full command execution
- Test with real HTTP server (httptest)
- Test config file operations
- Test version checking
- Can be run with `-short` flag to skip

### Running Tests

```bash
# Run all tests (tests are in internal/ directory)
go test ./internal/...

# Run with coverage
go test ./internal/... -cover

# Generate coverage report
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run with verbose output
go test ./internal/... -v

# Run only unit tests (skip integration)
go test ./internal/... -short

# Run specific package
go test ./internal/api/... -v
go test ./internal/service/... -v
```

### Test Guidelines

1. **Table-driven tests**: Use for testing multiple scenarios
2. **Test naming**: `TestFunctionName_Scenario_ExpectedBehavior`
3. **Subtests**: Use `t.Run()` for related test cases
4. **Setup/Teardown**: Use `TestMain()` for package-level setup
5. **Parallel tests**: Use `t.Parallel()` when tests are independent
6. **Mock dependencies**: Never call real APIs in unit tests
7. **Test fixtures**: Store mock API responses in `test/fixtures/`

### Example Test Structure

```go
func TestServersList(t *testing.T) {
    tests := []struct {
        name       string
        response   string
        wantErr    bool
        wantCount  int
    }{
        {
            name:      "successful list",
            response:  readFixture("servers_list.json"),
            wantErr:   false,
            wantCount: 3,
        },
        {
            name:      "empty list",
            response:  "[]",
            wantErr:   false,
            wantCount: 0,
        },
        {
            name:      "api error",
            response:  `{"error":"unauthorized"}`,
            wantErr:   true,
            wantCount: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### When Adding a New Command

**CHECKLIST** (must complete ALL items):
- [ ] Create command implementation in `cmd/`
- [ ] Create corresponding test file in `internal/service/*_test.go` or `internal/api/*_test.go`
- [ ] Test all flags and arguments
- [ ] Test all output formats (table, json, pretty)
- [ ] Test error cases (missing args, API errors, invalid input)
- [ ] Add integration test if command has complex workflow
- [ ] Update README.md with command documentation
- [ ] Run `go test ./internal/...` and ensure all tests pass
- [ ] Verify coverage: `go test ./internal/... -cover`

### CI/CD Integration

Tests run automatically on:
- Every pull request
- Every commit to main branch
- Before releases

**Pull requests will be blocked if:**
- Any test fails
- Coverage drops below 70%
- New code has no tests

## .cursorrules Context

The project follows Go 1.22+ idioms with standard library preference:
- Use `net/http` standard library (no external HTTP frameworks)
- Leverage Go 1.22 ServeMux features for any routing needs
- Follow RESTful patterns for API interactions
- Implement proper error handling with custom types when needed
- Use Go's concurrency features appropriately
- Write secure, efficient, and maintainable code
- **ALWAYS write tests** - see Testing Requirements section above