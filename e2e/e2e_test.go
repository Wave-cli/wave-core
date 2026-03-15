// Package e2e contains end-to-end tests for wave-core.
// These tests exercise the full pipeline: bootstrap -> config -> plugin execution -> error handling.
package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/wave-cli/wave-core/internal/bootstrap"
	"github.com/wave-cli/wave-core/internal/config"
	"github.com/wave-cli/wave-core/internal/errors"
	"github.com/wave-cli/wave-core/internal/pluginmgmt"
	"github.com/wave-cli/wave-core/internal/runner"
)

// getProjectRoot finds the project root by walking up to find go.mod.
func getProjectRoot(t *testing.T) string {
	t.Helper()
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (go.mod)")
		}
		dir = parent
	}
}

// buildEchoPlugin compiles the test echo plugin and returns the binary path.
func buildEchoPlugin(t *testing.T) string {
	t.Helper()
	projectRoot := getProjectRoot(t)
	echoSrc := filepath.Join(projectRoot, "testdata", "plugins", "echo")

	binDir := t.TempDir()
	binName := "echo"
	if runtime.GOOS == "windows" {
		binName = "echo.exe"
	}
	binPath := filepath.Join(binDir, binName)

	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = echoSrc
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build echo plugin: %v\n%s", err, out)
	}
	return binPath
}

// setupFakeWaveHome creates a complete wave home directory with a fake
// installed plugin pointing to the echo binary.
func setupFakeWaveHome(t *testing.T, echoBinPath string) (homeDir string, pluginsDir string) {
	t.Helper()
	homeDir = t.TempDir()

	// Bootstrap creates dirs and config
	gc, err := bootstrap.Ensure(homeDir)
	if err != nil {
		t.Fatalf("bootstrap.Ensure failed: %v", err)
	}

	pluginsDir = filepath.Join(homeDir, ".wave", "plugins")

	// Install fake plugin: wave-cli/echo v1.0.0
	versionDir := filepath.Join(pluginsDir, "wave-cli", "echo", "v1.0.0")
	binDir := filepath.Join(versionDir, "bin")
	assetsDir := filepath.Join(versionDir, "assets")
	os.MkdirAll(binDir, 0755)
	os.MkdirAll(assetsDir, 0755)

	// Copy echo binary
	data, _ := os.ReadFile(echoBinPath)
	destBin := filepath.Join(binDir, "echo")
	os.WriteFile(destBin, data, 0755)

	// Write Waveplugin
	wpContent := `[plugin]
name = "echo"
version = "1.0.0"
description = "Test echo plugin for E2E"
creator = "wave-cli"
license = "MIT"
homepage = "https://github.com/wave-cli/echo"

[compatibility]
min_wave_version = "0.1.0"
`
	os.WriteFile(filepath.Join(versionDir, "Waveplugin"), []byte(wpContent), 0644)

	// Create current symlink
	currentLink := filepath.Join(pluginsDir, "wave-cli", "echo", "current")
	os.Symlink(versionDir, currentLink)

	// Update global config with plugin
	gc.Plugins["wave-cli/echo"] = "1.0.0"
	configPath := filepath.Join(homeDir, ".wave", "config")
	config.WriteGlobalConfig(configPath, gc)

	return homeDir, pluginsDir
}

// =============================================================================
// E2E Test: Full Bootstrap Pipeline
// =============================================================================

func TestE2E_BootstrapCreatesDirectoryStructure(t *testing.T) {
	homeDir := t.TempDir()

	gc, err := bootstrap.Ensure(homeDir)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	// Verify all required directories exist
	dirs := []string{
		filepath.Join(homeDir, ".wave"),
		filepath.Join(homeDir, ".wave", "plugins"),
		gc.Core.LogsDir,
	}
	for _, d := range dirs {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			t.Errorf("Directory should exist: %s", d)
		}
	}

	// Verify config file exists and is valid
	configPath := filepath.Join(homeDir, ".wave", "config")
	readGc, err := config.ParseGlobalConfig(configPath)
	if err != nil {
		t.Fatalf("Config is not valid: %v", err)
	}
	if readGc.Core.LogsDir == "" {
		t.Error("Config should have logs_dir set")
	}
}

