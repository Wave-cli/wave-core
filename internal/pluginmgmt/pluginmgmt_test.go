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

	// Create: plugins/wave-cli/flow/v1.2.0/bin/flow
	binDir := filepath.Join(pluginsDir, "wave-cli", "flow", "v1.2.0", "bin")
	os.MkdirAll(binDir, 0755)
	binPath := filepath.Join(binDir, "flow")
	os.WriteFile(binPath, []byte("#!/bin/sh\necho ok"), 0755)

	// Create current symlink
	versionDir := filepath.Join(pluginsDir, "wave-cli", "flow", "v1.2.0")
	currentLink := filepath.Join(pluginsDir, "wave-cli", "flow", "current")
	os.Symlink(versionDir, currentLink)

	reg := NewRegistry(pluginsDir)
	resolved, err := reg.ResolveBinary("wave-cli/flow")
	if err != nil {
		t.Fatalf("ResolveBinary failed: %v", err)
	}

	expected := filepath.Join(pluginsDir, "wave-cli", "flow", "current", "bin", "flow")
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

	// Create: plugins/wave-cli/flow/v1.0.0/assets/
	assetsDir := filepath.Join(pluginsDir, "wave-cli", "flow", "v1.0.0", "assets")
	os.MkdirAll(assetsDir, 0755)
	versionDir := filepath.Join(pluginsDir, "wave-cli", "flow", "v1.0.0")
	currentLink := filepath.Join(pluginsDir, "wave-cli", "flow", "current")
	os.Symlink(versionDir, currentLink)

	reg := NewRegistry(pluginsDir)
	resolved, err := reg.ResolveAssets("wave-cli/flow")
	if err != nil {
		t.Fatalf("ResolveAssets failed: %v", err)
	}

	expected := filepath.Join(pluginsDir, "wave-cli", "flow", "current", "assets")
	if resolved != expected {
		t.Errorf("resolved = %q, want %q", resolved, expected)
	}
}

func TestResolveWavepluginPath(t *testing.T) {
	home := t.TempDir()
	pluginsDir := filepath.Join(home, ".wave", "plugins")

	// Create: plugins/wave-cli/flow/v1.0.0/Waveplugin
	versionDir := filepath.Join(pluginsDir, "wave-cli", "flow", "v1.0.0")
	os.MkdirAll(versionDir, 0755)
	wpPath := filepath.Join(versionDir, "Waveplugin")
	os.WriteFile(wpPath, []byte("[plugin]\nname=\"flow\"\nversion=\"1.0.0\"\ndescription=\"test\"\ncreator=\"x\"\n"), 0644)

	currentLink := filepath.Join(pluginsDir, "wave-cli", "flow", "current")
	os.Symlink(versionDir, currentLink)

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

// --- Waveplugin schema field ---

func TestParseWavepluginWithSchema(t *testing.T) {
	content := `
[plugin]
name = "flow"
version = "1.0.0"
description = "Flow plugin"
creator = "wave-cli"

[compatibility]
min_wave_version = "0.1.0"

[assets]
schema = "Waveschema"
files = ["templates/"]
`
	dir := t.TempDir()
	path := filepath.Join(dir, "Waveplugin")
	os.WriteFile(path, []byte(content), 0644)

	wp, err := ParseWaveplugin(path)
	if err != nil {
		t.Fatalf("ParseWaveplugin failed: %v", err)
	}
	if wp.Assets.Schema != "Waveschema" {
		t.Errorf("Schema = %q, want %q", wp.Assets.Schema, "Waveschema")
	}
}

func TestParseWavepluginWithoutSchema(t *testing.T) {
	content := `
[plugin]
name = "bare"
version = "0.1.0"
description = "No schema plugin"
creator = "test"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "Waveplugin")
	os.WriteFile(path, []byte(content), 0644)

	wp, err := ParseWaveplugin(path)
	if err != nil {
		t.Fatalf("ParseWaveplugin failed: %v", err)
	}
	if wp.Assets.Schema != "" {
		t.Errorf("Schema should be empty for plugins without schema, got %q", wp.Assets.Schema)
	}
}

// --- Registry.ResolveSchema ---

func TestResolveSchemaPath(t *testing.T) {
	home := t.TempDir()
	pluginsDir := filepath.Join(home, ".wave", "plugins")

	// Create plugin directory with Waveplugin and schema file
	versionDir := filepath.Join(pluginsDir, "wave-cli", "flow", "v1.0.0")
	os.MkdirAll(versionDir, 0755)

	// Write Waveplugin with schema reference
	wpContent := `
[plugin]
name = "flow"
version = "1.0.0"
description = "Flow plugin"
creator = "wave-cli"

[assets]
schema = "Waveschema"
`
	os.WriteFile(filepath.Join(versionDir, "Waveplugin"), []byte(wpContent), 0644)

	// Write the schema file
	schemaContent := `
plugin = "flow"

[fields.cmd]
type = "string"
required = true
desc = "Command to execute"
`
	os.WriteFile(filepath.Join(versionDir, "Waveschema"), []byte(schemaContent), 0644)

	// Create current symlink
	currentLink := filepath.Join(pluginsDir, "wave-cli", "flow", "current")
	os.Symlink(versionDir, currentLink)

	reg := NewRegistry(pluginsDir)
	schemaBytes, err := reg.ReadSchema("wave-cli/flow")
	if err != nil {
		t.Fatalf("ReadSchema failed: %v", err)
	}
	if len(schemaBytes) == 0 {
		t.Fatal("ReadSchema returned empty bytes")
	}
}

func TestResolveSchemaNoSchemaField(t *testing.T) {
	home := t.TempDir()
	pluginsDir := filepath.Join(home, ".wave", "plugins")

	versionDir := filepath.Join(pluginsDir, "wave-cli", "bare", "v1.0.0")
	os.MkdirAll(versionDir, 0755)

	// Write Waveplugin WITHOUT schema field
	wpContent := `
[plugin]
name = "bare"
version = "1.0.0"
description = "No schema"
creator = "test"
`
	os.WriteFile(filepath.Join(versionDir, "Waveplugin"), []byte(wpContent), 0644)

	currentLink := filepath.Join(pluginsDir, "wave-cli", "bare", "current")
	os.Symlink(versionDir, currentLink)

	reg := NewRegistry(pluginsDir)
	schemaBytes, err := reg.ReadSchema("wave-cli/bare")
	if err != nil {
		t.Fatalf("ReadSchema should not error for plugins without schema: %v", err)
	}
	if schemaBytes != nil {
		t.Errorf("ReadSchema should return nil for plugins without schema, got %d bytes", len(schemaBytes))
	}
}

func TestResolveSchemaNotInstalled(t *testing.T) {
	home := t.TempDir()
	pluginsDir := filepath.Join(home, ".wave", "plugins")
	os.MkdirAll(pluginsDir, 0755)

	reg := NewRegistry(pluginsDir)
	_, err := reg.ReadSchema("wave-cli/nonexistent")
	if err == nil {
		t.Error("ReadSchema should fail for non-installed plugin")
	}
}
