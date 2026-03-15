package schema

import (
	"testing"
)

// --- FieldType constants ---

func TestFieldTypeConstants(t *testing.T) {
	// Ensure all expected field types exist
	types := []FieldType{TypeString, TypeInt, TypeFloat, TypeBool, TypeMap, TypeArray, TypeAny}
	if len(types) != 7 {
		t.Errorf("Expected 7 field types, got %d", len(types))
	}
}

// --- FieldDef ---

func TestFieldDefRequired(t *testing.T) {
	fd := FieldDef{
		Type:     TypeString,
		Required: true,
		Desc:     "The command to execute",
	}
	if !fd.Required {
		t.Error("Field should be required")
	}
	if fd.Type != TypeString {
		t.Errorf("Type = %q, want %q", fd.Type, TypeString)
	}
}

// --- Schema construction ---

func TestNewSchema(t *testing.T) {
	s := Schema{
		Plugin: "flow",
		Fields: map[string]FieldDef{
			"cmd": {Type: TypeString, Required: true, Desc: "Command to execute"},
			"env": {Type: TypeMap, Required: false, Desc: "Environment variables"},
		},
	}

	if s.Plugin != "flow" {
		t.Errorf("Plugin = %q", s.Plugin)
	}
	if len(s.Fields) != 2 {
		t.Errorf("Fields len = %d", len(s.Fields))
	}
	if !s.Fields["cmd"].Required {
		t.Error("cmd should be required")
	}
}

// --- ValidateSection: happy path ---

func TestValidateSectionHappyPath(t *testing.T) {
	s := Schema{
		Plugin: "flow",
		Fields: map[string]FieldDef{
			"cmd":        {Type: TypeString, Required: true},
			"env":        {Type: TypeMap, Required: false},
			"on_success": {Type: TypeString, Required: false},
			"on_fail":    {Type: TypeString, Required: false},
			"watch":      {Type: TypeArray, Required: false},
		},
	}

	section := map[string]any{
		"build": map[string]any{
			"cmd":        "go build -o bin/app main.go",
			"on_success": "echo 'done'",
			"env":        map[string]any{"GOOS": "linux"},
		},
		"clean": map[string]any{
			"cmd": "rm -rf ./dist",
		},
	}

	errs := s.ValidateSection(section)
	if len(errs) != 0 {
		t.Errorf("Expected no errors, got %v", errs)
	}
}

// --- ValidateSection: missing required field ---

