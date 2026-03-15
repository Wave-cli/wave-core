# wave-core

A modular CLI orchestrator powered by plugins. Plugins are standalone binaries installed to `~/.wave/plugins/` and executed through the unified `wave` command.

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
- **Schema**: Optional Waveschema file for config validation

Plugins are installed to `~/.wave/plugins/<org>/<name>/<version>/` with a `current` symlink pointing to the active version.

## Schema validation

Plugins can ship a Waveschema file that defines the expected structure of their Wavefile section. wave-core validates config before passing it to the plugin:

- **Schema validation**: Checks required fields, types, and unknown keys
- **Rules engine**: Rejects structural errors like nested headers (`[flow.build]`) and leaked keys

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
