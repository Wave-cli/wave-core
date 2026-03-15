# wave-core — Architecture & Implementation Plan

## Overview

wave-core is the central orchestrator of the wave ecosystem. It does **three things well**:
1. Reads configuration from two layers (project-level `Wavefile`, global `~/.wave/config`)
2. Discovers, resolves, and executes installed plugins — exclusively via `wave <plugin> [args...]`
3. Provides a structured error protocol that plugins must follow, with user-friendly output and file-based logging

Plugins are **never** exposed as standalone binaries on `$PATH`. Running `wave-flow` directly does nothing. The only entry point is `wave <plugin>`.

---

## Tech Stack

| Layer | Choice | Why |
|-------|--------|-----|
| Language | Go 1.22+ | Single binary, fast startup, easy cross-compile |
| CLI framework | Cobra | Industry standard, supports dynamic sub-commands |
| Config parsing | Viper + `BurntSushi/toml` | Viper for CLI flag binding (--debug, --verbose); TOML for config files |
| Logging | `slog` (stdlib) | Structured, zero-dep |
| Plugin source | GitHub Releases | Familiar, uses release assets for distribution |
| Plugin transport | OS exec (stdin/stdout JSON) | Simple, language-agnostic |

Viper is used for binding CLI persistent flags (`--debug`, `--verbose`, etc.) to configuration. TOML files are parsed with `BurntSushi/toml`. Plugin parameters come only from the config files — CLI flags are reserved for wave-core's own behavior.

---

## 1. Project Structure

Implementation convention:
- Each Cobra command lives in its own file under `cmd/` (one file per command), so commands stay modular and easy to test.

```
wave-core/
├── cmd/
│   ├── root.go              # Cobra root command
│   ├── install.go           # `wave install <org/plugin>`
│   ├── uninstall.go         # `wave uninstall <plugin>`
│   ├── list.go              # `wave list` — show installed plugins
│   └── init.go              # `wave init` — scaffold a Wavefile
├── internal/
│   ├── config/
│   │   ├── config.go        # Unified config loader
│   │   ├── wavefile.go      # Wavefile schema & parser
│   │   └── global.go        # Global config schema & parser
│   ├── plugin/
│   │   ├── registry.go      # Plugin discovery & resolution
│   │   ├── executor.go      # Fork/exec plugin binaries
│   │   ├── installer.go     # Download from GitHub releases
│   │   └── assets.go        # Asset management (extract, store, verify)
│   ├── errors/
│   │   ├── protocol.go      # Error protocol types (code + message)
│   │   ├── handler.go       # Catch, format, display errors
│   │   └── logger.go        # Write errors to log files
│   └── ui/
│       └── printer.go       # Colored, structured terminal output
├── pkg/
│   └── sdk/
│       └── sdk.go           # Lightweight SDK plugins can import for err() helper
├── go.mod
├── go.sum
├── main.go                  # Entry point — calls cmd.Execute()
└── plan.md
```

---

## 2. Configuration System

Configuration is read from two TOML files for plugin parameters. Wave-core's own behavior flags (`--debug`, `--verbose`) are handled by Viper-bound CLI flags via Cobra. Plugin parameters are never overridden by CLI flags — what's in the files is what plugins get.

### 2.1 Global Config — `~/.wave/config`

Lives at `$HOME/.wave/config`. If it doesn't exist, wave-core creates it automatically on first run (along with required directories like `~/.wave/`, `~/.wave/plugins/`, and the configured logs directory).

```toml
# ~/.wave/config

[core]
logs_dir = "~/.wave/logs"

# Directories wave scans for projects containing Wavefiles.
# Useful when working across multiple workspaces.
[projects]
folders = [
    "~/projects",
    "~/work",
]

[plugins]
# org/name = version
"wave-cli/flow" = "1.2.0"
"wave-cli/test" = "0.5.3"
```

**What it stores:**
- Logs directory location
- Project folder locations (one or more directories where wave looks for projects)
- Installed plugin manifest (name -> version)

### 2.2 Project Config — `Wavefile`

Lives at the project root. Discovered by walking up from `cwd`.

