// Package schema provides Wavefile section validation against plugin-defined
// schemas. Each plugin can ship a schema file that describes the structure
// of its Wavefile section, and wave-core uses this to validate before
// passing config to the plugin.
package schema

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// FieldType represents the type of a schema field.
type FieldType string

const (
	TypeString FieldType = "string"
	TypeInt    FieldType = "int"
	TypeFloat  FieldType = "float"
	TypeBool   FieldType = "bool"
	TypeMap    FieldType = "map"
	TypeArray  FieldType = "array"
	TypeAny    FieldType = "any"
)

// FieldDef describes a single field in a plugin's config schema.
type FieldDef struct {
	Type     FieldType `toml:"type"`
	Required bool      `toml:"required"`
	Desc     string    `toml:"desc"`
}

// Schema describes the expected structure of a plugin's Wavefile section.
// Each entry in the section (e.g. build, clean, dev) is validated against
// these field definitions.
type Schema struct {
	Plugin string              `toml:"plugin"`
	Fields map[string]FieldDef `toml:"fields"`
}

// ValidationError represents a single validation failure.
type ValidationError struct {
	Entry   string // e.g. "build"
	Field   string // e.g. "cmd"
	Message string
}

func (e *ValidationError) Error() string {
	if e.Entry != "" && e.Field != "" {
		return fmt.Sprintf("[%s.%s] %s", e.Entry, e.Field, e.Message)
	}
	if e.Entry != "" {
		return fmt.Sprintf("[%s] %s", e.Entry, e.Message)
	}
	return e.Message
}

// ValidateSection validates a raw Wavefile section (map[string]any) against
// this schema. Each top-level key in the section is treated as a named
// command entry (e.g. "build", "clean") whose value must be a map matching
// the field definitions.
//
// Returns a slice of validation errors. Empty slice means valid.
func (s *Schema) ValidateSection(section map[string]any) []ValidationError {
	if section == nil || len(section) == 0 {
		return nil
	}

	var errs []ValidationError

	for entryName, entryVal := range section {
		entryMap, ok := entryVal.(map[string]any)
		if !ok {
			errs = append(errs, ValidationError{
				Entry:   entryName,
				Message: fmt.Sprintf("expected a map, got %T", entryVal),
			})
			continue
		}

		// Check required fields are present
		for fieldName, fieldDef := range s.Fields {
			if fieldDef.Required {
				if _, exists := entryMap[fieldName]; !exists {
					errs = append(errs, ValidationError{
						Entry:   entryName,
						Field:   fieldName,
						Message: fmt.Sprintf("required field %q is missing in %q", fieldName, entryName),
					})
				}
			}
		}

		// Validate each provided field
		for key, val := range entryMap {
			fieldDef, known := s.Fields[key]
			if !known {
				errs = append(errs, ValidationError{
					Entry:   entryName,
					Field:   key,
					Message: fmt.Sprintf("unknown field %q in %q", key, entryName),
				})
				continue
			}

			if err := validateType(val, fieldDef.Type); err != nil {
				errs = append(errs, ValidationError{
					Entry:   entryName,
					Field:   key,
					Message: fmt.Sprintf("field %q in %q: %s", key, entryName, err.Error()),
				})
			}
		}
	}

	return errs
}

// validateType checks whether val matches the expected FieldType.
func validateType(val any, expected FieldType) error {
	if expected == TypeAny {
		return nil
	}

	switch expected {
	case TypeString:
		if _, ok := val.(string); !ok {
			return fmt.Errorf("expected string, got %T", val)
		}
	case TypeInt:
		if _, ok := val.(int64); !ok {
			return fmt.Errorf("expected int, got %T", val)
		}
	case TypeFloat:
		if _, ok := val.(float64); !ok {
			return fmt.Errorf("expected float, got %T", val)
		}
	case TypeBool:
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", val)
		}
	case TypeMap:
		if _, ok := val.(map[string]any); !ok {
			return fmt.Errorf("expected map, got %T", val)
		}
	case TypeArray:
		if _, ok := val.([]any); !ok {
			return fmt.Errorf("expected array, got %T", val)
		}
	default:
		return fmt.Errorf("unknown field type %q", expected)
	}

	return nil
}

// ParseSchemaFromBytes parses a schema definition from TOML bytes.
func ParseSchemaFromBytes(data []byte) (*Schema, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty schema data")
	}

	var s Schema
	if _, err := toml.Decode(string(data), &s); err != nil {
		return nil, fmt.Errorf("parsing schema: %w", err)
	}

	if s.Plugin == "" {
		return nil, fmt.Errorf("schema missing required 'plugin' field")
	}

	if s.Fields == nil {
		s.Fields = make(map[string]FieldDef)
	}

	return &s, nil
}
