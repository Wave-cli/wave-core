// Package config provides configuration reading and parsing for wave plugins.
// Plugins receive their configuration section from the Wavefile as JSON via stdin.
package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
)

// Config wraps the raw plugin configuration map and provides typed access.
// wave-core passes the plugin's entire TOML section as a JSON object on stdin.
// The plugin receives it as-is — no schema validation, no transformation.
type Config struct {
	data map[string]any
}

// Raw returns the underlying configuration map.
func (c *Config) Raw() map[string]any {
	return c.data
}

// Get returns the value for key and whether it exists.
func (c *Config) Get(key string) (any, bool) {
	v, ok := c.data[key]
	return v, ok
}

// Has returns true if key exists in the config.
func (c *Config) Has(key string) bool {
	_, ok := c.data[key]
	return ok
}

// String returns the string value for key, or ("", false) if missing/wrong type.
func (c *Config) String(key string) (string, bool) {
	v, ok := c.data[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// Bool returns the bool value for key, or (false, false) if missing/wrong type.
func (c *Config) Bool(key string) (bool, bool) {
	v, ok := c.data[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

// Float returns the float64 value for key, or (0, false) if missing/wrong type.
// JSON numbers are always decoded as float64.
func (c *Config) Float(key string) (float64, bool) {
	v, ok := c.data[key]
	if !ok {
		return 0, false
	}
	f, ok := v.(float64)
	return f, ok
}

// Int returns the int value for key, or (0, false) if missing/wrong type.
// JSON numbers are decoded as float64, so this converts to int.
func (c *Config) Int(key string) (int, bool) {
	v, ok := c.data[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	}
	return 0, false
}

// Map returns the map value for key, or (nil, false) if missing/wrong type.
func (c *Config) Map(key string) (map[string]any, bool) {
	v, ok := c.data[key]
	if !ok {
		return nil, false
	}
	m, ok := v.(map[string]any)
	return m, ok
}

// Keys returns all top-level keys in sorted order.
func (c *Config) Keys() []string {
	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Read reads the plugin configuration from os.Stdin.
// wave-core passes the plugin's config section as a JSON object on stdin.
func Read() (*Config, error) {
	return ReadFrom(os.Stdin)
}

// ReadFrom reads plugin configuration from an arbitrary reader.
// This is useful for testing.
func ReadFrom(r io.Reader) (*Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading config from stdin: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty config input")
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing config JSON: %w", err)
	}

	return &Config{data: raw}, nil
}

// New creates a Config from an existing map.
// Useful for testing or constructing configs programmatically.
func New(data map[string]any) *Config {
	if data == nil {
		data = make(map[string]any)
	}
	return &Config{data: data}
}
