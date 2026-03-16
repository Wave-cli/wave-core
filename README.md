# wave-core

[![GitHub stars](https://img.shields.io/github/stars/Wave-cli/wave-core?style=flat&logo=github)](https://github.com/Wave-cli/wave-core/stargazers)
[![Issues](https://img.shields.io/github/issues/Wave-cli/wave-core?style=flat&logo=github)](https://github.com/Wave-cli/wave-core/issues)
[![License: MIT](https://img.shields.io/badge/license-MIT-brightgreen?style=flat)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.25.0-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/Wave-cli/wave-core?style=flat&logo=github)](https://github.com/Wave-cli/wave-core/releases)

A modular CLI orchestrator powered by plugins. Plugins are standalone binaries installed to `~/.wave/plugins/` and executed through the unified `wave` command.

## Table of contents

- [Install](#install)
- [Quick start](#quick-start)
- [Wavefile](#wavefile)
- [Built-in commands](#built-in-commands)
- [Plugin architecture](#plugin-architecture)
- [Local flow plugin testing](#local-flow-plugin-testing)
- [Development](#development)

## Install

```bash
go install github.com/wave-cli/wave-core@latest
```

## Quick start

```bash
# Initialize a project
wave init

# Install a plugin
wave install wave-cli/flow

# Use it
wave flow build
```

## Wavefile

Projects are configured through a `Wavefile` at the project root:

```toml
[project]
name = "my-app"
version = "1.0.0"
owner = "bouajila"

[flow]
build = { cmd = "go build -o bin/app", on_success = "echo done" }
clean = { cmd = "rm -rf bin/" }
dev   = { cmd = "go run ." }
```

Each section after `[project]` maps to an installed plugin. wave-core passes the section as JSON on stdin when executing the plugin.

## Built-in commands

| Command | Description |
|---------|-------------|
| `wave init` | Scaffold a Wavefile |
| `wave install <org/plugin>` | Install a plugin from GitHub Releases |
| `wave uninstall <plugin>` | Remove an installed plugin |
| `wave list` | List installed plugins |
| `wave config` | Show global configuration |
| `wave version` | Print version info |

## Plugin architecture

Plugins are standalone binaries that follow a simple protocol:

- **Config**: JSON on stdin (from the Wavefile section)
- **Environment**: `WAVE_PLUGIN_NAME`, `WAVE_PLUGIN_VERSION`, `WAVE_PLUGIN_DIR`, `WAVE_PLUGIN_ASSETS`, `WAVE_PROJECT_ROOT`
- **Errors**: Structured JSON on stderr (`{"wave_error": true, "code": "...", "message": "..."}`)

Plugins are installed to `~/.wave/plugins/<org>/<name>/<version>/` with a `current` symlink pointing to the active version.

## Local flow plugin testing

Use this when you are developing `wave-flow` locally and want to run it through `wave` without publishing a release:

```bash
# Build the plugin (from the sibling repo)
cd ../wave-flow
go build -o bin/flow .

# Install into local wave plugin dir
mkdir -p ~/.wave/plugins/wave-cli/flow/v0.1.0/bin
cp bin/flow ~/.wave/plugins/wave-cli/flow/v0.1.0/bin/flow
cp Waveplugin ~/.wave/plugins/wave-cli/flow/v0.1.0/Waveplugin
ln -sfn ~/.wave/plugins/wave-cli/flow/v0.1.0 ~/.wave/plugins/wave-cli/flow/current
```

Then add a `[flow]` section in your Wavefile and run:

```bash
wave flow --list
wave flow build
```

## Using flow in this repo

This repository includes a `Wavefile` that mirrors common `just` recipes. If you have the flow plugin installed, you can run the usual dev tasks via:

```bash
wave flow test
wave flow test-e2e
wave flow lint
wave flow build
wave flow ci
```

## Development

```bash
# Run all tests
go test ./...

# With coverage
go test ./... -cover

# Build
go build -o bin/wave .
```

See [docs/architecture.md](docs/architecture.md) for detailed architecture documentation.

## License

MIT
