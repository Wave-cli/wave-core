package rules

import (
	"strings"
	"testing"
)

// --- Rule interface ---

func TestBuiltInRulesImplementInterface(t *testing.T) {
	// Ensure all built-in rules implement the Rule interface
	var _ Rule = &NoNestedHeadersRule{}
	var _ Rule = &NoGlobalNamespaceRule{}
}

// --- NoNestedHeadersRule ---

func TestNoNestedHeadersRulePass(t *testing.T) {
	r := &NoNestedHeadersRule{PluginName: "flow"}

	// Valid: [flow] with inline maps, no [flow.build]
	raw := `
[project]
name = "test"
version = "0.1.0"

[flow]
clean = { cmd = "rm -rf ./dist" }
build = { cmd = "go build" }
`

	errs := r.Check([]byte(raw))
	if len(errs) != 0 {
		t.Errorf("Expected no errors for valid config, got %v", errs)
	}
}

func TestNoNestedHeadersRuleDetectsSubHeader(t *testing.T) {
	r := &NoNestedHeadersRule{PluginName: "flow"}

	// Invalid: [flow.build] is a nested header
	raw := `
[project]
name = "test"
version = "0.1.0"

[flow]
clean = { cmd = "rm -rf ./dist" }

[flow.build]
cmd = "go build"
`

	errs := r.Check([]byte(raw))
	if len(errs) == 0 {
		t.Fatal("Expected WaveStructureError for [flow.build]")
	}

	found := false
	for _, e := range errs {
		if e.Code == "WaveStructureError" && strings.Contains(e.Message, "flow.build") {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected WaveStructureError mentioning 'flow.build', got %v", errs)
	}
}

func TestNoNestedHeadersRuleMultipleSubHeaders(t *testing.T) {
	r := &NoNestedHeadersRule{PluginName: "flow"}

	raw := `
[flow]
clean = { cmd = "rm -rf" }

[flow.build]
cmd = "go build"

[flow.dev]
cmd = "go run"
`

	errs := r.Check([]byte(raw))
	if len(errs) < 2 {
		t.Fatalf("Expected at least 2 errors, got %d: %v", len(errs), errs)
	}
}

func TestNoNestedHeadersRuleIgnoresOtherPlugins(t *testing.T) {
	r := &NoNestedHeadersRule{PluginName: "flow"}

	// [test.unit] should NOT trigger flow rule
	raw := `
[flow]
build = { cmd = "go build" }

[test.unit]
cmd = "go test"
`

	errs := r.Check([]byte(raw))
	if len(errs) != 0 {
		t.Errorf("Should not flag [test.unit] for flow plugin, got %v", errs)
	}
}

func TestNoNestedHeadersRuleEmptyInput(t *testing.T) {
	r := &NoNestedHeadersRule{PluginName: "flow"}
	errs := r.Check([]byte(""))
	if len(errs) != 0 {
		t.Errorf("Empty input should be valid, got %v", errs)
	}
}

func TestNoNestedHeadersRuleNilInput(t *testing.T) {
	r := &NoNestedHeadersRule{PluginName: "flow"}
	errs := r.Check(nil)
	if len(errs) != 0 {
		t.Errorf("Nil input should be valid, got %v", errs)
	}
}

// --- NoGlobalNamespaceRule ---

func TestNoGlobalNamespaceRulePass(t *testing.T) {
	r := &NoGlobalNamespaceRule{
		PluginName: "flow",
		PluginKeys: []string{"cmd", "on_success", "on_fail", "env", "watch"},
	}

	// Valid: plugin-specific keys are inside [flow], not at root level
	raw := `
[project]
name = "test"
version = "0.1.0"

[flow]
build = { cmd = "go build", on_success = "echo done" }
`

	errs := r.Check([]byte(raw))
	if len(errs) != 0 {
		t.Errorf("Expected no errors, got %v", errs)
	}
}

func TestNoGlobalNamespaceRuleDetectsLeakedKeys(t *testing.T) {
	r := &NoGlobalNamespaceRule{
		PluginName: "flow",
		PluginKeys: []string{"cmd", "on_success", "on_fail", "env", "watch"},
	}

	// Invalid: "cmd" is at root level, not under [flow]
	raw := `
[project]
name = "test"
version = "0.1.0"

cmd = "go build"
on_success = "echo done"
`

	errs := r.Check([]byte(raw))
	if len(errs) == 0 {
		t.Fatal("Expected GlobalNamespaceError for 'cmd' at root level")
	}

	found := false
	for _, e := range errs {
		if e.Code == "GlobalNamespaceError" {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected GlobalNamespaceError, got %v", errs)
	}
}

func TestNoGlobalNamespaceRuleIgnoresProjectKeys(t *testing.T) {
	r := &NoGlobalNamespaceRule{
		PluginName: "flow",
		PluginKeys: []string{"cmd", "on_success", "on_fail"},
	}

	// "name" and "version" are NOT plugin keys; they belong to [project]
	raw := `
name = "test"
version = "0.1.0"
`

	errs := r.Check([]byte(raw))
	if len(errs) != 0 {
		t.Errorf("Should not flag non-plugin keys, got %v", errs)
	}
}

func TestNoGlobalNamespaceRuleEmptyInput(t *testing.T) {
	r := &NoGlobalNamespaceRule{
		PluginName: "flow",
		PluginKeys: []string{"cmd"},
	}
	errs := r.Check([]byte(""))
	if len(errs) != 0 {
		t.Errorf("Empty input should be valid, got %v", errs)
	}
}

// --- RuleEngine ---

func TestRuleEngineNoRules(t *testing.T) {
	engine := NewRuleEngine()
	errs := engine.Run([]byte("[flow]\nbuild = { cmd = \"go build\" }"))
	if len(errs) != 0 {
		t.Errorf("Engine with no rules should return no errors, got %v", errs)
	}
}

func TestRuleEngineMultipleRules(t *testing.T) {
	engine := NewRuleEngine()
	engine.Register(&NoNestedHeadersRule{PluginName: "flow"})
	engine.Register(&NoGlobalNamespaceRule{
		PluginName: "flow",
		PluginKeys: []string{"cmd", "on_success"},
	})

	// Valid input
	raw := `
[flow]
build = { cmd = "go build", on_success = "echo done" }
`
	errs := engine.Run([]byte(raw))
	if len(errs) != 0 {
		t.Errorf("Expected no errors for valid input, got %v", errs)
	}
}

func TestRuleEngineMultipleRulesWithErrors(t *testing.T) {
	engine := NewRuleEngine()
	engine.Register(&NoNestedHeadersRule{PluginName: "flow"})
	engine.Register(&NoGlobalNamespaceRule{
		PluginName: "flow",
		PluginKeys: []string{"cmd"},
	})

	// Both rules should fail
	raw := `
cmd = "leaked"

[flow.build]
cmd = "go build"
`
	errs := engine.Run([]byte(raw))
	if len(errs) < 2 {
		t.Fatalf("Expected at least 2 errors (one per rule), got %d: %v", len(errs), errs)
	}
}

func TestRuleEngineRegisterForPlugin(t *testing.T) {
	engine := NewRuleEngine()
	engine.RegisterForPlugin("flow", []string{"cmd", "on_success", "on_fail", "env", "watch"})

	// Should have registered both rules
	if len(engine.rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(engine.rules))
	}

	// Valid input should pass
	raw := `
[flow]
build = { cmd = "go build" }
`
	errs := engine.Run([]byte(raw))
	if len(errs) != 0 {
		t.Errorf("Expected no errors, got %v", errs)
	}
}

func TestRuleEngineRegisterForPluginCatchesErrors(t *testing.T) {
	engine := NewRuleEngine()
	engine.RegisterForPlugin("flow", []string{"cmd", "on_success"})

	// Nested header + leaked key
	raw := `
cmd = "leaked"

[flow.build]
cmd = "go build"
`
	errs := engine.Run([]byte(raw))
	if len(errs) < 2 {
		t.Fatalf("Expected at least 2 errors, got %d: %v", len(errs), errs)
	}
}

// --- RuleViolation ---

func TestRuleViolationError(t *testing.T) {
	v := RuleViolation{
		Code:    "WaveStructureError",
		Message: "nested headers disallowed: [flow.build]",
	}

	s := v.Error()
	if !strings.Contains(s, "WaveStructureError") {
		t.Errorf("Should contain code, got %q", s)
	}
	if !strings.Contains(s, "nested headers disallowed") {
		t.Errorf("Should contain message, got %q", s)
	}
}

func TestNoGlobalNamespaceRuleDetectsKeysInWrongSection(t *testing.T) {
	r := &NoGlobalNamespaceRule{
		PluginName: "flow",
		PluginKeys: []string{"cmd", "on_success", "on_fail", "env", "watch"},
	}

	// "cmd" under [project] should be flagged
	raw := `
[project]
name = "test"
cmd = "go build"

[flow]
build = { cmd = "go build" }
`

	errs := r.Check([]byte(raw))
	if len(errs) != 1 {
		t.Fatalf("Expected 1 error for 'cmd' under [project], got %d: %v", len(errs), errs)
	}
	if errs[0].Code != "GlobalNamespaceError" {
		t.Errorf("Expected GlobalNamespaceError, got %s", errs[0].Code)
	}
	if !strings.Contains(errs[0].Message, "[project] section") {
		t.Errorf("Error should mention [project] section, got %q", errs[0].Message)
	}
}

func TestNoGlobalNamespaceRuleAllowsKeysInOwnSection(t *testing.T) {
	r := &NoGlobalNamespaceRule{
		PluginName: "flow",
		PluginKeys: []string{"cmd", "on_success", "on_fail", "env", "watch"},
	}

	// Keys under [flow] should NOT be flagged, even bare ones
	raw := `
[project]
name = "test"

[flow]
cmd = "go build"
on_success = "echo done"
`

	errs := r.Check([]byte(raw))
	if len(errs) != 0 {
		t.Errorf("Keys under [flow] should be allowed, got %v", errs)
	}
}

func TestNoGlobalNamespaceRuleDetectsKeysInMultipleWrongSections(t *testing.T) {
	r := &NoGlobalNamespaceRule{
		PluginName: "flow",
		PluginKeys: []string{"cmd", "on_success"},
	}

	// "cmd" under [project] and "on_success" under [deploy] both flagged
	raw := `
[project]
name = "test"
cmd = "leaked"

[deploy]
on_success = "leaked too"

[flow]
build = { cmd = "go build" }
`

	errs := r.Check([]byte(raw))
	if len(errs) != 2 {
		t.Fatalf("Expected 2 errors, got %d: %v", len(errs), errs)
	}

	// First should mention [project]
	if !strings.Contains(errs[0].Message, "[project]") {
		t.Errorf("First error should mention [project], got %q", errs[0].Message)
	}
	// Second should mention [deploy]
	if !strings.Contains(errs[1].Message, "[deploy]") {
		t.Errorf("Second error should mention [deploy], got %q", errs[1].Message)
	}
}

func TestNoGlobalNamespaceRuleDetectsRootAndSectionLeaks(t *testing.T) {
	r := &NoGlobalNamespaceRule{
		PluginName: "flow",
		PluginKeys: []string{"cmd", "env"},
	}

	// "cmd" at root level + "env" under [project] — both flagged
	raw := `
cmd = "root leak"

[project]
name = "test"
env = "section leak"

[flow]
build = { cmd = "go build" }
`

	errs := r.Check([]byte(raw))
	if len(errs) != 2 {
		t.Fatalf("Expected 2 errors, got %d: %v", len(errs), errs)
	}

	if !strings.Contains(errs[0].Message, "global level") {
		t.Errorf("First error should mention global level, got %q", errs[0].Message)
	}
	if !strings.Contains(errs[1].Message, "[project] section") {
		t.Errorf("Second error should mention [project] section, got %q", errs[1].Message)
	}
}

func TestNoGlobalNamespaceRuleNilInput(t *testing.T) {
	r := &NoGlobalNamespaceRule{
		PluginName: "flow",
		PluginKeys: []string{"cmd"},
	}
	errs := r.Check(nil)
	if len(errs) != 0 {
		t.Errorf("Nil input should be valid, got %v", errs)
	}
}

// --- Edge cases ---

func TestNoNestedHeadersRuleWithComments(t *testing.T) {
	r := &NoNestedHeadersRule{PluginName: "flow"}

	raw := `
# This is a comment about [flow.build]
[flow]
build = { cmd = "go build" }
# Another comment [flow.dev]
`
	errs := r.Check([]byte(raw))
	if len(errs) != 0 {
		t.Errorf("Comments containing [flow.xxx] should not trigger errors, got %v", errs)
	}
}

func TestNoNestedHeadersRuleInlineStringContainingBrackets(t *testing.T) {
	r := &NoNestedHeadersRule{PluginName: "flow"}

	// A value containing "[flow.build]" as a string should not trigger
	raw := `
[flow]
build = { cmd = "echo '[flow.build] is done'" }
`
	errs := r.Check([]byte(raw))
	if len(errs) != 0 {
		t.Errorf("String values containing [flow.xxx] should not trigger errors, got %v", errs)
	}
}
