# Contributing to Coolify CLI

Thank you for your interest in contributing to the Coolify CLI! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing Requirements](#testing-requirements)
- [Submitting Changes](#submitting-changes)
- [Code Style](#code-style)
- [Project Architecture](#project-architecture)

## Getting Started

Before you start contributing:

1. **Fork the repository** on GitHub
2. **Clone your fork** locally
3. **Create a new branch** for your feature or bug fix
4. **Read the CLAUDE.md** file for detailed architectural guidance

## Development Setup

### Prerequisites

- Go 1.22 or higher
- Git

### Building the Project

```bash
# Clone the repository
git clone https://github.com/coollabsio/coolify-cli.git
cd coolify-cli

# Build the CLI
go build -o coolify .

# Install locally
go install
```

### Running the CLI

```bash
# Run without installing
go run main.go [command]

# Example commands
go run main.go instances list
go run main.go servers list --debug
```

## Making Changes

### Before You Code

1. **Check existing issues** to see if your feature/bug is already being worked on
2. **Open an issue** to discuss your proposed changes (for large features)
3. **Review the API specification** at https://github.com/coollabsio/coolify/blob/v4.x/openapi.json

### Adding a New Command

When adding a new command, follow this checklist:

- [ ] Create command implementation in `cmd/`
- [ ] Implement all three output formats: `table`, `json`, `pretty`
- [ ] Call `CheckDefaultThings(nil)` for version validation
- [ ] Use `Fetch()`, `Post()`, or `Delete()` helpers for API calls
- [ ] Create corresponding test file(s)
- [ ] Test all flags, arguments, and error cases
- [ ] Add integration tests if needed
- [ ] Update README.md with command documentation
- [ ] Verify test coverage meets requirements

## Testing Requirements

**All code changes MUST include tests.** This is non-negotiable.

### Coverage Requirements

- **Minimum coverage**: 70% for all packages
- **New features**: 80%+ coverage required
- **Bug fixes**: Must include regression tests

### Running Tests

```bash
# Run all tests
go test ./internal/...

# Run with coverage
go test ./internal/... -cover

# Generate coverage report
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run with verbose output
go test ./internal/... -v
```

### Writing Tests

Use table-driven tests for multiple scenarios:

```go
func TestYourFunction(t *testing.T) {
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
            got, err := YourFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("YourFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("YourFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Test Guidelines

- **Never call real APIs** in unit tests - use `httptest.NewServer()` for mocks
- Use descriptive test names: `TestFunctionName_Scenario_ExpectedBehavior`
- Use `t.Run()` for subtests
- Use `t.Parallel()` when tests are independent
- Store mock API responses in `test/fixtures/`

## Submitting Changes

### Before Committing

```bash
# 1. Run tests
go test ./internal/...

# 2. Check coverage
go test ./internal/... -cover

# 3. Format code
go fmt ./...

# 4. Run linter (if available)
golangci-lint run
```

### Commit Messages

Write clear, descriptive commit messages:

```
Add servers delete command

- Implement DELETE /servers/{uuid} endpoint
- Add confirmation prompt with --force flag
- Include tests for success and error cases
- Update README with new command documentation
```

### Pull Requests

1. **Push your branch** to your fork
2. **Open a pull request** against the `v4.x` branch
3. **Describe your changes** clearly in the PR description
4. **Link related issues** using "Fixes #123" or "Closes #123"
5. **Ensure CI passes** - tests must pass and coverage must meet requirements

### PR Checklist

- [ ] Tests pass locally (`go test ./internal/...`)
- [ ] Code coverage meets requirements (70%+ minimum)
- [ ] Code is formatted (`go fmt ./...`)
- [ ] README.md updated (if adding new commands)
- [ ] CLAUDE.md updated (if changing architecture)
- [ ] Commit messages are descriptive
- [ ] PR description explains the changes

## Code Style

### Go Standards

- Follow standard Go idioms and conventions
- Use Go 1.22+ features
- Prefer standard library over external dependencies
- Use `net/http` for HTTP operations
- Implement proper error handling

### Project Conventions

- Use Cobra command pattern for CLI commands
- Use Viper for configuration management
- Support all three output formats: `table`, `json`, `pretty`
- Handle sensitive data with `--show-sensitive` flag
- Follow RESTful patterns for API interactions

## Project Architecture

### Command Structure

```
cmd/
├── root.go          # Root command with core utilities
├── context.go       # Manage Coolify instances
├── servers.go       # Server management
├── projects.go      # Project management
├── resources.go     # Resource management
├── deploy.go        # Deployment operations
├── domains.go       # Domain management
├── privatekeys.go   # SSH key management
├── update.go        # Self-update functionality
└── version.go       # Version information
```

### Key Patterns

- Use `Fetch(url)` for GET requests
- Use `Post(url, input)` for POST requests
- Use `Delete(url)` for DELETE requests
- Call `CheckDefaultThings(nil)` for version validation
- All API calls use `/api/v1/` base path

## Getting Help

- **Issues**: Open an issue on GitHub for bugs or feature requests
- **Documentation**: Read CLAUDE.md for detailed architectural guidance
- **API Spec**: Refer to the OpenAPI schema at https://github.com/coollabsio/coolify/blob/v4.x/openapi.json

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to Coolify CLI! 🚀
