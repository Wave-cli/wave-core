package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/wave-cli/wave-core/internal/config"
)

func TestConfigSetUserUpdatesUserName(t *testing.T) {
	root := t.TempDir()
	setTestHome(t, root)

	resetCmdState()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"config", "set-user", "Ada"})
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
}

func TestConfigSetOrgUpdatesUserOrg(t *testing.T) {
	root := t.TempDir()
	setTestHome(t, root)

	resetCmdState()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"config", "set-org", "Wave"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	configPath := filepath.Join(root, ".wave", "config")
	gc, err := config.ParseGlobalConfig(configPath)
	if err != nil {
		t.Fatalf("ParseGlobalConfig failed: %v", err)
	}

	if gc.User.Org != "Wave" {
		t.Fatalf("User.Org = %q, want %q", gc.User.Org, "Wave")
	}
}

func TestConfigSetUserRequiresArg(t *testing.T) {
	root := t.TempDir()
	setTestHome(t, root)

	resetCmdState()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"config", "set-user"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("Expected error when no username provided")
	}
}

func TestConfigSetOrgRequiresArg(t *testing.T) {
	root := t.TempDir()
	setTestHome(t, root)

	resetCmdState()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"config", "set-org"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("Expected error when no org provided")
	}
}

func TestConfigSetUserPreservesOrg(t *testing.T) {
	root := t.TempDir()
	setTestHome(t, root)

	// Create initial config with org set
	waveDir := filepath.Join(root, ".wave")
	if err := os.MkdirAll(waveDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	gc := config.DefaultGlobalConfig(root)
	gc.User.Org = "ExistingOrg"
	configPath := filepath.Join(waveDir, "config")
	if err := config.WriteGlobalConfig(configPath, gc); err != nil {
		t.Fatalf("WriteGlobalConfig failed: %v", err)
	}

	resetCmdState()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"config", "set-user", "NewUser"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	gc, err := config.ParseGlobalConfig(configPath)
	if err != nil {
		t.Fatalf("ParseGlobalConfig failed: %v", err)
	}

	if gc.User.Name != "NewUser" {
		t.Fatalf("User.Name = %q, want %q", gc.User.Name, "NewUser")
	}
	if gc.User.Org != "ExistingOrg" {
		t.Fatalf("User.Org = %q, want %q (should be preserved)", gc.User.Org, "ExistingOrg")
	}
}

func TestConfigSetOrgPreservesName(t *testing.T) {
	root := t.TempDir()
	setTestHome(t, root)

	// Create initial config with name set
	waveDir := filepath.Join(root, ".wave")
	if err := os.MkdirAll(waveDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	gc := config.DefaultGlobalConfig(root)
	gc.User.Name = "ExistingUser"
	configPath := filepath.Join(waveDir, "config")
	if err := config.WriteGlobalConfig(configPath, gc); err != nil {
		t.Fatalf("WriteGlobalConfig failed: %v", err)
	}

	resetCmdState()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"config", "set-org", "NewOrg"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	gc, err := config.ParseGlobalConfig(configPath)
	if err != nil {
		t.Fatalf("ParseGlobalConfig failed: %v", err)
	}

	if gc.User.Org != "NewOrg" {
		t.Fatalf("User.Org = %q, want %q", gc.User.Org, "NewOrg")
	}
	if gc.User.Name != "ExistingUser" {
		t.Fatalf("User.Name = %q, want %q (should be preserved)", gc.User.Name, "ExistingUser")
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
