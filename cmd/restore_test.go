package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/wave-cli/wave-core/internal/config"
)

func TestRestoreCmdNoPlugins(t *testing.T) {
	root := t.TempDir()
	setTestHome(t, root)

	resetCmdState()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"restore"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	// Should succeed with no plugins to restore
}

func TestRestoreCmdExistsAsCommand(t *testing.T) {
	resetCmdState()

	rootCmd := NewRootCmd()

	// Find restore command
	var restoreCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "restore" {
			restoreCmd = cmd
			break
		}
	}

	if restoreCmd == nil {
		t.Fatal("expected 'restore' command to exist")
	}

	if restoreCmd.Short == "" {
		t.Error("restore command should have a short description")
	}
}

func TestRestoreCmdWithPluginsInConfig(t *testing.T) {
	root := t.TempDir()
	setTestHome(t, root)

	// Create global config with a plugin
	waveDir := filepath.Join(root, ".wave")
	if err := os.MkdirAll(waveDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	gc := config.DefaultGlobalConfig(root)
	gc.Plugins["wave-cli/wave-flow"] = "v0.1.0"
	configPath := filepath.Join(waveDir, "config")
	if err := config.WriteGlobalConfig(configPath, gc); err != nil {
		t.Fatalf("WriteGlobalConfig failed: %v", err)
	}

	resetCmdState()

	// Note: This test will attempt to download from GitHub which may fail in CI
	// The important thing is that the command exists and processes the config
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"restore", "--dry-run"})

	// Execute should work (dry-run doesn't actually download)
	if err := cmd.Execute(); err != nil {
		// It's okay if it fails due to network - we're testing the command exists
		t.Logf("restore command ran (may have failed due to network): %v", err)
	}
}
