// Package rules provides a Wavefile rule engine that enforces structural
// constraints on plugin configuration. Plugins register rules that are
// checked against the raw Wavefile content before parsing.
//
// Built-in rules:
//   - NoNestedHeadersRule: disallows [plugin.subkey] TOML headers
//   - NoGlobalNamespaceRule: disallows plugin-specific keys outside their section
//
// The engine is extensible: any type implementing the Rule interface can be
// registered.
package rules

import (
	"fmt"
	"strings"
)

// RuleViolation represents a single rule check failure.
type RuleViolation struct {
	Code    string // e.g. "WaveStructureError", "GlobalNamespaceError"
	Message string
}

func (v *RuleViolation) Error() string {
	return fmt.Sprintf("%s: %s", v.Code, v.Message)
}

// Rule is the interface for Wavefile structural rules.
// Implementations receive the raw Wavefile bytes and return any violations.
type Rule interface {
	Check(raw []byte) []RuleViolation
}

// RuleEngine manages a set of rules and runs them against Wavefile content.
type RuleEngine struct {
	rules []Rule
}

// NewRuleEngine creates an empty rule engine.
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{}
}

// Register adds a rule to the engine.
func (e *RuleEngine) Register(r Rule) {
	e.rules = append(e.rules, r)
}

// RegisterForPlugin registers the standard rules for a plugin:
// NoNestedHeadersRule and NoGlobalNamespaceRule.
func (e *RuleEngine) RegisterForPlugin(pluginName string, pluginKeys []string) {
	e.Register(&NoNestedHeadersRule{PluginName: pluginName})
	e.Register(&NoGlobalNamespaceRule{
		PluginName: pluginName,
		PluginKeys: pluginKeys,
	})
}

// Run executes all registered rules against the raw Wavefile content.
// Returns all collected violations.
func (e *RuleEngine) Run(raw []byte) []RuleViolation {
	var all []RuleViolation
	for _, r := range e.rules {
		if violations := r.Check(raw); len(violations) > 0 {
			all = append(all, violations...)
		}
	}
	return all
}

// =============================================================================
// NoNestedHeadersRule
// =============================================================================

// NoNestedHeadersRule checks that no TOML sub-headers like [plugin.subkey]
// exist in the Wavefile. All commands must be defined as inline maps under
// the single [plugin] header.
type NoNestedHeadersRule struct {
	PluginName string
}

// Check scans raw TOML lines for sub-headers like [flow.build].
func (r *NoNestedHeadersRule) Check(raw []byte) []RuleViolation {
	if len(raw) == 0 {
		return nil
	}

	var violations []RuleViolation
	prefix := "[" + r.PluginName + "."

	lines := strings.Split(string(raw), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip comments
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Skip lines that contain = before [ (inline values with brackets)
		if eqIdx := strings.Index(trimmed, "="); eqIdx >= 0 {
			bracketIdx := strings.Index(trimmed, prefix)
			if bracketIdx < 0 || bracketIdx > eqIdx {
				continue
			}
		}

		// Check for [plugin.subkey] pattern at the start of a line
		if strings.HasPrefix(trimmed, prefix) && strings.HasSuffix(trimmed, "]") {
			// Extract the sub-header name
			header := trimmed[1 : len(trimmed)-1] // strip [ and ]
			violations = append(violations, RuleViolation{
				Code:    "WaveStructureError",
				Message: fmt.Sprintf("nested headers are disallowed: [%s]. All commands must be defined as inline maps under [%s]", header, r.PluginName),
			})
		}
	}

	return violations
}

// =============================================================================
// NoGlobalNamespaceRule
// =============================================================================

// NoGlobalNamespaceRule checks that plugin-specific keys (like "cmd",
// "on_success") do not appear at the root level of the Wavefile, outside
// any [section] header.
type NoGlobalNamespaceRule struct {
	PluginName string
	PluginKeys []string
}

// Check scans the raw TOML for plugin keys appearing outside the plugin's
// own section. It tracks the current TOML section and flags any plugin-specific
// keys found in the root level or under a different section.
func (r *NoGlobalNamespaceRule) Check(raw []byte) []RuleViolation {
	if len(raw) == 0 {
		return nil
	}

	var violations []RuleViolation
	currentSection := "" // empty = root level

	lines := strings.Split(string(raw), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Track section headers
		if strings.HasPrefix(trimmed, "[") && strings.Contains(trimmed, "]") {
			// Extract section name: [foo] -> "foo", [foo.bar] -> "foo.bar"
			header := trimmed[1:strings.Index(trimmed, "]")]
			currentSection = strings.TrimSpace(header)
			continue
		}

		// Only flag keys that appear outside the plugin's section
		if currentSection == r.PluginName {
			continue
		}

		// Extract key name (before =)
		if eqIdx := strings.Index(trimmed, "="); eqIdx > 0 {
			key := strings.TrimSpace(trimmed[:eqIdx])
			for _, pk := range r.PluginKeys {
				if key == pk {
					location := "global level"
					if currentSection != "" {
						location = fmt.Sprintf("[%s] section", currentSection)
					}
					violations = append(violations, RuleViolation{
						Code:    "GlobalNamespaceError",
						Message: fmt.Sprintf("key %q belongs to the [%s] section but was found at the %s. Move it under [%s]", key, r.PluginName, location, r.PluginName),
					})
				}
			}
		}
	}

	return violations
}