// =============================================================================
// E2E Test: Init Command
// =============================================================================

func TestE2E_InitCreatesWavefile(t *testing.T) {
	projectDir := t.TempDir()
	wavefilePath := filepath.Join(projectDir, "Wavefile")

	// Manually write a Wavefile (simulating init)
	projectName := filepath.Base(projectDir)
	content := "[project]\nname = \"" + projectName + "\"\nversion = \"0.1.0\"\n"
	os.WriteFile(wavefilePath, []byte(content), 0644)

	// Parse it back
	wf, err := config.ParseWavefile(wavefilePath)
	if err != nil {
		t.Fatalf("ParseWavefile failed: %v", err)
	}
	if wf.Project.Name != projectName {
		t.Errorf("Project name = %q, want %q", wf.Project.Name, projectName)
	}
}

func TestE2E_InitWavefileDiscovery(t *testing.T) {
	// Create: root/sub1/sub2/
	root := t.TempDir()
	sub := filepath.Join(root, "sub1", "sub2")
	os.MkdirAll(sub, 0755)

	wavefilePath := filepath.Join(root, "Wavefile")
	os.WriteFile(wavefilePath, []byte("[project]\nname=\"e2e-test\"\nversion=\"1.0.0\"\n"), 0644)

	// Discover from deepest directory
	found, err := config.DiscoverWavefile(sub)
	if err != nil {
		t.Fatalf("DiscoverWavefile failed: %v", err)
	}
	if found != wavefilePath {
		t.Errorf("Found %q, want %q", found, wavefilePath)
	}
}

// =============================================================================
// E2E Test: Plugin Execution with Config via Stdin
// =============================================================================

func TestE2E_PluginExecutionWithConfig(t *testing.T) {
	echoBin := buildEchoPlugin(t)

	section := map[string]any{
		"environment": "production",
		"port":        int64(8080),
		"debug":       true,
	}

	result, err := runner.Execute(echoBin, []string{"deploy"}, section, "echo", "1.0.0", "/tmp/project")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0\nStderr: %s", result.ExitCode, result.Stderr)
	}

	stdout := result.Stdout
	if !strings.Contains(stdout, "OK echo deploy") {
		t.Errorf("Stdout should contain 'OK echo deploy', got: %q", stdout)
	}
	if !strings.Contains(stdout, "environment=production") {
		t.Errorf("Stdout should contain config values, got: %q", stdout)
	}
	if !strings.Contains(stdout, "port=8080") {
		t.Errorf("Stdout should contain port, got: %q", stdout)
	}
}

func TestE2E_PluginExecutionNoConfig(t *testing.T) {
	echoBin := buildEchoPlugin(t)

	result, err := runner.Execute(echoBin, []string{"status"}, nil, "echo", "1.0.0", "/tmp")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d", result.ExitCode)
	}
	if !strings.Contains(result.Stdout, "OK echo status") {
		t.Errorf("Stdout = %q", result.Stdout)
	}
}

// =============================================================================
// E2E Test: Plugin Error Handling
// =============================================================================

func TestE2E_PluginErrorHandling(t *testing.T) {
	echoBin := buildEchoPlugin(t)

	result, err := runner.Execute(echoBin, []string{"fail"}, nil, "echo", "1.0.0", "/tmp")
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	// Should have non-zero exit
	if result.ExitCode == 0 {
		t.Error("ExitCode should be non-zero for 'fail' command")
	}

	// Should have structured error
	if result.PluginError == nil {
		t.Fatal("PluginError should be parsed from stderr")
	}
	if result.PluginError.Code != "ECHO_FAIL" {
		t.Errorf("Error code = %q, want ECHO_FAIL", result.PluginError.Code)
	}
	if result.PluginError.Message != "intentional failure for testing" {
		t.Errorf("Error message = %q", result.PluginError.Message)
	}
}

