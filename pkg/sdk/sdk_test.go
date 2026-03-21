package sdk

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// =============================================================================
// Config Reading
// =============================================================================

func TestReadConfig(t *testing.T) {
	input := `{"build":{"cmd":"go build","on_success":"echo done"},"clean":{"cmd":"rm -rf dist"}}`
	r := strings.NewReader(input)

	cfg, err := ReadConfigFrom(r)
	if err != nil {
		t.Fatalf("ReadConfigFrom failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("Config should not be nil")
	}
	if cfg.Raw() == nil {
		t.Fatal("Raw should not be nil")
	}
}

func TestReadConfigEmpty(t *testing.T) {
	r := strings.NewReader("{}")
	cfg, err := ReadConfigFrom(r)
	if err != nil {
		t.Fatalf("ReadConfigFrom failed: %v", err)
	}
	if len(cfg.Raw()) != 0 {
		t.Errorf("Expected empty config, got %v", cfg.Raw())
	}
}

func TestReadConfigInvalidJSON(t *testing.T) {
	r := strings.NewReader("not json")
	_, err := ReadConfigFrom(r)
	if err == nil {
		t.Error("Should fail for invalid JSON")
	}
}

func TestReadConfigEmptyInput(t *testing.T) {
	r := strings.NewReader("")
	_, err := ReadConfigFrom(r)
	if err == nil {
		t.Error("Should fail for empty input")
	}
}

// =============================================================================
// Config Typed Access
// =============================================================================

func TestConfigGetString(t *testing.T) {
	cfg := NewConfig(map[string]any{"name": "hello"})
	if v, ok := cfg.String("name"); !ok || v != "hello" {
		t.Errorf("String('name') = %q, %v", v, ok)
	}
	if _, ok := cfg.String("missing"); ok {
		t.Error("String('missing') should return false")
	}
}

func TestConfigGetMap(t *testing.T) {
	inner := map[string]any{"cmd": "go build", "on_success": "echo done"}
	cfg := NewConfig(map[string]any{"build": inner})

	m, ok := cfg.Map("build")
	if !ok {
		t.Fatal("Map('build') should return true")
	}
	if m["cmd"] != "go build" {
		t.Errorf("Map('build')['cmd'] = %q", m["cmd"])
	}
}

func TestConfigGetMapMissing(t *testing.T) {
	cfg := NewConfig(map[string]any{})
	_, ok := cfg.Map("missing")
	if ok {
		t.Error("Map('missing') should return false")
	}
}

func TestConfigGetBool(t *testing.T) {
	cfg := NewConfig(map[string]any{"debug": true})
	if v, ok := cfg.Bool("debug"); !ok || !v {
		t.Errorf("Bool('debug') = %v, %v", v, ok)
	}
}

func TestConfigGetFloat(t *testing.T) {
	cfg := NewConfig(map[string]any{"port": float64(8080)})
	if v, ok := cfg.Float("port"); !ok || v != 8080 {
		t.Errorf("Float('port') = %v, %v", v, ok)
	}
}

func TestConfigKeys(t *testing.T) {
	cfg := NewConfig(map[string]any{
		"build": map[string]any{"cmd": "go build"},
		"clean": map[string]any{"cmd": "rm -rf dist"},
	})
	keys := cfg.Keys()
	if len(keys) != 2 {
		t.Fatalf("Keys() len = %d, want 2", len(keys))
	}
}

func TestConfigHas(t *testing.T) {
	cfg := NewConfig(map[string]any{"build": "x"})
	if !cfg.Has("build") {
		t.Error("Has('build') should be true")
	}
	if cfg.Has("missing") {
		t.Error("Has('missing') should be false")
	}
}

func TestConfigGet(t *testing.T) {
	cfg := NewConfig(map[string]any{"x": int64(42)})
	v, ok := cfg.Get("x")
	if !ok {
		t.Fatal("Get('x') should be true")
	}
	if v != int64(42) {
		t.Errorf("Get('x') = %v", v)
	}
}

// =============================================================================
// Error Emission — small-letters-and-dashes format
// =============================================================================

func TestFormatWaveError(t *testing.T) {
	var buf bytes.Buffer
	FormatWaveError(&buf, "config-parse-error", "failed to parse config", "check TOML syntax")

	var pe struct {
		WaveError bool   `json:"wave_error"`
		Code      string `json:"code"`
		Message   string `json:"message"`
		Details   string `json:"details"`
	}

	if err := json.Unmarshal(buf.Bytes(), &pe); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nGot: %s", err, buf.String())
	}
	if !pe.WaveError {
		t.Error("wave_error should be true")
	}
	if pe.Code != "config-parse-error" {
		t.Errorf("code = %q, want 'config-parse-error'", pe.Code)
	}
	if pe.Message != "failed to parse config" {
		t.Errorf("message = %q", pe.Message)
	}
	if pe.Details != "check TOML syntax" {
		t.Errorf("details = %q", pe.Details)
	}
}

func TestFormatWaveErrorNoDetails(t *testing.T) {
	var buf bytes.Buffer
	FormatWaveError(&buf, "not-found", "resource not found", "")

	var pe map[string]any
	json.Unmarshal(buf.Bytes(), &pe)
	if d, ok := pe["details"]; ok && d != "" {
		t.Errorf("details should be empty, got %v", d)
	}
}

