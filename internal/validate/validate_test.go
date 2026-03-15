package validate

import (
	"strings"
	"testing"
)

// --- ValidatePluginConfig ---

func TestValidatePluginConfigHappyPath(t *testing.T) {
	rawWavefile := []byte(`
[project]
name = "test"
version = "1.0.0"

[flow]
build = { cmd = "go build", on_success = "echo done" }
clean = { cmd = "rm -rf dist" }
`)

	schemaBytes := []byte(`
plugin = "flow"

[fields.cmd]
type = "string"
required = true
desc = "Command to execute"

[fields.on_success]
type = "string"
required = false
desc = "Command on success"

[fields.on_fail]
type = "string"
required = false
desc = "Command on failure"

[fields.env]
type = "map"
required = false
desc = "Environment variables"

[fields.watch]
type = "any"
required = false
desc = "Watch patterns"
`)

	section := map[string]any{
		"build": map[string]any{
			"cmd":        "go build",
			"on_success": "echo done",
		},
		"clean": map[string]any{
			"cmd": "rm -rf dist",
		},
	}

	errs := ValidatePluginConfig("flow", rawWavefile, section, schemaBytes)
	if len(errs) != 0 {
		t.Errorf("Expected no errors for valid config, got %v", errs)
	}
}

func TestValidatePluginConfigSchemaViolation(t *testing.T) {
	rawWavefile := []byte(`
[flow]
build = { on_success = "echo done" }
`)

	schemaBytes := []byte(`
plugin = "flow"

[fields.cmd]
type = "string"
required = true

[fields.on_success]
type = "string"
required = false
`)

	section := map[string]any{
		"build": map[string]any{
			"on_success": "echo done",
		},
	}

	errs := ValidatePluginConfig("flow", rawWavefile, section, schemaBytes)
	if len(errs) == 0 {
		t.Fatal("Expected schema validation error for missing required 'cmd'")
	}

	found := false
	for _, e := range errs {
		if strings.Contains(e, "cmd") {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected error about 'cmd', got %v", errs)
	}
}

func TestValidatePluginConfigRulesViolation(t *testing.T) {
	rawWavefile := []byte(`
[flow.build]
cmd = "go build"
`)

	schemaBytes := []byte(`
plugin = "flow"

[fields.cmd]
type = "string"
required = true
`)

	// Section doesn't matter here — the rules check raw TOML
	section := map[string]any{}

	errs := ValidatePluginConfig("flow", rawWavefile, section, schemaBytes)
	if len(errs) == 0 {
		t.Fatal("Expected rules violation for [flow.build] nested header")
	}

	found := false
	for _, e := range errs {
		if strings.Contains(e, "WaveStructureError") {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected WaveStructureError, got %v", errs)
	}
}

func TestValidatePluginConfigGlobalNamespaceViolation(t *testing.T) {
	rawWavefile := []byte(`
cmd = "leaked"

[flow]
build = { cmd = "go build" }
`)

	schemaBytes := []byte(`
plugin = "flow"

[fields.cmd]
type = "string"
required = true
`)

	section := map[string]any{
		"build": map[string]any{
			"cmd": "go build",
		},
	}

	errs := ValidatePluginConfig("flow", rawWavefile, section, schemaBytes)
	if len(errs) == 0 {
		t.Fatal("Expected GlobalNamespaceError for leaked 'cmd'")
	}

	found := false
	for _, e := range errs {
		if strings.Contains(e, "GlobalNamespaceError") {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected GlobalNamespaceError, got %v", errs)
	}
}

func TestValidatePluginConfigNilSchema(t *testing.T) {
	rawWavefile := []byte(`
[flow]
build = { cmd = "go build" }
`)

	section := map[string]any{
		"build": map[string]any{
			"cmd": "go build",
		},
	}

	// nil schema = plugin has no schema, skip validation
	errs := ValidatePluginConfig("flow", rawWavefile, section, nil)
	if len(errs) != 0 {
		t.Errorf("Expected no errors when schema is nil, got %v", errs)
	}
}

func TestValidatePluginConfigNilSection(t *testing.T) {
	rawWavefile := []byte(`
[project]
name = "test"
`)

	schemaBytes := []byte(`
plugin = "flow"

[fields.cmd]
type = "string"
required = true
`)

	// nil section = plugin section not in Wavefile, skip schema validation
	// but still run rules
	errs := ValidatePluginConfig("flow", rawWavefile, nil, schemaBytes)
	if len(errs) != 0 {
		t.Errorf("Expected no errors when section is nil, got %v", errs)
	}
}

func TestValidatePluginConfigInvalidSchemaBytes(t *testing.T) {
	rawWavefile := []byte(`
[flow]
build = { cmd = "go build" }
`)

	section := map[string]any{
		"build": map[string]any{
			"cmd": "go build",
		},
	}

	errs := ValidatePluginConfig("flow", rawWavefile, section, []byte("[[[broken"))
	if len(errs) == 0 {
		t.Fatal("Expected error for invalid schema bytes")
	}
}

func TestValidatePluginConfigBothRulesAndSchemaErrors(t *testing.T) {
	rawWavefile := []byte(`
cmd = "leaked"

[flow]
build = { on_success = "echo done" }
`)

	schemaBytes := []byte(`
plugin = "flow"

[fields.cmd]
type = "string"
required = true

[fields.on_success]
type = "string"
required = false
`)

	section := map[string]any{
		"build": map[string]any{
			"on_success": "echo done",
		},
	}

	errs := ValidatePluginConfig("flow", rawWavefile, section, schemaBytes)
	// Should have both: GlobalNamespaceError + missing required 'cmd'
	if len(errs) < 2 {
		t.Fatalf("Expected at least 2 errors, got %d: %v", len(errs), errs)
	}
}

func TestValidatePluginConfigEmptySchemaBytes(t *testing.T) {
	rawWavefile := []byte(`
[flow]
build = { cmd = "go build" }
`)

	section := map[string]any{
		"build": map[string]any{
			"cmd": "go build",
		},
	}

	errs := ValidatePluginConfig("flow", rawWavefile, section, []byte(""))
	// Empty schema should be treated like nil (skip validation) —
	// but ParseSchemaFromBytes returns error for empty, so we get an error
	// Actually, empty bytes → skip like nil
	if len(errs) != 0 {
		t.Errorf("Expected no errors for empty schema bytes, got %v", errs)
	}
}
