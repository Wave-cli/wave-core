# wave-core Agent Guidelines

This document provides conventions and workflows for agents contributing to wave-core.

## Project Overview

wave-core is a Go-based modular CLI orchestrator powered by plugins. Plugins are standalone binaries installed to `~/.wave/plugins/` and executed through the unified `wave` command.

- **Language**: Go 1.25.0
- **Module**: `github.com/wave-cli/wave-core`
- **CLI framework**: Cobra + Viper
- **Config format**: TOML (via BurntSushi/toml)

## Build / Lint / Test Commands

### Via `just` (preferred)

```bash
# Run all tests
just test

# Run tests with verbose output
just test-v

# Run tests with coverage report
just test-cover

# Run unit tests only (skip e2e)
just test-unit

# Run e2e tests only
just test-e2e

# Run tests for a specific package
just test-pkg config

# Format + vet
just lint

# Full CI pipeline
just ci

# Build binary
just build

# Run wave with args
just run version
```

### Via `go` directly

```bash
# Run all tests
go test ./...

# Run all tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run a specific test by name
go test ./internal/errors/ -run TestParseStderrStructured -v

# Run a specific package
go test -v ./internal/config/

# Run e2e tests
go test -v ./e2e/...

# Format
go fmt ./...

# Vet
go vet ./...

# Build
go build -o bin/wave .
```

## Code Style

### Formatting

- Use `go fmt ./...` before every commit. No exceptions.
- Use `go vet ./...` in the lint step.
- No line length limit enforced, but keep lines reasonably short.
- Use a linter/formatter that auto-formats on save if available.

### Imports

Group imports in three blocks separated by blank lines:

```go
import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"

    "github.com/wave-cli/wave-core/internal/bootstrap"
    "github.com/wave-cli/wave-core/internal/config"
)
```

1. Standard library
2. External packages (from `go.mod`)
3. Internal packages (from this module)

### Naming

- **Packages**: lowercase, single-word or short compound (e.g., `executor`, `pluginmgmt`, `downloader`)
- **Types / Functions / Variables**: PascalCase for exported, camelCase for unexported
- **Interfaces**: PascalCase with "er" suffix for single-method interfaces (e.g., `Reader`, `Writer`)
- **Constants**: PascalCase for exported, camelCase for unexported
- **Acronyms**: preserve original casing (e.g., `HTTP`, `URL`, `JSON`, `API`)
- **Error variables**: `var ErrXXX = errors.New("...")` for sentinel errors, or `ErrXXX` suffix on error wrapper
- **Test helpers**: `buildEchoPlugin(t *testing.T) string`, `setupFakeWaveHome(t *testing.T, binPath string) (homeDir, pluginsDir string)`

### Package Documentation

Every package should have a package-level doc comment:

```go
// Package executor handles plugin execution: fork/exec with stdin config,
// environment variables, and stderr parsing.
package executor
```

Every exported function should have a doc comment:

```go
// Execute runs a plugin binary with the given arguments, config section
// (as stdin JSON), and environment variables.
func Execute(...) (*Result, error) { ... }
```

### Error Handling

- Wrap errors with `fmt.Errorf("context: %w", err)` for Go-level errors.
- Return structured `*PluginError` for plugin-level errors (defined in `internal/errors/protocol.go`).
- Never silently ignore errors with `_`: always handle or log.
- In CLI commands, prefer `RunE` functions that return errors.
- Exit non-zero via `os.Exit(code)` for CLI-level failures after error reporting.

### Return Values

- Functions that can fail return `(value, error)`.
- Constructors that can fail return `(Type, error)`.
- Do not use named return values except for simple, obvious cases.

### Context

- For long-running operations that may need cancellation, prefer `context.Context`.
- For plugin execution, use `context.WithTimeout` if a timeout is needed.

### Goroutines

- Every goroutine must have a clear lifetime tied to a `context.Context` or explicit channel.
- Use `sync.WaitGroup` when launching multiple goroutines that must all complete.

### Testing Conventions

- Test files are named `*_test.go` and live in the same package as the code under test.
- Use `t.TempDir()` for all temporary test directories (automatically cleaned up).
- Use `t.Helper()` for test helper functions.
- Use table-driven tests when testing multiple cases:

```go
func TestParseStderrNotWaveError(t *testing.T) {
    cases := []string{
        `{"wave_error":false,"code":"X","message":"Y"}`,
        `{"code":"X","message":"Y"}`,
    }
    for _, c := range cases {
        pe := ParseStderr([]byte(c))
        if pe != nil {
            t.Errorf("Should return nil for non-wave JSON: %s", c)
        }
    }
}
```

- Test function names: `Test<Package>_<Method>_<Scenario>` (e.g., `TestParseStderrStructured`, `TestE2E_BootstrapCreatesDirectoryStructure`).
- E2E tests are in `e2e/e2e_test.go` and exercise the full pipeline.
- Test the echo plugin at `testdata/plugins/echo/main.go` for protocol testing.
- Prefer `t.Fatalf` / `t.Fatalf` for setup failures, `t.Errorf` for assertion failures.
- Mock HTTP with `net/http/httptest` for downloader tests.

## Project Structure

```
cmd/               # Cobra command implementations
internal/
  bootstrap/      # First-run wave home setup
  config/         # Wavefile and global config parsing
  downloader/     # GitHub release fetching and install
  errors/         # Plugin error protocol and logging
  executor/       # Plugin fork/exec runner
  pluginmgmt/     # Plugin registry, resolution, Waveplugin parsing
  ui/             # Printer with verbosity levels
  version/        # Build-time version injection
pkg/sdk/          # Plugin SDK (for plugin authors)
e2e/              # End-to-end tests
testdata/plugins/ # Test plugin binaries (echo)
```

## Configuration

- **Global config**: `~/.wave/config` (TOML, managed by Viper)
- **Wavefile**: `Wavefile` at project root (TOML, parsed by BurntSushi/toml)
- **Plugin directory**: `~/.wave/plugins/<name>/` (single version, no org folder)
- **Plugin logs**: `~/.wave/logs/<date>.log` (JSONL)

## Dependency Management

```bash
# Tidy dependencies
just tidy

# Download dependencies
just deps
```

Always run `go mod tidy` before committing changes to `go.mod` or `go.sum`.

## CI Pipeline

The full CI pipeline runs: `fmt` -> `vet` -> `test` -> `build`. Run `just ci` locally before pushing.
