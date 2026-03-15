package pluginmgmt

import (
	"fmt"
	"strings"
)

// PluginRef is a parsed reference to a plugin: org/name[@version].
type PluginRef struct {
	Org     string
	Name    string
	Version string // empty when unspecified
}

// FullName returns "org/name" without a version qualifier.
func (r PluginRef) FullName() string {
	return r.Org + "/" + r.Name
}

// ParsePluginRef parses a string like "wave-cli/flow" or "wave-cli/flow@1.2.0"
// into a PluginRef. It requires exactly one slash and non-empty org and name.
func ParsePluginRef(s string) (PluginRef, error) {
	if s == "" {
		return PluginRef{}, fmt.Errorf("empty plugin reference")
	}

	// Split off optional @version first.
	var version string
	if idx := strings.Index(s, "@"); idx >= 0 {
		version = s[idx+1:]
		s = s[:idx]
	}

	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return PluginRef{}, fmt.Errorf("invalid plugin reference %q: must be org/name", s)
	}

	org, name := parts[0], parts[1]
	if org == "" {
		return PluginRef{}, fmt.Errorf("invalid plugin reference: empty org")
	}
	if name == "" {
		return PluginRef{}, fmt.Errorf("invalid plugin reference: empty name")
	}

	return PluginRef{
		Org:     org,
		Name:    name,
		Version: version,
	}, nil
}