func TestE2E_PluginErrorLogging(t *testing.T) {
	logsDir := t.TempDir()

	pe := &errors.PluginError{
		WaveError: true,
		Code:      "E2E_TEST_ERR",
		Message:   "end-to-end test error",
		Details:   "this is a test",
	}

	err := errors.LogError(logsDir, "echo", pe, []string{"fail"})
	if err != nil {
		t.Fatalf("LogError failed: %v", err)
	}

	// Verify log file exists
	entries, _ := os.ReadDir(logsDir)
	if len(entries) == 0 {
		t.Fatal("No log file created")
	}

	// Read and verify content
	data, _ := os.ReadFile(filepath.Join(logsDir, entries[0].Name()))
	var logEntry map[string]any
	if err := json.Unmarshal(data, &logEntry); err != nil {
		t.Fatalf("Log entry is not valid JSON: %v", err)
	}
	if logEntry["code"] != "E2E_TEST_ERR" {
		t.Errorf("Logged code = %v", logEntry["code"])
	}
	if logEntry["plugin"] != "echo" {
		t.Errorf("Logged plugin = %v", logEntry["plugin"])
	}
}

func TestE2E_ErrorFormatDisplay(t *testing.T) {
	pe := &errors.PluginError{
		WaveError: true,
		Code:      "DEPLOY_FAILED",
		Message:   "deployment timed out",
		Details:   "check network connectivity",
	}

	output := errors.FormatError("echo", "1.0.0", pe, "/tmp/logs/2026-03-15.log")

	if !strings.Contains(output, "DEPLOY_FAILED") {
		t.Error("Should contain error code")
	}
	if !strings.Contains(output, "deployment timed out") {
		t.Error("Should contain message")
	}
	if !strings.Contains(output, "check network connectivity") {
		t.Error("Should contain details")
	}
	if !strings.Contains(output, "echo v1.0.0") {
		t.Error("Should contain plugin name and version")
	}
	if !strings.Contains(output, "/tmp/logs/2026-03-15.log") {
		t.Error("Should contain log path")
	}
}

// =============================================================================
// E2E Test: Plugin Registry Resolution
// =============================================================================

func TestE2E_PluginRegistryResolution(t *testing.T) {
	echoBin := buildEchoPlugin(t)
	_, pluginsDir := setupFakeWaveHome(t, echoBin)

	reg := pluginmgmt.NewRegistry(pluginsDir)

	// Resolve binary
	binPath, err := reg.ResolveBinary("wave-cli/echo")
	if err != nil {
		t.Fatalf("ResolveBinary failed: %v", err)
	}
	if !strings.Contains(binPath, "current/bin/echo") {
		t.Errorf("BinPath = %q, should contain current/bin/echo", binPath)
	}

	// Resolve assets
	assetsPath, err := reg.ResolveAssets("wave-cli/echo")
	if err != nil {
		t.Fatalf("ResolveAssets failed: %v", err)
	}
	if !strings.Contains(assetsPath, "current/assets") {
		t.Errorf("AssetsPath = %q", assetsPath)
	}

	// Read Waveplugin
	wp, err := reg.ReadWaveplugin("wave-cli/echo")
	if err != nil {
		t.Fatalf("ReadWaveplugin failed: %v", err)
	}
	if wp.Plugin.Name != "echo" {
		t.Errorf("Plugin name = %q", wp.Plugin.Name)
	}
	if wp.Plugin.Version != "1.0.0" {
		t.Errorf("Plugin version = %q", wp.Plugin.Version)
	}
}

// =============================================================================
// E2E Test: List Installed Plugins
// =============================================================================