func TestFormatWaveErrorCodeFormat(t *testing.T) {
	// Error codes should be lowercase with dashes
	var buf bytes.Buffer
	FormatWaveError(&buf, "flow-no-command", "no command specified", "")

	var pe struct {
		Code string `json:"code"`
	}
	json.Unmarshal(buf.Bytes(), &pe)
	if pe.Code != "flow-no-command" {
		t.Errorf("code = %q, want 'flow-no-command'", pe.Code)
	}
	// Verify it only contains lowercase letters, digits and dashes
	for _, c := range pe.Code {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			t.Errorf("code contains invalid character %q", string(c))
		}
	}
}

// =============================================================================
// GetPluginEnv
// =============================================================================

func TestGetPluginEnv(t *testing.T) {
	os.Setenv("WAVE_PLUGIN_NAME", "flow")
	os.Setenv("WAVE_PLUGIN_VERSION", "1.2.0")
	os.Setenv("WAVE_PLUGIN_DIR", "/home/.wave/plugins/flow")
	os.Setenv("WAVE_PLUGIN_ASSETS", "/home/.wave/plugins/flow/assets")
	os.Setenv("WAVE_PROJECT_ROOT", "/projects/my-app")
	defer func() {
		os.Unsetenv("WAVE_PLUGIN_NAME")
		os.Unsetenv("WAVE_PLUGIN_VERSION")
		os.Unsetenv("WAVE_PLUGIN_DIR")
		os.Unsetenv("WAVE_PLUGIN_ASSETS")
		os.Unsetenv("WAVE_PROJECT_ROOT")
	}()

	env := GetPluginEnv()
	if env.Name != "flow" {
		t.Errorf("Name = %q", env.Name)
	}
	if env.Version != "1.2.0" {
		t.Errorf("Version = %q", env.Version)
	}
	if env.Dir != "/home/.wave/plugins/flow" {
		t.Errorf("Dir = %q", env.Dir)
	}
	if env.Assets != "/home/.wave/plugins/flow/assets" {
		t.Errorf("Assets = %q", env.Assets)
	}
	if env.ProjectRoot != "/projects/my-app" {
		t.Errorf("ProjectRoot = %q", env.ProjectRoot)
	}
}

// =============================================================================
// Init — full plugin initialization from stdin + env
// =============================================================================

func TestInitFrom(t *testing.T) {
	os.Setenv("WAVE_PLUGIN_NAME", "flow")
	os.Setenv("WAVE_PLUGIN_VERSION", "0.1.0")
	os.Setenv("WAVE_PLUGIN_DIR", "/tmp/plugins/flow")
	os.Setenv("WAVE_PLUGIN_ASSETS", "/tmp/plugins/flow/assets")
	os.Setenv("WAVE_PROJECT_ROOT", "/tmp/project")
	defer func() {
		os.Unsetenv("WAVE_PLUGIN_NAME")
		os.Unsetenv("WAVE_PLUGIN_VERSION")
		os.Unsetenv("WAVE_PLUGIN_DIR")
		os.Unsetenv("WAVE_PLUGIN_ASSETS")
		os.Unsetenv("WAVE_PROJECT_ROOT")
	}()

	input := `{"build":{"cmd":"go build"},"clean":{"cmd":"rm -rf dist"}}`
	r := strings.NewReader(input)

	p, err := InitFrom(r)
	if err != nil {
		t.Fatalf("InitFrom failed: %v", err)
	}
	if p.Env.Name != "flow" {
		t.Errorf("Env.Name = %q", p.Env.Name)
	}
	if p.Config == nil {
		t.Fatal("Config should not be nil")
	}
	if !p.Config.Has("build") {
		t.Error("Config should have 'build' key")
	}
}

func TestInitFromInvalidJSON(t *testing.T) {
	r := strings.NewReader("bad json")
	_, err := InitFrom(r)
	if err == nil {
		t.Error("InitFrom should fail for invalid JSON")
	}
}

// =============================================================================
// Errf — error creation helper (replaces fmt.Errorf)
// =============================================================================

func TestErrf(t *testing.T) {
	err := Errf("something went wrong")
	if err == nil {
		t.Fatal("Errf should return an error")
	}
	if err.Error() != "something went wrong" {
		t.Errorf("Error message = %q", err.Error())
	}
}

func TestErrfFormatted(t *testing.T) {
	err := Errf("failed to open file %q: %s", "test.txt", "not found")
	if err == nil {
		t.Fatal("Errf should return an error")
	}
	expected := `failed to open file "test.txt": not found`
	if err.Error() != expected {
		t.Errorf("Error message = %q, want %q", err.Error(), expected)
	}
}

func TestErrfWrap(t *testing.T) {
	original := Errf("original error")
	wrapped := Errf("wrapper: %w", original)

	if wrapped == nil {
		t.Fatal("Errf should return an error")
	}

	// Should be unwrappable
	if !strings.Contains(wrapped.Error(), "original error") {
		t.Errorf("Wrapped error should contain original: %q", wrapped.Error())
	}
}

func TestErrfEmpty(t *testing.T) {
	err := Errf("")
	if err == nil {
		t.Fatal("Errf with empty string should still return an error")
	}
}

func TestErrfWithNumbers(t *testing.T) {
	err := Errf("port %d is invalid", 65536)
	if err == nil {
		t.Fatal("Errf should return an error")
	}
	if !strings.Contains(err.Error(), "65536") {
		t.Errorf("Error should contain number: %q", err.Error())
	}
}
