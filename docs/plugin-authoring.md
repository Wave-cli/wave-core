# Creating Wave Plugins

This guide explains how to create a plugin for the wave ecosystem.

## Overview

A wave plugin is a standalone binary that:
1. Reads configuration from **stdin** (JSON object)
2. Reads environment variables set by wave-core
3. Writes output to **stdout**
4. Reports errors as structured JSON on **stderr**

Plugins can be written in **any language** - the protocol is language-agnostic.

## Quick Start (Go)

### 1. Create Plugin Binary

```go
package main

import (
    "fmt"
    "os"

    "github.com/wave-cli/wave-core/pkg/sdk"
)

func main() {
    // Read config from stdin (passed from Wavefile)
    cfg, err := sdk.ReadConfig()
    if err != nil {
        sdk.Err("CONFIG_ERROR", "failed to read config: "+err.Error())
    }

    // Read wave environment
    env := sdk.GetPluginEnv()
    fmt.Printf("Running %s v%s\n", env.Name, env.Version)
    fmt.Printf("Project: %s\n", env.ProjectRoot)

    // Access config values
    if port, ok := cfg["port"]; ok {
        fmt.Printf("Port: %v\n", port)
    }

    // Handle subcommands
    args := os.Args[1:]
    if len(args) == 0 {
        sdk.Err("NO_COMMAND", "no subcommand provided")
    }

    switch args[0] {
    case "dev":
        fmt.Println("Starting dev server...")
    case "build":
        fmt.Println("Building project...")
    default:
        sdk.Err("UNKNOWN_CMD", "unknown command: "+args[0])
    }
}
```

### 2. Create Waveplugin Metadata

Every release must include a `Waveplugin` file:

```toml
[plugin]
name = "myflow"
version = "1.0.0"
description = "Development workflow automation"
creator = "my-org"
license = "MIT"
homepage = "https://github.com/my-org/myflow"

[compatibility]
min_wave_version = "0.1.0"

[assets]
files = ["templates/", "defaults.toml"]
```

### 3. Create GitHub Release

Build binaries for each platform and create a GitHub release with these assets:

```
myflow-linux-amd64.tar.gz    # Linux binary (tar.gz)
myflow-darwin-arm64.tar.gz   # macOS ARM binary (tar.gz)
myflow-darwin-amd64.tar.gz   # macOS Intel binary (tar.gz)
myflow-windows-amd64.zip     # Windows binary (zip)
Waveplugin                   # Metadata file
```

### 4. Users Install Your Plugin

```bash
wave install my-org/myflow
wave install my-org/myflow@1.0.0  # specific version
```

## Plugin Protocol

### Stdin (Configuration)

wave-core passes the plugin's config section from the Wavefile as a JSON object on stdin:

```json
{"environment": "staging", "port": 3000, "watch": true}
```

This comes from the `[myflow]` section in the project's Wavefile:

```toml
[myflow]
environment = "staging"
port = 3000
watch = true
```

### Environment Variables

wave-core sets these environment variables before executing your plugin:

| Variable | Description | Example |
|----------|-------------|---------|
| `WAVE_PLUGIN_NAME` | Plugin short name | `myflow` |
| `WAVE_PLUGIN_VERSION` | Installed version | `1.0.0` |
| `WAVE_PLUGIN_DIR` | Plugin installation directory | `~/.wave/plugins/my-org/myflow/current` |
| `WAVE_PLUGIN_ASSETS` | Plugin assets directory | `~/.wave/plugins/my-org/myflow/current/assets` |
| `WAVE_PROJECT_ROOT` | Project root (where Wavefile is) | `/home/user/my-project` |

### Arguments

Plugin arguments come from the wave command line:

```bash
wave myflow dev --port 8080
# Plugin receives: args = ["dev", "--port", "8080"]
```

### Error Protocol

Report errors by writing JSON to **stderr**:

```json
{"wave_error": true, "code": "MY_ERROR_CODE", "message": "Human readable message", "details": "Optional extra info"}
```

**Required fields:** `wave_error` (must be `true`), `code`, `message`
**Optional fields:** `details`

Using the Go SDK:

```go
// Simple error (exits with code 1)
sdk.Err("CONFIG_MISSING", "configuration file not found")

// Error with details
sdk.ErrWithDetails("BUILD_FAILED", "build step failed", "check build logs at /tmp/build.log")
```

wave-core will:
1. Detect the structured error in stderr
2. Display it in a formatted way to the user
3. Log it to `~/.wave/logs/<date>.log` as JSONL

## Testing Your Plugin

### Unit Testing

Test your plugin logic independently:

```go
func TestDevCommand(t *testing.T) {
    cfg := map[string]any{"port": float64(3000)}
    // test your business logic...
}
```

### Integration Testing with wave-core

Build your plugin and test it with wave-core's runner:

```go
import "github.com/wave-cli/wave-core/internal/runner"

func TestPluginExecution(t *testing.T) {
    result, err := runner.Execute(
        "/path/to/myflow",
        []string{"dev"},
        map[string]any{"port": int64(3000)},
        "myflow", "1.0.0", "/project/root",
    )
    if err != nil {
        t.Fatal(err)
    }
    // assert on result.Stdout, result.ExitCode, etc.
}
```

### Test Echo Plugin Reference

See `testdata/plugins/echo/main.go` for a minimal reference implementation that demonstrates all protocol features.

## Non-Go Plugins

Any language can implement a wave plugin. The protocol is:

1. **Read stdin**: Parse JSON from stdin for config
2. **Read env vars**: `WAVE_PLUGIN_NAME`, `WAVE_PLUGIN_VERSION`, etc.
3. **Write stdout**: Normal output
4. **Write stderr**: JSON error protocol on failure

### Python Example

```python
import json, sys, os

# Read config
config = json.load(sys.stdin)

# Read env
plugin_name = os.environ.get("WAVE_PLUGIN_NAME", "unknown")
project_root = os.environ.get("WAVE_PROJECT_ROOT", ".")

# Handle commands
command = sys.argv[1] if len(sys.argv) > 1 else None

if command == "fail":
    error = {"wave_error": True, "code": "PY_ERROR", "message": "something went wrong"}
    print(json.dumps(error), file=sys.stderr)
    sys.exit(1)

print(f"OK {plugin_name} {command}")
```

### Node.js Example

```javascript
const config = JSON.parse(require('fs').readFileSync('/dev/stdin', 'utf8'));
const name = process.env.WAVE_PLUGIN_NAME;
const command = process.argv[2];

if (command === 'fail') {
    console.error(JSON.stringify({wave_error: true, code: "JS_ERROR", message: "failed"}));
    process.exit(1);
}

console.log(`OK ${name} ${command}`);
```

## Release Checklist

- [ ] Binary builds for linux/amd64, darwin/arm64, darwin/amd64
- [ ] Each binary packaged as `<name>-<os>-<arch>.tar.gz`
- [ ] `Waveplugin` metadata file included in release
- [ ] `min_wave_version` set correctly in `[compatibility]`
- [ ] All assets declared in `[assets].files` are included
- [ ] GitHub Release created with all assets attached
