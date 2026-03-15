package runner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// --- BuildEnv ---

func TestBuildEnv(t *testing.T) {
	env := BuildEnv("flow", "1.2.0", "/home/user/.wave/plugins/wave-cli/flow/current", "/path/to/project")

	expected := map[string]string{
		"WAVE_PLUGIN_NAME":    "flow",
		"WAVE_PLUGIN_VERSION": "1.2.0",
		"WAVE_PLUGIN_DIR":     "/home/user/.wave/plugins/wave-cli/flow/current",
		"WAVE_PLUGIN_ASSETS":  "/home/user/.wave/plugins/wave-cli/flow/current/assets",
		"WAVE_PROJECT_ROOT":   "/path/to/project",
	}

	envMap := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	for key, want := range expected {
		got, ok := envMap[key]
		if !ok {
			t.Errorf("Missing env var %s", key)
			continue
		}
		if got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestBuildEnvInheritsOSEnv(t *testing.T) {
	env := BuildEnv("flow", "1.0.0", "/plugin/dir", "/project")

	// Should include OS env vars (PATH at minimum)
	found := false
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			found = true
			break
		}
	}
	if !found {
		t.Error("BuildEnv should inherit OS environment including PATH")
	}
}

// --- BuildStdin ---

func TestBuildStdin(t *testing.T) {
	section := map[string]any{
		"environment": "staging",
		"port":        int64(3000),
		"watch":       true,
	}

	data, err := BuildStdin(section)
	if err != nil {
		t.Fatalf("BuildStdin failed: %v", err)
	}

	// Should be valid JSON
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("BuildStdin output is not valid JSON: %v", err)
	}

	if parsed["environment"] != "staging" {
		t.Errorf("environment = %v", parsed["environment"])
	}
}

func TestBuildStdinNilSection(t *testing.T) {
	data, err := BuildStdin(nil)
	if err != nil {
		t.Fatalf("BuildStdin(nil) failed: %v", err)
	}

	// Should produce empty JSON object
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Not valid JSON: %v", err)
	}
	if len(parsed) != 0 {
		t.Errorf("Expected empty object for nil section, got %v", parsed)
	}
}

func TestBuildStdinEmptySection(t *testing.T) {
	data, err := BuildStdin(map[string]any{})
	if err != nil {
		t.Fatalf("BuildStdin({}) failed: %v", err)
	}
	if string(data) != "{}" {
		t.Errorf("Expected '{}', got %q", string(data))
	}
}

// --- Execute (with real test binary) ---

func TestExecuteSimplePlugin(t *testing.T) {
	// Build the test echo plugin
	echoDir := filepath.Join(getProjectRoot(t), "testdata", "plugins", "echo")
	if _, err := os.Stat(filepath.Join(echoDir, "main.go")); os.IsNotExist(err) {
		t.Skip("echo test plugin not built yet")
	}

	binDir := t.TempDir()
	binName := "echo"
	if runtime.GOOS == "windows" {
		binName = "echo.exe"
	}
	binPath := filepath.Join(binDir, binName)

	// Build the echo plugin
	buildEchoPlugin(t, echoDir, binPath)

	result, err := Execute(binPath, []string{"hello"}, nil, "echo", "1.0.0", "/tmp/project")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if !strings.Contains(result.Stdout, "OK") {
		t.Errorf("Stdout should contain OK, got %q", result.Stdout)
	}
}

func TestExecutePluginWithStdin(t *testing.T) {
	echoDir := filepath.Join(getProjectRoot(t), "testdata", "plugins", "echo")
	if _, err := os.Stat(filepath.Join(echoDir, "main.go")); os.IsNotExist(err) {
		t.Skip("echo test plugin not built yet")
	}

	binDir := t.TempDir()
	binName := "echo"
	if runtime.GOOS == "windows" {
		binName = "echo.exe"
	}
	binPath := filepath.Join(binDir, binName)
	buildEchoPlugin(t, echoDir, binPath)

	section := map[string]any{
		"environment": "staging",
		"port":        int64(3000),
	}

	result, err := Execute(binPath, []string{"dev"}, section, "echo", "1.0.0", "/tmp/project")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d", result.ExitCode)
	}
	if !strings.Contains(result.Stdout, "environment=staging") {
		t.Errorf("Stdout should contain config values, got %q", result.Stdout)
	}
}

func TestExecutePluginError(t *testing.T) {
	echoDir := filepath.Join(getProjectRoot(t), "testdata", "plugins", "echo")
	if _, err := os.Stat(filepath.Join(echoDir, "main.go")); os.IsNotExist(err) {
		t.Skip("echo test plugin not built yet")
	}

	binDir := t.TempDir()
	binName := "echo"
	if runtime.GOOS == "windows" {
		binName = "echo.exe"
	}
	binPath := filepath.Join(binDir, binName)
	buildEchoPlugin(t, echoDir, binPath)

	result, err := Execute(binPath, []string{"fail"}, nil, "echo", "1.0.0", "/tmp/project")
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if result.ExitCode == 0 {
		t.Error("ExitCode should be non-zero for fail command")
	}
	if result.PluginError == nil {
		t.Error("PluginError should be parsed from stderr")
	} else {
		if result.PluginError.Code != "ECHO_FAIL" {
			t.Errorf("Error code = %q, want ECHO_FAIL", result.PluginError.Code)
		}
	}
}

func TestExecuteNonexistentBinary(t *testing.T) {
	_, err := Execute("/nonexistent/binary", []string{}, nil, "test", "1.0.0", "/tmp")
	if err == nil {
		t.Error("Should fail for nonexistent binary")
	}
}

// --- helpers ---

func getProjectRoot(t *testing.T) string {
	t.Helper()
	// Walk up from this test file to find go.mod
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root")
		}
		dir = parent
	}
}

func buildEchoPlugin(t *testing.T, srcDir, binPath string) {
	t.Helper()
	// Use go build to compile the echo plugin
	cmd := execCommand("go", "build", "-o", binPath, ".")
	cmd.Dir = srcDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build echo plugin: %v\n%s", err, out)
	}
}