```toml
# Wavefile

[project]
name = "my-app"
version = "1.0.0"
owner = "bouajila"
category = "backend"
tags = ["api", "microservice", "go"]

# Plugin-specific parameters — each plugin gets its own section
[flow]
environment = "staging"
port = 3000
watch = true

[test]
coverage = true
threshold = 80
```

**What it stores:**
- Project metadata (name, version, owner, category, tags)
- Per-plugin parameters (each `[plugin-name]` section is passed to that plugin as-is)

### 2.3 Config Read Strategy

```
Plugin parameters come from ONLY two sources:
  1. Wavefile            (project-level, per-plugin sections)
  2. ~/.wave/config      (global settings, installed plugin list)

Wave-core's own flags (--debug, --verbose) are CLI-only via Cobra/Viper.
Plugin parameters are never overridden by CLI flags or env vars.
Wavefile values take precedence over global config for the same key.
```

**Implementation — `internal/config/config.go`:**

```go
type Config struct {
    Core        CoreConfig                    // logs_dir
    Projects    ProjectsConfig                // folders list
    Plugins     map[string]string             // name -> version (from global)
    Sections    map[string]map[string]any     // per-plugin params from Wavefile
    ProjectMeta ProjectMeta                   // [project] from Wavefile
    Debug       bool                          // --debug flag (via Viper)
    Verbose     bool                          // --verbose flag (via Viper)
}

type CoreConfig struct {
    LogsDir string `toml:"logs_dir"`
}

type ProjectsConfig struct {
    Folders []string `toml:"folders"`
}

type ProjectMeta struct {
    Name     string   `toml:"name"`
    Version  string   `toml:"version"`
    Owner    string   `toml:"owner"`
    Category string   `toml:"category"`
    Tags     []string `toml:"tags"`
}

func Load() (*Config, error) {
    // 1. Parse global ~/.wave/config with BurntSushi/toml
    // 2. Find Wavefile by walking up from cwd
    // 3. Parse Wavefile
    // 4. Merge: global < wavefile (wavefile wins for overlapping keys)
    // 5. Bind Viper CLI flags (--debug, --verbose) into Config
    // 6. Return unified Config
}
```

---

## 3. Plugin System

### 3.1 Naming Convention

Plugins follow a git-style naming: `<org>/<name>`.

- `wave install wave-cli/flow`

The full `org/name` is always required. There is no shorthand.

### 3.2 The `Waveplugin` Metadata File

Every plugin release **must** include a `Waveplugin` file in its release assets. This is how wave-core understands what a plugin is.

```toml
# Waveplugin — ships alongside the plugin binary in the release

[plugin]
name = "flow"
version = "1.2.0"
description = "Development workflow automation for wave projects"
creator = "wave-cli"
license = "MIT"
homepage = "https://github.com/wave-cli/flow"

[compatibility]
min_wave_version = "0.1.0"

[assets]
# Declares what files this plugin ships (besides the binary itself).
# wave-core downloads and manages all of these.
files = [
    "templates/",
    "defaults.toml",
]
```

**What it stores:**
- Plugin identity (name, version, description, creator)
- License and homepage
- Minimum compatible wave-core version
- Additional asset files the plugin needs (templates, default configs, etc.)

wave-core reads this file during install to validate compatibility and to know which assets to download and manage.

### 3.3 Plugin Storage

Installed plugins and their assets live at:

```
~/.wave/plugins/
├── wave-cli/
│   └── flow/
│       ├── v1.2.0/
│       │   ├── bin/
│       │   │   └── flow             # the binary (NOT on PATH, NOT called wave-flow)
│       │   ├── assets/
│       │   │   ├── templates/       # plugin-declared assets
│       │   │   └── defaults.toml
│       │   └── Waveplugin           # metadata file
│       └── current -> v1.2.0/       # symlink to active version
└── wave-cli/
    └── test/
        ├── v0.5.3/
        │   ├── bin/
        │   │   └── test
        │   ├── assets/
        │   └── Waveplugin
        └── current -> v0.5.3/
```

Key points:
- The binary lives inside `bin/` and is **never** added to `$PATH`
- Additional assets are stored in `assets/` and managed by wave-core
- `Waveplugin` metadata is preserved for runtime queries
- `current` symlink points to the active version (enables rollback)

