package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGlobalConfig(t *testing.T) {
	content := `
[core]
logs_dir = "/tmp/wave-logs"

[projects]
folders = ["/home/user/projects", "/home/user/work"]

[plugins]
"wave-cli/flow" = "1.2.0"
"wave-cli/test" = "0.5.3"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	gc, err := ParseGlobalConfig(path)
	if err != nil {
		t.Fatalf("ParseGlobalConfig failed: %v", err)
	}

	if gc.Core.LogsDir != "/tmp/wave-logs" {
		t.Errorf("LogsDir = %q, want /tmp/wave-logs", gc.Core.LogsDir)
	}
	if len(gc.Projects.Folders) != 2 {
		t.Fatalf("Folders len = %d, want 2", len(gc.Projects.Folders))
	}
	if gc.Projects.Folders[0] != "/home/user/projects" {
		t.Errorf("Folders[0] = %q", gc.Projects.Folders[0])
	}
	if gc.Plugins["wave-cli/flow"] != "1.2.0" {
		t.Errorf("Plugin flow version = %q, want 1.2.0", gc.Plugins["wave-cli/flow"])
	}
	if gc.Plugins["wave-cli/test"] != "0.5.3" {
		t.Errorf("Plugin test version = %q, want 0.5.3", gc.Plugins["wave-cli/test"])
	}
}

func TestParseGlobalConfigMinimal(t *testing.T) {
	// Minimal config with only [core]
	content := `
[core]
logs_dir = "/tmp/logs"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	os.WriteFile(path, []byte(content), 0644)

	gc, err := ParseGlobalConfig(path)
	if err != nil {
		t.Fatalf("ParseGlobalConfig failed: %v", err)
	}
	if gc.Plugins == nil {
		t.Error("Plugins map should be initialized even if empty")
	}
	if len(gc.Projects.Folders) != 0 {
		t.Errorf("Folders should be empty, got %d", len(gc.Projects.Folders))
	}
}

func TestParseGlobalConfigMissing(t *testing.T) {
	_, err := ParseGlobalConfig("/nonexistent/path/config")
	if err == nil {
		t.Error("ParseGlobalConfig should fail for missing file")
	}
}

func TestParseGlobalConfigInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	os.WriteFile(path, []byte("this is not valid toml [[["), 0644)

	_, err := ParseGlobalConfig(path)
	if err == nil {
		t.Error("ParseGlobalConfig should fail for invalid TOML")
	}
}

func TestWriteGlobalConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")

	gc := &GlobalConfig{
		Core: CoreConfig{LogsDir: "/tmp/wave-logs"},
		Projects: ProjectsConfig{
			Folders: []string{"/home/user/work"},
		},
		Plugins: map[string]string{
			"wave-cli/flow": "1.0.0",
		},
	}

	if err := WriteGlobalConfig(path, gc); err != nil {
		t.Fatalf("WriteGlobalConfig failed: %v", err)
	}

	// Read it back
	gc2, err := ParseGlobalConfig(path)
	if err != nil {
		t.Fatalf("Re-read failed: %v", err)
	}
	if gc2.Core.LogsDir != "/tmp/wave-logs" {
		t.Errorf("LogsDir roundtrip failed: %q", gc2.Core.LogsDir)
	}
	if gc2.Plugins["wave-cli/flow"] != "1.0.0" {
		t.Errorf("Plugin roundtrip failed: %q", gc2.Plugins["wave-cli/flow"])
	}
}

func TestDefaultGlobalConfig(t *testing.T) {
	gc := DefaultGlobalConfig("/home/testuser")
	if gc.Core.LogsDir == "" {
		t.Error("Default LogsDir should not be empty")
	}
	if gc.Plugins == nil {
		t.Error("Default Plugins should be initialized")
	}
}
