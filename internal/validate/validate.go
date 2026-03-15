// Package validate orchestrates Wavefile validation for a plugin.
// It combines the schema validation module and the rules engine to provide
// a single entry point for pre-execution config validation.
package validate

import (
	"github.com/wave-cli/wave-core/internal/rules"
	"github.com/wave-cli/wave-core/internal/schema"
)

// ValidatePluginConfig validates a plugin's Wavefile configuration using both
// the rules engine (structural checks on raw TOML) and schema validation
// (type/required checks on parsed section).
//
// Parameters:
//   - pluginName: the plugin's section name (e.g. "flow")
//   - rawWavefile: raw bytes of the entire Wavefile (for rules engine)
//   - section: the parsed plugin section (map[string]any, may be nil)
//   - schemaBytes: raw bytes of the plugin's schema file (nil = skip)
//
// Returns a slice of human-readable error strings. Empty = valid.
func ValidatePluginConfig(pluginName string, rawWavefile []byte, section map[string]any, schemaBytes []byte) []string {
	if len(schemaBytes) == 0 {
		return nil
	}

	var errs []string

	// Parse the schema
	s, err := schema.ParseSchemaFromBytes(schemaBytes)
	if err != nil {
		return []string{"schema parse error: " + err.Error()}
	}

	// 1. Rules engine: structural checks on raw TOML
	engine := rules.NewRuleEngine()
	pluginKeys := make([]string, 0, len(s.Fields))
	for k := range s.Fields {
		pluginKeys = append(pluginKeys, k)
	}
	engine.RegisterForPlugin(pluginName, pluginKeys)

	violations := engine.Run(rawWavefile)
	for _, v := range violations {
		errs = append(errs, v.Error())
	}

	// 2. Schema validation: type/required checks on parsed section
	if section != nil {
		schemaErrs := s.ValidateSection(section)
		for _, e := range schemaErrs {
			errs = append(errs, e.Error())
		}
	}

	return errs
}