### 3.4 Plugin Discovery & Execution

When the user runs `wave flow dev`:

```
1. Cobra sees "flow" is not a built-in command
2. Root command's RunE triggers plugin discovery
3. Look up "flow" in installed plugins -> found as "wave-cli/flow" v1.2.0
4. Resolve binary: ~/.wave/plugins/wave-cli/flow/current/bin/flow
5. Read Waveplugin metadata for the active version
6. Build execution context:
   - argv:        ["flow", "dev"]
   - stdin:       JSON blob of [flow] section from Wavefile
   - env:
       WAVE_PLUGIN_NAME=flow
       WAVE_PLUGIN_VERSION=1.2.0
       WAVE_PLUGIN_DIR=~/.wave/plugins/wave-cli/flow/current/
       WAVE_PLUGIN_ASSETS=~/.wave/plugins/wave-cli/flow/current/assets/
       WAVE_PROJECT_ROOT=/path/to/project
7. Fork/exec the binary
8. Stream stdout/stderr to terminal in real-time
9. On exit, parse stderr for structured wave errors
```

**Critical: `wave flow` works. Running `flow` or `wave-flow` directly does NOT.** The binary is buried inside `~/.wave/plugins/` and never exposed.

**Implementation — `internal/plugin/executor.go`:**

```go
type ExecResult struct {
    ExitCode int
    Error    *PluginError  // parsed from stderr if structured wave error found
}

func Execute(pluginName string, args []string, cfg *config.Config) (*ExecResult, error) {
    binPath := registry.Resolve(pluginName)  // -> ~/.wave/plugins/<org>/<name>/current/bin/<name>

    // Serialize the plugin's config section as JSON via stdin
    pluginConfig := cfg.Sections[pluginName]
    configJSON, _ := json.Marshal(pluginConfig)

    cmd := exec.Command(binPath, args...)
    cmd.Stdin = bytes.NewReader(configJSON)
    cmd.Stdout = os.Stdout  // stream directly to terminal
    cmd.Env = buildPluginEnv(pluginName, cfg)

    // Capture stderr for error parsing while also showing it
    var stderrBuf bytes.Buffer
    cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

    err := cmd.Run()

    // Parse stderr for structured wave errors
    if err != nil {
        if pluginErr := errors.ParseStderr(stderrBuf.Bytes()); pluginErr != nil {
            return &ExecResult{ExitCode: cmd.ProcessState.ExitCode(), Error: pluginErr}, nil
        }
    }
    return &ExecResult{ExitCode: cmd.ProcessState.ExitCode()}, err
}
```

### 3.5 Plugin Installation — `wave install`

Modeled after `dotnet tool install`. Downloads plugin releases from GitHub.

```
wave install wave-cli/flow              # latest release
wave install wave-cli/flow@1.2.0        # specific version
```

**Installation flow:**

```
1. Parse full name
   - Input must be "org/name" format (e.g. "wave-cli/flow")
   - Optional @version suffix (e.g. "wave-cli/flow@1.2.0")

2. Query GitHub Releases API
   - GET https://api.github.com/repos/wave-cli/flow/releases/latest
   - Or for a specific version: /releases/tags/v1.2.0

3. Find the right assets
   - Look for the Waveplugin file in release assets -> download and parse it
   - Determine OS/arch binary:  flow-linux-amd64, flow-darwin-arm64, etc.
   - Identify additional assets declared in Waveplugin [assets].files

4. Download all assets
   - Binary -> ~/.wave/plugins/wave-cli/flow/v1.2.0/bin/flow
   - Assets -> ~/.wave/plugins/wave-cli/flow/v1.2.0/assets/
   - Waveplugin -> ~/.wave/plugins/wave-cli/flow/v1.2.0/Waveplugin

5. Set permissions
   - chmod +x on the binary

6. Create/update symlink
   - current -> v1.2.0/

7. Update global config
   - Add "wave-cli/flow" = "1.2.0" to [plugins] in ~/.wave/config

8. Validate compatibility
   - Check Waveplugin [compatibility].min_wave_version against current wave version
   - Warn if incompatible
```

