# wave-core - Task Tracker

## Status Legend
- [x] Complete (tests passing, coverage 80%+)
- [~] In Progress
- [ ] Not Started

---

## Phase 1 - Core Infrastructure (Complete)

- [x] `internal/version/` - Build-time version info, semver comparison
  - Tests: `version_test.go` - 10 tests
  - Coverage: 90%+
  - Functions: `Get()`, `String()`, `Full()`, `SatisfiesMin()`

- [x] `internal/ui/` - Terminal output with verbosity levels
  - Tests: `printer_test.go` - 12 tests
  - Coverage: 95%+
  - Types: `Printer`, `Level` (Quiet/Normal/Verbose/Debug)
  - Methods: `Error`, `Warn`, `Info`, `Success`, `Verbose`, `Debug`

- [x] `internal/config/` - Configuration parsing (global + Wavefile)
  - Tests: `global_test.go` (6 tests), `wavefile_test.go` (6 tests)
  - Coverage: 90%+
  - Functions: `ParseGlobalConfig`, `WriteGlobalConfig`, `DefaultGlobalConfig`, `ParseWavefile`, `DiscoverWavefile`

- [x] `internal/bootstrap/` - First-run directory/config setup
  - Tests: `bootstrap_test.go` - 7 tests
  - Coverage: 85%+
  - Functions: `Ensure(homeDir)` - idempotent, preserves existing config

- [x] `internal/errors/` - Error protocol, parsing, formatting, logging
  - Tests: `errors_test.go` - 12 tests
  - Coverage: 90%+
  - Functions: `ParseStderr`, `FormatError`, `LogError`

- [x] `internal/pluginmgmt/` - Plugin metadata, name resolution, registry
  - Tests: `pluginmgmt_test.go` - 15 tests
  - Coverage: 90.2%
  - Functions: `ParseWaveplugin`, `ParsePluginRef`, `NewRegistry`, `ResolveBinary`, `ResolveAssets`, `ReadWaveplugin`, `ListInstalled`

---

## Phase 2 - Plugin Execution

- [x] `internal/downloader/` - GitHub Releases client + asset selection
  - Tests: `downloader_test.go`
  - Functions: `NewClient`, `FetchRelease`, `SelectAsset`, `Download`, `ExtractTarGz`, `InstallPlugin`

- [x] `internal/executor/` - Plugin executor (fork/exec, stdin JSON, env vars)
  - Tests: `runner_test.go`
  - Functions: `Execute`, `BuildEnv`, `BuildStdin`

---

## Phase 3 - CLI Commands

- [x] `cmd/root.go` - Cobra root command, Viper flag binding, dynamic plugin dispatch
- [x] `cmd/version.go` - `wave version` / `wave --version`
- [x] `cmd/init.go` - `wave init` scaffold Wavefile
- [x] `cmd/install.go` - `wave install <org/plugin>[@version]`
- [x] `cmd/uninstall.go` - `wave uninstall <plugin>`
- [x] `cmd/list.go` - `wave list` installed plugins
- [x] `main.go` - Entry point

---

## Phase 4 - SDK & Test Plugin

- [x] `pkg/sdk/` - Go SDK for plugin authors (`Err()`, `ReadConfig()`)
  - Tests: `sdk_test.go`

- [x] `testdata/plugins/echo/` - Test echo plugin binary
  - Reads stdin JSON, prints `OK <plugin> <subcmd> <key=value>`
  - On `fail` arg, emits structured wave error

---

## Phase 5 - E2E Tests

- [x] `e2e/` - End-to-end integration tests
  - `e2e/e2e_test.go` - Full pipeline tests:
    - Bootstrap creates directory structure
    - Init scaffolds Wavefile
    - Plugin execution with config via stdin
    - Plugin error handling and logging
    - List installed plugins
    - Uninstall removes plugin
    - Version command output
    - Dynamic plugin dispatch

---

## Phase 6 - Documentation

- [x] `docs/architecture.md` - Architecture overview with test examples
- [x] `docs/testing.md` - Testing guide with commands and patterns
- [x] `docs/plugin-authoring.md` - How to create wave plugins

---

## Phase 7 - Git & CI

- [x] Initialize git repo
- [x] Push to `Wave-cli/wave-core` on GitHub
