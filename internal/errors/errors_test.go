package errors

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Protocol tests ---

func TestParseStderrStructured(t *testing.T) {
	stderr := `{"wave_error":true,"code":"FLOW_ENV_MISSING","message":"Environment not configured","details":"Check Wavefile"}`

	pe := ParseStderr([]byte(stderr))
	if pe == nil {
		t.Fatal("Should parse structured wave error")
	}
	if pe.Code != "FLOW_ENV_MISSING" {
		t.Errorf("Code = %q", pe.Code)
	}
	if pe.Message != "Environment not configured" {
		t.Errorf("Message = %q", pe.Message)
	}
	if pe.Details != "Check Wavefile" {
		t.Errorf("Details = %q", pe.Details)
	}
}

func TestParseStderrNotWaveError(t *testing.T) {
	// Valid JSON but wave_error is false or missing
	cases := []string{
		`{"wave_error":false,"code":"X","message":"Y"}`,
		`{"code":"X","message":"Y"}`,
		`{"some":"other json"}`,
	}
	for _, c := range cases {
		pe := ParseStderr([]byte(c))
		if pe != nil {
			t.Errorf("Should return nil for non-wave JSON: %s", c)
		}
	}
}

func TestParseStderrRawText(t *testing.T) {
	// Not JSON at all
	pe := ParseStderr([]byte("panic: runtime error: index out of range"))
	if pe != nil {
		t.Error("Should return nil for non-JSON stderr")
	}
}

func TestParseStderrEmpty(t *testing.T) {
	pe := ParseStderr([]byte(""))
	if pe != nil {
		t.Error("Should return nil for empty stderr")
	}
	pe = ParseStderr(nil)
	if pe != nil {
		t.Error("Should return nil for nil stderr")
	}
}

func TestParseStderrMixedOutput(t *testing.T) {
	// Plugin prints some text then a wave error on the last line
	stderr := "some debug output\nmore stuff\n" +
		`{"wave_error":true,"code":"TEST_FAIL","message":"tests failed"}` + "\n"

	pe := ParseStderr([]byte(stderr))
	if pe == nil {
		t.Fatal("Should find wave error in mixed output")
	}
	if pe.Code != "TEST_FAIL" {
		t.Errorf("Code = %q, want TEST_FAIL", pe.Code)
	}
}

// --- Logger tests ---

func TestLogErrorCreatesFile(t *testing.T) {
	dir := t.TempDir()

	pe := &PluginError{
		Code:    "TEST_ERR",
		Message: "something broke",
	}

	err := LogError(dir, "flow", pe, []string{"dev"})
	if err != nil {
		t.Fatalf("LogError failed: %v", err)
	}

	// Should create a .log file
	entries, _ := os.ReadDir(dir)
	if len(entries) == 0 {
		t.Fatal("No log file created")
	}
	if !strings.HasSuffix(entries[0].Name(), ".log") {
		t.Errorf("Log file should end in .log, got %q", entries[0].Name())
	}
}

func TestLogErrorContentIsJSONL(t *testing.T) {
	dir := t.TempDir()

	pe := &PluginError{
		Code:    "TEST_ERR",
		Message: "something broke",
		Details: "check config",
	}

	LogError(dir, "flow", pe, []string{"dev", "--port=3000"})

	// Read the log file
	entries, _ := os.ReadDir(dir)
	data, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatalf("Reading log file: %v", err)
	}

	// Should be valid JSON
	var entry map[string]any
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("Log entry is not valid JSON: %v\nContent: %s", err, data)
	}

	// Check fields
	if entry["plugin"] != "flow" {
		t.Errorf("plugin = %v", entry["plugin"])
	}
	if entry["code"] != "TEST_ERR" {
		t.Errorf("code = %v", entry["code"])
	}
	if entry["message"] != "something broke" {
		t.Errorf("message = %v", entry["message"])
	}
	if entry["ts"] == nil {
		t.Error("ts (timestamp) should be present")
	}
}

func TestLogErrorAppendsMultiple(t *testing.T) {
	dir := t.TempDir()

	pe1 := &PluginError{Code: "ERR_1", Message: "first"}
	pe2 := &PluginError{Code: "ERR_2", Message: "second"}

	LogError(dir, "flow", pe1, []string{})
	LogError(dir, "test", pe2, []string{})

	entries, _ := os.ReadDir(dir)
	data, _ := os.ReadFile(filepath.Join(dir, entries[0].Name()))

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines, got %d", len(lines))
	}
}

// --- Handler / Format tests ---

func TestFormatErrorDebugMode(t *testing.T) {
	pe := &PluginError{
		WaveError: true,
		Code:      "FLOW_CRASH",
		Message:   "plugin crashed",
		Details:   "check logs",
	}

	// Debug mode shows raw JSON
	out := FormatError("flow", "1.0.0", pe, "/tmp/logs/2026-03-15.log", true)
	if !strings.Contains(out, "wave_error") {
		t.Errorf("Debug mode should contain JSON with wave_error, got:\n%s", out)
	}
	if !strings.Contains(out, "FLOW_CRASH") {
		t.Errorf("Should contain error code, got:\n%s", out)
	}
	if !strings.Contains(out, "plugin crashed") {
		t.Errorf("Should contain message, got:\n%s", out)
	}
}

func TestFormatErrorSimpleModeWithMessage(t *testing.T) {
	pe := &PluginError{
		Code:    "USER_ERR",
		Message: "something went wrong",
		Details: "try again",
	}

	// Simple mode: "code: message\ndetails"
	out := FormatError("flow", "1.0.0", pe, "/tmp/logs/2026-03-15.log", false)
	if !strings.Contains(out, "USER_ERR: something went wrong") {
		t.Errorf("Should contain 'code: message' format, got:\n%s", out)
	}
	if !strings.Contains(out, "try again") {
		t.Errorf("Should contain details, got:\n%s", out)
	}
}

func TestFormatErrorSimpleModeWithoutMessage(t *testing.T) {
	pe := &PluginError{
		Code:    "SIMPLE_ERR",
		Message: "",
		Details: "some details",
	}

	// Simple mode without message: "code\ndetails"
	out := FormatError("test", "0.5.0", pe, "", false)
	if !strings.Contains(out, "SIMPLE_ERR") {
		t.Errorf("Should contain error code, got:\n%s", out)
	}
	// Should NOT have colon when message is empty
	if strings.Contains(out, "SIMPLE_ERR:") {
		t.Errorf("Should NOT have colon when message is empty, got:\n%s", out)
	}
	if !strings.Contains(out, "some details") {
		t.Errorf("Should contain details, got:\n%s", out)
	}
}

func TestFormatErrorSimpleModeWithoutDetails(t *testing.T) {
	pe := &PluginError{
		Code:    "NO_DETAILS",
		Message: "error occurred",
	}

	// Simple mode without details: just "code: message"
	out := FormatError("test", "0.5.0", pe, "", false)
	if !strings.Contains(out, "NO_DETAILS: error occurred") {
		t.Errorf("Should contain 'code: message' format, got:\n%s", out)
	}
	// Should not have extra newlines for missing details
	if strings.Count(out, "\n") > 0 {
		// Strip ANSI codes for comparison
		stripped := strings.ReplaceAll(out, "\033[31m", "")
		stripped = strings.ReplaceAll(stripped, "\033[0m", "")
		if strings.Contains(stripped, "\n") && strings.TrimSpace(strings.Split(stripped, "\n")[1]) != "" {
			t.Errorf("Should not have content after first line when no details, got:\n%s", out)
		}
	}
}
