package pluginmgmt

import (
	"os"
	"path/filepath"
	"testing"
)

// --- Waveplugin metadata parsing ---

func TestParseWaveplugin(t *testing.T) {
	content := `
[plugin]
name = "flow"
version = "1.2.0"
description = "Development workflow automation"
creator = "wave-cli"
license = "MIT"
homepage = "https://github.com/wave-cli/flow"

[compatibility]
min_wave_version = "0.1.0"

[assets]
files = ["templates/", "defaults.toml"]
`
	dir := t.TempDir()
	path := filepath.Join(dir, "Waveplugin")
	os.WriteFile(path, []byte(content), 0644)

	wp, err := ParseWaveplugin(path)
	if err != nil {
		t.Fatalf("ParseWaveplugin failed: %v", err)
	}
	if wp.Plugin.Name != "flow" {
		t.Errorf("Name = %q", wp.Plugin.Name)
	}
	if wp.Plugin.Version != "1.2.0" {
		t.Errorf("Version = %q", wp.Plugin.Version)
	}
	if wp.Plugin.Description != "Development workflow automation" {
		t.Errorf("Description = %q", wp.Plugin.Description)
	}
	if wp.Plugin.Creator != "wave-cli" {
		t.Errorf("Creator = %q", wp.Plugin.Creator)
	}
	if wp.Plugin.License != "MIT" {
		t.Errorf("License = %q", wp.Plugin.License)
	}
	if wp.Plugin.Homepage != "https://github.com/wave-cli/flow" {
		t.Errorf("Homepage = %q", wp.Plugin.Homepage)
	}
	if wp.Compatibility.MinWaveVersion != "0.1.0" {
		t.Errorf("MinWaveVersion = %q", wp.Compatibility.MinWaveVersion)
	}
	if len(wp.Assets.Files) != 2 {
		t.Fatalf("Assets.Files len = %d, want 2", len(wp.Assets.Files))
	}
}