**Asset management by wave-core:**

wave-core is responsible for the full lifecycle of plugin assets:
- **Download**: Fetches all declared assets from the GitHub release
- **Store**: Places them in the versioned `assets/` directory
- **Provide**: Passes `WAVE_PLUGIN_ASSETS` env var so the plugin knows where its assets are
- **Clean up**: Removes assets when a plugin version is uninstalled

### 3.6 Dynamic Cobra Registration

On startup, wave-core reads the installed plugins list and registers each as a Cobra command dynamically:

```go
func RegisterPlugins(rootCmd *cobra.Command, cfg *config.Config) {
    for fullName, version := range cfg.Plugins {
        pluginName := extractShortName(fullName) // "wave-cli/flow" -> "flow"

        // Read Waveplugin metadata for description
        meta, _ := ReadWaveplugin(fullName)

        cmd := &cobra.Command{
            Use:                pluginName,
            Short:              meta.Description,
            DisableFlagParsing: true,  // pass ALL args through to plugin
            RunE: func(cmd *cobra.Command, args []string) error {
                result, err := plugin.Execute(pluginName, args, cfg)
                if err != nil {
                    return err
                }
                return handleResult(result)
            },
        }
        rootCmd.AddCommand(cmd)
    }
}
```

This means `wave --help` shows all installed plugins with their descriptions from `Waveplugin`.

---

## 4. Error Handling Protocol

### 4.1 The Protocol

Plugins communicate errors by writing a JSON object to **stderr** with a specific structure:

```json
{
    "wave_error": true,
    "code": "FLOW_ENV_MISSING",
    "message": "Environment 'staging' is not configured",
    "details": "Check your Wavefile [flow] section"
}
```

Plugins using the Go SDK call:

```go
// pkg/sdk/sdk.go
func Err(code string, message string) {
    e := WaveError{
        WaveError: true,
        Code:      code,
        Message:   message,
    }
    json.NewEncoder(os.Stderr).Encode(e)
    os.Exit(1)
}
```

Non-Go plugins just need to write the same JSON to stderr — the protocol is language-agnostic.

### 4.2 Error Display

wave-core parses stderr, detects the `wave_error: true` marker, and renders:

```
 ERROR [FLOW_ENV_MISSING]
   Environment 'staging' is not configured

   Check your Wavefile [flow] section

   Plugin: flow v1.2.0
   Logged: ~/.wave/logs/2026-03-14.log
```

If stderr is **not** valid wave error JSON, wave-core treats it as an unstructured crash and displays the raw stderr with a generic error wrapper.

### 4.3 Error Logging

All errors (structured and unstructured) are appended to a daily log file at the location specified in `~/.wave/config` `[core].logs_dir`:

```
~/.wave/logs/2026-03-14.log
```

Log format (JSON Lines):

```json
{"ts":"2026-03-14T10:30:00Z","plugin":"flow","code":"FLOW_ENV_MISSING","message":"Environment 'staging' is not configured","args":["dev"],"cwd":"/home/user/my-app"}
```

**Implementation — `internal/errors/logger.go`:**

```go
func LogError(pluginName string, err *PluginError, args []string) {
    logsDir := config.Global().Core.LogsDir  // from ~/.wave/config [core].logs_dir

    filename := time.Now().Format("2006-01-02") + ".log"

    entry := LogEntry{
        Timestamp: time.Now(),
        Plugin:    pluginName,
        Code:      err.Code,
        Message:   err.Message,
        Args:      args,
        Cwd:       os.Getwd(),
    }

    f, _ := os.OpenFile(filepath.Join(logsDir, filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    defer f.Close()
    json.NewEncoder(f).Encode(entry)
}
```

---

## 5. Built-in Commands

| Command | Description |
|---------|-------------|
| `wave init` | Create a `Wavefile` in the current directory |
| `wave install <org/plugin>[@version]` | Download and install a plugin from GitHub releases |
| `wave uninstall <plugin>` | Remove a plugin and its assets |
| `wave list` | List installed plugins with version and metadata |
| `wave <plugin> [args...]` | Execute a plugin (dynamic, the only way to run plugins) |