func TestE2E_ListInstalledPlugins(t *testing.T) {
	echoBin := buildEchoPlugin(t)
	homeDir, pluginsDir := setupFakeWaveHome(t, echoBin)

	// Read config
	configPath := filepath.Join(homeDir, ".wave", "config")
	gc, err := config.ParseGlobalConfig(configPath)
	if err != nil {
		t.Fatalf("ParseGlobalConfig failed: %v", err)
	}

	reg := pluginmgmt.NewRegistry(pluginsDir)
	list := reg.ListInstalled(gc.Plugins)

	if len(list) != 1 {
		t.Fatalf("Expected 1 plugin, got %d", len(list))
	}
	if list[0].FullName != "wave-cli/echo" {
		t.Errorf("FullName = %q", list[0].FullName)
	}
	if list[0].Version != "1.0.0" {
		t.Errorf("Version = %q", list[0].Version)
	}
}

// =============================================================================
// E2E Test: Uninstall Plugin
// =============================================================================

func TestE2E_UninstallPlugin(t *testing.T) {
	echoBin := buildEchoPlugin(t)
	homeDir, pluginsDir := setupFakeWaveHome(t, echoBin)

	configPath := filepath.Join(homeDir, ".wave", "config")

	// Verify plugin is installed
	gc, _ := config.ParseGlobalConfig(configPath)
	if _, ok := gc.Plugins["wave-cli/echo"]; !ok {
		t.Fatal("Plugin should be installed before uninstall test")
	}

	// Simulate uninstall: remove directory and update config
	pluginDir := filepath.Join(pluginsDir, "wave-cli", "echo")
	os.RemoveAll(pluginDir)
	delete(gc.Plugins, "wave-cli/echo")
	config.WriteGlobalConfig(configPath, gc)

	// Verify plugin is gone
	gc2, _ := config.ParseGlobalConfig(configPath)
	if _, ok := gc2.Plugins["wave-cli/echo"]; ok {
		t.Error("Plugin should be removed from config after uninstall")
	}
	if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
		t.Error("Plugin directory should be removed after uninstall")
	}
}

// =============================================================================
// E2E Test: Version Compatibility Check
// =============================================================================

func TestE2E_VersionCompatibility(t *testing.T) {
	echoBin := buildEchoPlugin(t)
	_, pluginsDir := setupFakeWaveHome(t, echoBin)

	reg := pluginmgmt.NewRegistry(pluginsDir)
	wp, err := reg.ReadWaveplugin("wave-cli/echo")
	if err != nil {
		t.Fatalf("ReadWaveplugin failed: %v", err)
	}

	// Our test Waveplugin requires min_wave_version = "0.1.0"
	// Current version (dev) should satisfy anything
	from := wp.Compatibility.MinWaveVersion
	if from != "0.1.0" {
		t.Errorf("MinWaveVersion = %q, want 0.1.0", from)
	}
}

// =============================================================================
// E2E Test: Full Pipeline (Config -> Execute -> Parse Error -> Log)
// =============================================================================