func TestValidateSectionMissingRequired(t *testing.T) {
	s := Schema{
		Plugin: "flow",
		Fields: map[string]FieldDef{
			"cmd": {Type: TypeString, Required: true},
		},
	}

	section := map[string]any{
		"build": map[string]any{
			"on_success": "echo done", // missing "cmd"
		},
	}

	errs := s.ValidateSection(section)
	if len(errs) == 0 {
		t.Fatal("Expected error for missing required field 'cmd'")
	}

	found := false
	for _, e := range errs {
		if containsAll(e.Error(), "cmd", "required", "build") {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected error about missing 'cmd' in 'build', got %v", errs)
	}
}

// --- ValidateSection: wrong type ---

func TestValidateSectionWrongType(t *testing.T) {
	s := Schema{
		Plugin: "flow",
		Fields: map[string]FieldDef{
			"cmd": {Type: TypeString, Required: true},
			"env": {Type: TypeMap, Required: false},
		},
	}

	section := map[string]any{
		"build": map[string]any{
			"cmd": 123,         // should be string
			"env": "not a map", // should be map
		},
	}

	errs := s.ValidateSection(section)
	if len(errs) < 2 {
		t.Fatalf("Expected at least 2 errors, got %d: %v", len(errs), errs)
	}
}

// --- ValidateSection: unknown field ---

func TestValidateSectionUnknownField(t *testing.T) {
	s := Schema{
		Plugin: "flow",
		Fields: map[string]FieldDef{
			"cmd": {Type: TypeString, Required: true},
		},
	}

	section := map[string]any{
		"build": map[string]any{
			"cmd":     "go build",
			"unknown": "value", // not in schema
		},
	}

	errs := s.ValidateSection(section)
	if len(errs) == 0 {
		t.Fatal("Expected error for unknown field 'unknown'")
	}

	found := false
	for _, e := range errs {
		if containsAll(e.Error(), "unknown", "build") {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected error about 'unknown' field in 'build', got %v", errs)
	}
}

// --- ValidateSection: nil section ---

func TestValidateSectionNil(t *testing.T) {
	s := Schema{
		Plugin: "flow",
		Fields: map[string]FieldDef{
			"cmd": {Type: TypeString, Required: true},
		},
	}

	errs := s.ValidateSection(nil)
	if len(errs) != 0 {
		t.Errorf("Nil section should be valid (no commands defined), got %v", errs)
	}
}

// --- ValidateSection: empty section ---

func TestValidateSectionEmpty(t *testing.T) {
	s := Schema{
		Plugin: "flow",
		Fields: map[string]FieldDef{
			"cmd": {Type: TypeString, Required: true},
		},
	}

	errs := s.ValidateSection(map[string]any{})
	if len(errs) != 0 {
		t.Errorf("Empty section should be valid, got %v", errs)
	}
}

// --- ValidateSection: entry is not a map ---

func TestValidateSectionEntryNotMap(t *testing.T) {
	s := Schema{
		Plugin: "flow",
		Fields: map[string]FieldDef{
			"cmd": {Type: TypeString, Required: true},
		},
	}

	section := map[string]any{
		"build": "just a string", // should be a map
	}

	errs := s.ValidateSection(section)
	if len(errs) == 0 {
		t.Fatal("Expected error when entry is not a map")
	}
}

// --- ValidateSection: array type ---

func TestValidateSectionArrayType(t *testing.T) {
	s := Schema{
		Plugin: "flow",
		Fields: map[string]FieldDef{
			"cmd":   {Type: TypeString, Required: true},
			"watch": {Type: TypeArray, Required: false},
		},
	}

	section := map[string]any{
		"dev": map[string]any{
			"cmd":   "go run main.go",
			"watch": []any{"src/**/*.go", "templates/*.html"},
		},
	}

	errs := s.ValidateSection(section)
	if len(errs) != 0 {
		t.Errorf("Expected no errors, got %v", errs)
	}
}

// --- ValidateSection: bool type ---

func TestValidateSectionBoolType(t *testing.T) {
	s := Schema{
		Plugin: "test",
		Fields: map[string]FieldDef{
			"verbose": {Type: TypeBool, Required: false},
		},
	}

	section := map[string]any{
		"unit": map[string]any{
			"verbose": true,
		},
	}

	errs := s.ValidateSection(section)
	if len(errs) != 0 {
		t.Errorf("Expected no errors for bool type, got %v", errs)
	}
}

// --- ValidateSection: int64 type (TOML integers) ---

func TestValidateSectionIntType(t *testing.T) {
	s := Schema{
		Plugin: "test",
		Fields: map[string]FieldDef{
			"port": {Type: TypeInt, Required: false},
		},
	}

	section := map[string]any{
		"server": map[string]any{
			"port": int64(8080),
		},
	}

	errs := s.ValidateSection(section)
	if len(errs) != 0 {
		t.Errorf("Expected no errors for int type, got %v", errs)
	}
}

// --- ValidateSection: float type ---

func TestValidateSectionFloatType(t *testing.T) {
	s := Schema{
		Plugin: "test",
		Fields: map[string]FieldDef{
			"threshold": {Type: TypeFloat, Required: false},
		},
	}

	section := map[string]any{
		"coverage": map[string]any{
			"threshold": 80.5,
		},
	}

	errs := s.ValidateSection(section)
	if len(errs) != 0 {
		t.Errorf("Expected no errors for float type, got %v", errs)
	}
}

// --- ValidateSection: TypeAny accepts anything ---

func TestValidateSectionAnyType(t *testing.T) {
	s := Schema{
		Plugin: "test",
		Fields: map[string]FieldDef{
			"data": {Type: TypeAny, Required: false},
		},
	}

	// Test with different types
	for _, val := range []any{"string", int64(42), true, 3.14, []any{"a", "b"}, map[string]any{"k": "v"}} {
		section := map[string]any{
			"entry": map[string]any{
				"data": val,
			},
		}
		errs := s.ValidateSection(section)
		if len(errs) != 0 {
			t.Errorf("TypeAny should accept %T, got errors: %v", val, errs)
		}
	}
}

// --- ParseSchemaFile ---

func TestParseSchemaFile(t *testing.T) {
	content := `plugin = "flow"

[fields.cmd]
type = "string"
required = true
desc = "The command to execute"

[fields.env]
type = "map"
required = false
desc = "Environment variables"

[fields.on_success]
type = "string"
required = false
desc = "Command to run on success"

[fields.on_fail]
type = "string"
required = false
desc = "Command to run on failure"

[fields.watch]
type = "array"
required = false
desc = "File patterns to watch"
`

	s, err := ParseSchemaFromBytes([]byte(content))
	if err != nil {
		t.Fatalf("ParseSchemaFromBytes failed: %v", err)
	}

	if s.Plugin != "flow" {
		t.Errorf("Plugin = %q", s.Plugin)
	}
	if len(s.Fields) != 5 {
		t.Errorf("Fields len = %d, want 5", len(s.Fields))
	}
	if !s.Fields["cmd"].Required {
		t.Error("cmd should be required")
	}
	if s.Fields["env"].Type != TypeMap {
		t.Errorf("env type = %q", s.Fields["env"].Type)
	}
	if s.Fields["watch"].Type != TypeArray {
		t.Errorf("watch type = %q", s.Fields["watch"].Type)
	}
}

func TestParseSchemaFromBytesInvalidTOML(t *testing.T) {
	_, err := ParseSchemaFromBytes([]byte("[[[broken"))
	if err == nil {
		t.Error("Should fail for invalid TOML")
	}
}

func TestParseSchemaFromBytesEmpty(t *testing.T) {
	_, err := ParseSchemaFromBytes([]byte(""))
	if err == nil {
		t.Error("Should fail for empty input")
	}
}

func TestParseSchemaFromBytesMissingPlugin(t *testing.T) {
	content := `[fields.cmd]
type = "string"
required = true
`
	_, err := ParseSchemaFromBytes([]byte(content))
	if err == nil {
		t.Error("Should fail when plugin name is missing")
	}
}

// --- Multiple errors in one entry ---

func TestValidateSectionMultipleErrors(t *testing.T) {
	s := Schema{
		Plugin: "flow",
		Fields: map[string]FieldDef{
			"cmd":        {Type: TypeString, Required: true},
			"on_success": {Type: TypeString, Required: false},
		},
	}

	section := map[string]any{
		"build": map[string]any{
			"on_success": 42,         // wrong type, and cmd is missing
			"unknown":    "whatever", // unknown field
		},
	}

	errs := s.ValidateSection(section)
	if len(errs) < 3 {
		t.Fatalf("Expected at least 3 errors (missing cmd, wrong type on_success, unknown field), got %d: %v", len(errs), errs)
	}
}

// --- helper ---

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