Notes:
- `~/.wave/config` is auto-created on first run; there is no `init --global`.

---

## 6. Implementation Order

### Phase 1 — Skeleton & Config
1. `go mod init`, set up project structure
2. Implement `main.go` + Cobra root command with persistent flags (`--debug`, `--verbose`) bound via Viper
3. Implement first-run bootstrap: ensure `~/.wave/` exists; if `~/.wave/config` missing, create with defaults
4. Implement global config reader/writer with `BurntSushi/toml`
5. Implement Wavefile discovery (walk up from cwd) and parser
6. Implement config merge logic (wavefile > global for same keys)
7. `wave init` command (project Wavefile scaffolding)

### Phase 2 — Plugin Execution
7. `Waveplugin` metadata parser
8. Plugin registry (resolve name -> binary path via `~/.wave/plugins/`)
9. Plugin executor (fork/exec, stdin config, env vars, stderr capture)
10. Dynamic Cobra sub-command registration from installed plugins
11. Test with a small “echo” plugin that prints results directly

### Phase 3 — Error Handling
12. Error protocol types (`WaveError` struct)
13. Stderr parser (detect structured JSON vs raw crash output)
14. Error display formatting (colored terminal output)
15. Error file logger (JSON Lines to `logs_dir`)
16. Go SDK package (`pkg/sdk`) with `Err()` helper

### Phase 4 — Plugin Management (dotnet-style)
17. GitHub Releases API client
18. `wave install` — download binary + Waveplugin + assets for correct OS/arch
19. Asset management — store, provide path, clean up
20. `wave uninstall` — remove plugin dir, update global config
21. `wave list` — show installed plugins with Waveplugin metadata
22. Version pinning and `current` symlink management

### Phase 5 — Polish
23. Help text (pull descriptions from Waveplugin metadata)
24. Shell completions for installed plugins
25. CI/CD pipeline for wave-core itself
26. Plugin authoring guide (how to create a Waveplugin file, release assets, etc.)

---

## 8. Testing Strategy (Minimal, End-to-End)

Use a tiny test plugin to validate the whole pipeline early (install -> discover -> execute -> config stdin -> output -> errors).

Test plugin goals:
- Reads stdin (JSON) and prints a single line to stdout, e.g. `OK <plugin> <subcommand> <key=value>`
- On a specific argument (e.g. `fail`), emits a structured wave error JSON to stderr and exits non-zero

This validates:
- Dynamic Cobra registration and `wave <plugin> ...` routing
- Passing Wavefile section to plugin via stdin
- Streaming stdout/stderr
- Structured error parsing + pretty printing
- Error log file creation in `logs_dir`

---

## 7. Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **TOML for config files** | More readable than YAML for flat key-value, less noisy than JSON |
| **Viper for CLI flags only** | Binds `--debug`, `--verbose` etc. to Cobra; plugin parameters stay file-only |
| **Plugin params only from files** | Predictable behavior — what's in the files is what plugins get, no hidden overrides |
| **Exec over RPC** | Simpler, language-agnostic, no daemon process needed |
| **Config via stdin** | Avoids temp files, plugins can parse once on startup |
| **JSON error protocol on stderr** | Clean separation from plugin stdout; parseable yet simple |
| **Plugins never on PATH** | Prevents accidental direct execution; `wave <plugin>` is the only entry point |
| **Waveplugin metadata file** | Self-describing plugins — wave-core knows version, creator, compatibility, and assets without external registries |
| **wave-core manages assets** | Plugins declare what they need; wave-core downloads, stores, and provides paths — plugins don't manage their own file layout |
| **GitHub Releases as source** | No custom registry infrastructure; works with existing GitHub workflows (like dotnet tools use NuGet) |
| **Symlinked `current` version** | Enables instant rollback and multi-version coexistence |
| **Multiple project folders** | Users often work across several directories; wave can discover Wavefiles in all of them |
| **Daily log rotation** | Simple, no external log rotation needed |
| **Full org/name always required** | Explicit, no ambiguity — prevents confusion about where a plugin comes from |
| **Owner, category, tags in Wavefile** | Projects are self-describing; enables future features like filtering, search, dashboards |
