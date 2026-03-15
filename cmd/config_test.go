package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/wave-cli/wave-core/internal/config"
)

func TestConfigSetUpdatesUserFields(t *testing.T) {
	root := t.TempDir()
	setTestHome(t, root)

	resetCmdState()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"config", "set", "--name", "Ada", "--org", "Wave"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	configPath := filepath.Join(root, ".wave", "config")
	gc, err := config.ParseGlobalConfig(configPath)
	if err != nil {
		t.Fatalf("ParseGlobalConfig failed: %v", err)
	}

	if gc.User.Name != "Ada" {
		t.Fatalf("User.Name = %q, want %q", gc.User.Name, "Ada")
	}
	if gc.User.Org != "Wave" {
		t.Fatalf("User.Org = %q, want %q", gc.User.Org, "Wave")
	}
}

func TestConfigSetRequiresAtLeastOneField(t *testing.T) {
	root := t.TempDir()
	setTestHome(t, root)

	resetCmdState()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"config", "set"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("Expected error when no fields provided")
	}
}

func setTestHome(t *testing.T, home string) {
	t.Helper()
	if err := os.Setenv("HOME", home); err != nil {
		t.Fatalf("Setenv HOME failed: %v", err)
	}
	if err := os.Setenv("USERPROFILE", home); err != nil {
		t.Fatalf("Setenv USERPROFILE failed: %v", err)
	}
}

func resetCmdState() {
	printer = nil
	globalCfg = nil
	viper.Reset()
}
