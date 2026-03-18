# wave-core Testing Guide

## Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run with coverage
go test ./... -cover

# Run specific package
go test ./internal/config/ -v

# Run E2E tests
go test ./e2e/ -v

# Run a specific test by name
go test ./internal/errors/ -run TestParseStderrStructured -v
```

## Test Organization

Tests follow Go conventions: `*_test.go` files alongside implementation in the same package.

### Unit Tests

Each internal package has comprehensive unit tests:

```
internal/version/version_test.go      # Version info and semver tests
internal/ui/printer_test.go           # Printer output at each verbosity level
internal/config/global_test.go        # Global config parse/write/defaults
internal/config/wavefile_test.go      # Wavefile parse/discover
internal/bootstrap/bootstrap_test.go  # First-run setup, idempotency
internal/errors/errors_test.go        # Error protocol, parsing, formatting, logging
internal/pluginmgmt/pluginmgmt_test.go # Waveplugin, PluginRef, Registry
internal/downloader/downloader_test.go # HTTP mocking, tar.gz extraction, install
internal/executor/runner_test.go      # Plugin execution, env vars, stdin
pkg/sdk/sdk_test.go                   # SDK config reading, error formatting
```

### E2E Tests

End-to-end tests in `e2e/e2e_test.go` exercise the full pipeline:

```
e2e/e2e_test.go
  TestE2E_BootstrapCreatesDirectoryStructure  # Verifies ~/.wave setup
  TestE2E_InitCreatesWavefile                 # Wavefile scaffolding
  TestE2E_InitWavefileDiscovery               # Walk-up Wavefile discovery
  TestE2E_PluginExecutionWithConfig           # Execute with Wavefile config
  TestE2E_PluginExecutionNoConfig             # Execute without config
  TestE2E_PluginErrorHandling                 # Structured error parsing
  TestE2E_PluginErrorLogging                  # JSONL log file creation
  TestE2E_ErrorFormatDisplay                  # Human-friendly error output
  TestE2E_PluginRegistryResolution            # Binary/assets/Waveplugin resolution
  TestE2E_ListInstalledPlugins                # Plugin listing
  TestE2E_UninstallPlugin                     # Plugin removal
  TestE2E_VersionCompatibility                # Semver min version check
  TestE2E_FullPipeline                        # Complete flow: config -> exec -> error -> log
  TestE2E_PluginLookupByShortName             # Short name -> full name resolution
  TestE2E_WavefileMultiplePluginSections      # Multiple plugin configs
```

## Test Patterns

### Mocking HTTP (downloader tests)

The downloader tests use `httptest.NewServer` to mock GitHub API:

```go
func TestFetchRelease(t *testing.T) {
    release := Release{
        TagName: "v1.2.0",
        Assets:  []ReleaseAsset{{Name: "flow-linux-amd64.tar.gz", URL: "..."}},
    }

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(release)
    }))
    defer server.Close()

    client := NewClient(server.URL, "")
    rel, err := client.FetchRelease("wave-cli", "flow", "")
    // assertions...
}
```

### Testing Plugin Execution (runner tests)

Runner tests build the echo test plugin and execute it:

```go
func TestExecuteSimplePlugin(t *testing.T) {
    // Build the test echo plugin from testdata/plugins/echo/
    binPath := buildEchoPlugin(t)

    result, err := Execute(binPath, []string{"hello"}, nil, "echo", "1.0.0", "/tmp")
    if err != nil {
        t.Fatalf("Execute failed: %v", err)
    }
    if result.ExitCode != 0 {
        t.Errorf("ExitCode = %d", result.ExitCode)
    }
    if !strings.Contains(result.Stdout, "OK") {
        t.Errorf("Stdout should contain OK")
    }
}
```

### E2E Test Setup Pattern

E2E tests create a complete fake wave home with installed plugins:

```go
func TestE2E_FullPipeline(t *testing.T) {
    echoBin := buildEchoPlugin(t)
    homeDir, _ := setupFakeWaveHome(t, echoBin)

    // Create a project with Wavefile
    projectDir := filepath.Join(homeDir, "myproject")
    os.MkdirAll(projectDir, 0755)
    os.WriteFile(filepath.Join(projectDir, "Wavefile"), []byte(`
        [project]
        name = "test"
        version = "1.0.0"

        [echo]
        environment = "staging"
    `), 0644)

    // Parse, execute, verify
    wf, _ := config.ParseWavefile(filepath.Join(projectDir, "Wavefile"))
    result, _ := runner.Execute(echoBin, []string{"dev"}, wf.Sections["echo"], ...)
    // assertions...
}
```

### Test Echo Plugin

Located at `testdata/plugins/echo/main.go`. Behavior:

| Input | Output |
|-------|--------|
| `echo hello` | `OK echo hello` |
| `echo dev` with stdin `{"port":3000}` | `OK echo dev port=3000` |
| `echo fail` | stderr: `{"wave_error":true,"code":"ECHO_FAIL",...}`, exit 1 |

## Edge Cases Tested

- Nil/empty inputs (empty configs, nil sections, empty plugin lists)
- Invalid TOML/JSON (malformed config files)
- Missing files (nonexistent paths, missing Wavefile)
- Network errors (HTTP 404/500 via mock servers)
- Path traversal (tar.gz extraction security)
- Direct plugin resolution (no version directory)
- Concurrent-safe (each test uses `t.TempDir()`)
- Unicode and special characters in config values

## Adding New Tests

1. **Unit test**: Add to the `_test.go` file in the same package
2. **Integration test**: Add to the package test file using `httptest` for HTTP
3. **E2E test**: Add to `e2e/e2e_test.go` following the existing patterns

Always follow the TDD cycle:
1. Write failing test (RED)
2. Implement minimum code (GREEN)
3. Refactor (IMPROVE)
