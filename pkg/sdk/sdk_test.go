package sdk

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// --- ReadConfig ---

func TestReadConfig(t *testing.T) {
	input := `{"environment":"staging","port":3000,"watch":true}`
	r := strings.NewReader(input)

	cfg, err := ReadConfigFrom(r)
	if err != nil {
		t.Fatalf("ReadConfigFrom failed: %v", err)
	}
	if cfg["environment"] != "staging" {
		t.Errorf("environment = %v", cfg["environment"])
	}
	if cfg["port"] != float64(3000) {
		t.Errorf("port = %v (type %T)", cfg["port"], cfg["port"])
	}
}

func TestReadConfigEmpty(t *testing.T) {
	r := strings.NewReader("{}")
	cfg, err := ReadConfigFrom(r)
	if err != nil {
		t.Fatalf("ReadConfigFrom failed: %v", err)
	}
	if len(cfg) != 0 {
		t.Errorf("Expected empty config, got %v", cfg)
	}
}

func TestReadConfigInvalidJSON(t *testing.T) {
	r := strings.NewReader("not json")
	_, err := ReadConfigFrom(r)
	if err == nil {
		t.Error("Should fail for invalid JSON")
	}
}

func TestReadConfigNilReader(t *testing.T) {
	r := strings.NewReader("")
	_, err := ReadConfigFrom(r)
	if err == nil {
		t.Error("Should fail for empty input")
	}
}

// --- FormatError ---

func TestFormatWaveError(t *testing.T) {
	var buf bytes.Buffer
	FormatWaveError(&buf, "TEST_CODE", "something broke", "check config")

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
	if pe.Code != "TEST_CODE" {
		t.Errorf("code = %q", pe.Code)
	}
	if pe.Message != "something broke" {
		t.Errorf("message = %q", pe.Message)
	}
	if pe.Details != "check config" {
		t.Errorf("details = %q", pe.Details)
	}
}

func TestFormatWaveErrorNoDetails(t *testing.T) {
	var buf bytes.Buffer
	FormatWaveError(&buf, "ERR", "oops", "")

	var pe map[string]any
	json.Unmarshal(buf.Bytes(), &pe)
	// Details should be omitted or empty
	if d, ok := pe["details"]; ok && d != "" {
		t.Errorf("details should be empty, got %v", d)
	}
}

// --- GetEnv helpers ---

func TestGetPluginEnv(t *testing.T) {
	os.Setenv("WAVE_PLUGIN_NAME", "flow")
	os.Setenv("WAVE_PLUGIN_VERSION", "1.2.0")
	os.Setenv("WAVE_PLUGIN_DIR", "/home/.wave/plugins/wave-cli/flow/current")
	os.Setenv("WAVE_PLUGIN_ASSETS", "/home/.wave/plugins/wave-cli/flow/current/assets")
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
	if env.Dir != "/home/.wave/plugins/wave-cli/flow/current" {
		t.Errorf("Dir = %q", env.Dir)
	}
	if env.Assets != "/home/.wave/plugins/wave-cli/flow/current/assets" {
		t.Errorf("Assets = %q", env.Assets)
	}
	if env.ProjectRoot != "/projects/my-app" {
		t.Errorf("ProjectRoot = %q", env.ProjectRoot)
	}
}
