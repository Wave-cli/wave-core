package bootstrap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wave-cli/wave-core/internal/config"
)

func TestEnsureCreatesWaveDir(t *testing.T) {
	home := t.TempDir()
	waveDir := filepath.Join(home, ".wave")

	gc, err := Ensure(home)
	if err != nil {
		t.Fatalf("Ensure failed: %v", err)
	}

	// .wave/ should exist
	if _, err := os.Stat(waveDir); os.IsNotExist(err) {
		t.Error("~/.wave/ should be created")
	}

	// Config should not be nil
	if gc == nil {
		t.Fatal("returned config should not be nil")
	}
}

func TestEnsureCreatesPluginsDir(t *testing.T) {
	home := t.TempDir()
	pluginsDir := filepath.Join(home, ".wave", "plugins")

	Ensure(home)

	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		t.Error("~/.wave/plugins/ should be created")
	}
}

func TestEnsureCreatesLogsDir(t *testing.T) {
	home := t.TempDir()

	gc, _ := Ensure(home)

	// Logs dir from returned config should exist
	if _, err := os.Stat(gc.Core.LogsDir); os.IsNotExist(err) {
		t.Errorf("logs dir %q should be created", gc.Core.LogsDir)
	}
}

func TestEnsureCreatesConfigFile(t *testing.T) {
	home := t.TempDir()
	configPath := filepath.Join(home, ".wave", "config")

	Ensure(home)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("~/.wave/config should be created")
	}
}

func TestEnsureConfigIsValidTOML(t *testing.T) {
	home := t.TempDir()
	configPath := filepath.Join(home, ".wave", "config")

	Ensure(home)

	// Should be parseable
	gc, err := config.ParseGlobalConfig(configPath)
	if err != nil {
		t.Fatalf("Generated config is not valid TOML: %v", err)
	}
	if gc.Core.LogsDir == "" {
		t.Error("Default config should have logs_dir set")
	}
	if gc.Plugins == nil {
		t.Error("Default config should have plugins map initialized")
	}
}

func TestEnsureIdempotent(t *testing.T) {
	home := t.TempDir()

	// Run twice — should not fail or overwrite
	gc1, err := Ensure(home)
	if err != nil {
		t.Fatalf("First Ensure failed: %v", err)
	}

	gc2, err := Ensure(home)
	if err != nil {
		t.Fatalf("Second Ensure failed: %v", err)
	}

	// Config should be the same
	if gc1.Core.LogsDir != gc2.Core.LogsDir {
		t.Errorf("Config changed between runs: %q vs %q", gc1.Core.LogsDir, gc2.Core.LogsDir)
	}
}

func TestEnsurePreservesExistingConfig(t *testing.T) {
	home := t.TempDir()
	waveDir := filepath.Join(home, ".wave")
	os.MkdirAll(waveDir, 0755)

	// Write a custom config first
	customLogsDir := filepath.Join(home, "custom-logs")
	customConfig := &config.GlobalConfig{
		Core: config.CoreConfig{LogsDir: customLogsDir},
		Projects: config.ProjectsConfig{
			Folders: []string{"/my/projects"},
		},
		Plugins: map[string]string{
			"wave-cli/flow": "9.9.9",
		},
	}
	configPath := filepath.Join(waveDir, "config")
	config.WriteGlobalConfig(configPath, customConfig)

	// Ensure should NOT overwrite
	gc, err := Ensure(home)
	if err != nil {
		t.Fatalf("Ensure failed: %v", err)
	}
	if gc.Core.LogsDir != customLogsDir {
		t.Errorf("LogsDir was overwritten: got %q, want %q", gc.Core.LogsDir, customLogsDir)
	}
	if gc.Plugins["wave-cli/flow"] != "9.9.9" {
		t.Errorf("Plugin version was overwritten: got %q", gc.Plugins["wave-cli/flow"])
	}
}