func TestParseWavepluginMinimal(t *testing.T) {
	content := `
[plugin]
name = "bare"
version = "0.1.0"
description = "Minimal plugin"
creator = "test"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "Waveplugin")
	os.WriteFile(path, []byte(content), 0644)

	wp, err := ParseWaveplugin(path)
	if err != nil {
		t.Fatalf("ParseWaveplugin failed: %v", err)
	}
	if wp.Plugin.Name != "bare" {
		t.Errorf("Name = %q", wp.Plugin.Name)
	}
	if wp.Assets.Files == nil {
		t.Error("Assets.Files should be initialized even if empty")
	}
}

func TestParseWavepluginMissing(t *testing.T) {
	_, err := ParseWaveplugin("/nonexistent/Waveplugin")
	if err == nil {
		t.Error("Should fail for missing file")
	}
}

// --- Name parsing ---

func TestParsePluginRef(t *testing.T) {
	tests := []struct {
		input   string
		org     string
		name    string
		version string
		err     bool
	}{
		{"wave-cli/flow", "wave-cli", "flow", "", false},
		{"wave-cli/flow@1.2.0", "wave-cli", "flow", "1.2.0", false},
		{"my-org/my-plugin@0.1.0-beta", "my-org", "my-plugin", "0.1.0-beta", false},
		{"flow", "", "", "", true},      // no org
		{"", "", "", "", true},          // empty
		{"a/b/c", "", "", "", true},     // too many slashes
		{"wave-cli/", "", "", "", true}, // empty name
		{"/flow", "", "", "", true},     // empty org
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ref, err := ParsePluginRef(tt.input)
			if tt.err {
				if err == nil {
					t.Errorf("Expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePluginRef(%q) failed: %v", tt.input, err)
			}
			if ref.Org != tt.org {
				t.Errorf("Org = %q, want %q", ref.Org, tt.org)
			}
			if ref.Name != tt.name {
				t.Errorf("Name = %q, want %q", ref.Name, tt.name)
			}
			if ref.Version != tt.version {
				t.Errorf("Version = %q, want %q", ref.Version, tt.version)
			}
		})
	}
}

func TestPluginRefFullName(t *testing.T) {
	ref := PluginRef{Org: "wave-cli", Name: "flow", Version: "1.0.0"}
	if ref.FullName() != "wave-cli/flow" {
		t.Errorf("FullName = %q", ref.FullName())
	}
}

// --- Registry (resolve installed plugin to binary path) ---

func TestResolveBinaryPath(t *testing.T) {
	// Set up a fake plugin store
	home := t.TempDir()
	pluginsDir := filepath.Join(home, ".wave", "plugins")

	// Create: plugins/flow/bin/flow (no org folder)
	binDir := filepath.Join(pluginsDir, "flow", "bin")
	os.MkdirAll(binDir, 0755)
	binPath := filepath.Join(binDir, "flow")
	os.WriteFile(binPath, []byte("#!/bin/sh\necho ok"), 0755)

	reg := NewRegistry(pluginsDir)
	resolved, err := reg.ResolveBinary("wave-cli/flow")
	if err != nil {
		t.Fatalf("ResolveBinary failed: %v", err)
	}

	expected := filepath.Join(pluginsDir, "flow", "bin", "flow")
	if resolved != expected {
		t.Errorf("resolved = %q, want %q", resolved, expected)
	}
}

func TestResolveBinaryNotInstalled(t *testing.T) {
	home := t.TempDir()
	pluginsDir := filepath.Join(home, ".wave", "plugins")
	os.MkdirAll(pluginsDir, 0755)

	reg := NewRegistry(pluginsDir)
	_, err := reg.ResolveBinary("wave-cli/nonexistent")
	if err == nil {
		t.Error("Should fail for non-installed plugin")
	}
}

func TestResolveAssetsPath(t *testing.T) {
	home := t.TempDir()
	pluginsDir := filepath.Join(home, ".wave", "plugins")

	// Create: plugins/flow/assets/ (no org folder)
	assetsDir := filepath.Join(pluginsDir, "flow", "assets")
	os.MkdirAll(assetsDir, 0755)

	reg := NewRegistry(pluginsDir)
	resolved, err := reg.ResolveAssets("wave-cli/flow")
	if err != nil {
		t.Fatalf("ResolveAssets failed: %v", err)
	}

	expected := filepath.Join(pluginsDir, "flow", "assets")
	if resolved != expected {
		t.Errorf("resolved = %q, want %q", resolved, expected)
	}
}

func TestResolveWavepluginPath(t *testing.T) {
	home := t.TempDir()
	pluginsDir := filepath.Join(home, ".wave", "plugins")

	// Create: plugins/flow/Waveplugin (no org folder)
	pluginDir := filepath.Join(pluginsDir, "flow")
	os.MkdirAll(pluginDir, 0755)
	wpPath := filepath.Join(pluginDir, "Waveplugin")
	os.WriteFile(wpPath, []byte("[plugin]\nname=\"flow\"\nversion=\"1.0.0\"\ndescription=\"test\"\ncreator=\"x\"\n"), 0644)

	reg := NewRegistry(pluginsDir)
	wp, err := reg.ReadWaveplugin("wave-cli/flow")
	if err != nil {
		t.Fatalf("ReadWaveplugin failed: %v", err)
	}
	if wp.Plugin.Name != "flow" {
		t.Errorf("Name = %q", wp.Plugin.Name)
	}
}

// --- List installed plugins ---

func TestListInstalled(t *testing.T) {
	home := t.TempDir()
	pluginsDir := filepath.Join(home, ".wave", "plugins")

	plugins := map[string]string{
		"wave-cli/flow": "1.2.0",
		"wave-cli/test": "0.5.3",
	}

	reg := NewRegistry(pluginsDir)
	list := reg.ListInstalled(plugins)

	if len(list) != 2 {
		t.Fatalf("ListInstalled returned %d, want 2", len(list))
	}
}

func TestListInstalledEmpty(t *testing.T) {
	home := t.TempDir()
	pluginsDir := filepath.Join(home, ".wave", "plugins")
	os.MkdirAll(pluginsDir, 0755)

	reg := NewRegistry(pluginsDir)
	list := reg.ListInstalled(map[string]string{})
	if len(list) != 0 {
		t.Errorf("ListInstalled should return empty for no plugins, got %d", len(list))
	}
}
