# wave-core Architecture

## Overview

wave-core is a modular CLI orchestrator that discovers, resolves, and executes plugins. It reads configuration from two layers (global `~/.wave/config` and project-level `Wavefile`) and provides a structured error protocol.

**Key principle:** Plugins are never exposed as standalone binaries on `$PATH`. The only way to run a plugin is `wave <plugin> [args...]`.

## Directory Structure

```
wave-core/
├── cmd/                        # Cobra CLI commands (one file per command)
│   ├── root.go                 # Root command + Viper flag binding + plugin dispatch
│   ├── version.go              # wave version
│   ├── init.go                 # wave init (scaffold Wavefile)
│   ├── install.go              # wave install <org/plugin>[@version]
│   ├── uninstall.go            # wave uninstall <plugin>
│   └── list.go                 # wave list
├── internal/                   # Internal packages (not importable externally)
│   ├── version/                # Build-time version info + semver comparison
│   ├── ui/                     # Printer with verbosity levels
│   ├── config/                 # Global config + Wavefile parsing
│   ├── bootstrap/              # First-run directory/config setup
│   ├── errors/                 # Error protocol, parsing, formatting, logging
│   ├── pluginmgmt/             # Plugin metadata, name resolution, registry
│   ├── downloader/             # GitHub Releases client + tar.gz extraction
│   └── runner/                 # Plugin executor (fork/exec, stdin, env vars)
├── pkg/sdk/                    # Go SDK for plugin authors
├── e2e/                        # End-to-end integration tests
├── testdata/plugins/echo/      # Test echo plugin
├── main.go                     # Entry point
├── plan/                        # Planning docs
│   ├── plan.md                  # Full architecture plan
│   └── tasks.md                 # Task tracker
```

## Module Dependency Graph

```
main.go
  └── cmd/root.go
        ├── internal/bootstrap   (first-run setup)
        ├── internal/config      (parse global config + Wavefile)
        ├── internal/pluginmgmt  (resolve plugin paths)
        ├── internal/executor    (execute plugin binaries)
        ├── internal/errors      (parse/format/log errors)
        ├── internal/ui          (terminal output)
        ├── internal/version     (version info)
        └── internal/downloader  (GitHub Releases client)
```

## Configuration System

### Global Config (`~/.wave/config`)

TOML file auto-created on first run:

```toml
[core]
logs_dir = "~/.wave/logs"

[projects]
folders = ["~/projects"]

[plugins]
"wave-cli/flow" = "1.2.0"
```

### Project Config (`Wavefile`)

Per-project TOML file discovered by walking up from `cwd`:

```toml
[project]
name = "my-app"
version = "1.0.0"
owner = "bouajila"

[flow]
environment = "staging"
port = 3000
```

## Plugin Execution Flow

```
wave flow dev
  1. Root command sees "flow" is not built-in
  2. Looks up "flow" in installed plugins -> "wave-cli/flow" v1.2.0
  3. Resolves binary: ~/.wave/plugins/flow/bin/flow
  4. Discovers Wavefile, extracts [flow] section
  5. Serializes section as JSON, passes via stdin
  6. Sets WAVE_* env vars (name, version, dir, assets, project root)
  7. Fork/exec the plugin binary
  8. Streams stdout/stderr in real-time
  9. Parses stderr for structured wave error JSON
  10. Logs errors to daily JSONL file
```

## Error Protocol

Plugins emit structured errors as JSON on stderr:

```json
{"wave_error":true,"code":"FLOW_ENV_MISSING","message":"Environment not configured","details":"Check Wavefile"}
```

wave-core detects `wave_error: true`, formats for display, and logs to `~/.wave/logs/<date>.log`.

## Testing

All modules follow TDD (test-first). Run the full suite:

```bash
# Unit tests (all packages)
go test ./...

# With coverage
go test ./... -cover

# Verbose
go test ./... -v

# E2E tests only
go test ./e2e/ -v

# Specific package
go test ./internal/config/ -v
```

### Test Coverage Summary

| Package | Coverage | Tests |
|---------|----------|-------|
| `internal/version` | 93% | 10 |
| `internal/ui` | 100% | 12 |
| `internal/config` | 93% | 12 |
| `internal/bootstrap` | 74% | 7 |
| `internal/errors` | 91% | 12 |
| `internal/pluginmgmt` | 90% | 15 |
| `internal/downloader` | 78% | 13 |
| `internal/executor` | 50% | 9 |
| `pkg/sdk` | 65% | 7 |
| **e2e** | - | 15 |
| **Total** | | **112** |

See [testing.md](testing.md) for detailed testing guide.