func TestE2E_FullPipeline(t *testing.T) {
	echoBin := buildEchoPlugin(t)
	homeDir, _ := setupFakeWaveHome(t, echoBin)

	// 1. Create a project with Wavefile
	projectDir := filepath.Join(homeDir, "myproject")
	os.MkdirAll(projectDir, 0755)

	wavefileContent := `[project]
name = "e2e-project"
version = "2.0.0"
owner = "tester"
category = "testing"
tags = ["e2e", "integration"]

[echo]
environment = "staging"
port = 3000
debug = true
`
	os.WriteFile(filepath.Join(projectDir, "Wavefile"), []byte(wavefileContent), 0644)

	// 2. Parse Wavefile
	wf, err := config.ParseWavefile(filepath.Join(projectDir, "Wavefile"))
	if err != nil {
		t.Fatalf("ParseWavefile failed: %v", err)
	}
	if wf.Project.Name != "e2e-project" {
		t.Errorf("Project name = %q", wf.Project.Name)
	}

	section := wf.Sections["echo"]
	if section == nil {
		t.Fatal("echo section missing from Wavefile")
	}

	// 3. Execute plugin with Wavefile config
	result, err := runner.Execute(echoBin, []string{"dev"}, section, "echo", "1.0.0", projectDir)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d\nStderr: %s", result.ExitCode, result.Stderr)
	}

	// Verify config was passed through
	stdout := result.Stdout
	if !strings.Contains(stdout, "OK echo dev") {
		t.Errorf("Missing OK prefix: %q", stdout)
	}
	if !strings.Contains(stdout, "environment=staging") {
		t.Errorf("Missing environment: %q", stdout)
	}
	if !strings.Contains(stdout, "port=3000") {
		t.Errorf("Missing port: %q", stdout)
	}

	// 4. Test error path
	errResult, err := runner.Execute(echoBin, []string{"fail"}, nil, "echo", "1.0.0", projectDir)
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	if errResult.PluginError == nil {
		t.Fatal("Should have structured error")
	}

	// 5. Log the error
	gc, _ := config.ParseGlobalConfig(filepath.Join(homeDir, ".wave", "config"))
	err = errors.LogError(gc.Core.LogsDir, "echo", errResult.PluginError, []string{"fail"})
	if err != nil {
		t.Fatalf("LogError failed: %v", err)
	}

	// 6. Verify log was written
	logEntries, _ := os.ReadDir(gc.Core.LogsDir)
	if len(logEntries) == 0 {
		t.Error("Log file should exist after error logging")
	}

	// Read log and verify structure
	logData, _ := os.ReadFile(filepath.Join(gc.Core.LogsDir, logEntries[0].Name()))
	var entry map[string]any
	if err := json.Unmarshal(logData, &entry); err != nil {
		t.Fatalf("Log is not valid JSON: %v", err)
	}
	if entry["plugin"] != "echo" {
		t.Errorf("Log plugin = %v", entry["plugin"])
	}
	if entry["code"] != "ECHO_FAIL" {
		t.Errorf("Log code = %v", entry["code"])
	}
}

// =============================================================================
// E2E Test: Plugin Lookup by Short Name
// =============================================================================

func TestE2E_PluginLookupByShortName(t *testing.T) {
	plugins := map[string]string{
		"wave-cli/flow":  "1.2.0",
		"wave-cli/test":  "0.5.3",
		"other-org/echo": "2.0.0",
	}

	fullName, version, found := runner.LookupPlugin("flow", plugins)
	if !found {
		t.Fatal("Should find 'flow' plugin")
	}
	if fullName != "wave-cli/flow" {
		t.Errorf("FullName = %q", fullName)
	}
	if version != "1.2.0" {
		t.Errorf("Version = %q", version)
	}

	_, _, found = runner.LookupPlugin("nonexistent", plugins)
	if found {
		t.Error("Should not find nonexistent plugin")
	}
}

// =============================================================================
// E2E Test: Multiple Plugin Config Sections
// =============================================================================

func TestE2E_WavefileMultiplePluginSections(t *testing.T) {
	content := `[project]
name = "multi-plugin-test"
version = "1.0.0"

[flow]
environment = "production"
port = 8080

[test]
coverage = true
threshold = 90

[deploy]
target = "aws"
region = "us-east-1"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "Wavefile")
	os.WriteFile(path, []byte(content), 0644)

	wf, err := config.ParseWavefile(path)
	if err != nil {
		t.Fatalf("ParseWavefile failed: %v", err)
	}

	if len(wf.Sections) != 3 {
		t.Errorf("Expected 3 sections, got %d", len(wf.Sections))
	}

	// Each plugin should get its own isolated section
	if wf.Sections["flow"]["environment"] != "production" {
		t.Error("flow section missing")
	}
	if wf.Sections["test"]["coverage"] != true {
		t.Error("test section missing")
	}
	if wf.Sections["deploy"]["target"] != "aws" {
		t.Error("deploy section missing")
	}
}
