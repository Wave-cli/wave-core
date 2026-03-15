// Package version provides build-time version info and compatibility checks.
// Version, Commit, and Date are set via -ldflags at build time.
package version

import (
	"fmt"
	"strconv"
	"strings"
)

// Set via -ldflags:
//
//	go build -ldflags "-X github.com/wave-cli/wave-core/internal/version.version=1.0.0
//	  -X github.com/wave-cli/wave-core/internal/version.commit=abc1234
//	  -X github.com/wave-cli/wave-core/internal/version.date=2026-03-15"
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Info holds structured version information.
type Info struct {
	Version string
	Commit  string
	Date    string
}

// Get returns the current build version info.
func Get() Info {
	return Info{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
}

// String returns a short version string like "wave v1.0.0".
func (i Info) String() string {
	return fmt.Sprintf("wave %s", i.Version)
}

// Full returns a detailed version string including commit and date.
func (i Info) Full() string {
	return fmt.Sprintf("wave %s (commit: %s, built: %s)", i.Version, i.Commit, i.Date)
}

// SatisfiesMin checks whether current >= minVersion using semver comparison.
// Returns true if:
//   - minVersion is empty (no constraint)
//   - current is "dev" (development builds satisfy anything)
//   - current >= minVersion by semver rules
//
// Invalid versions are treated permissively (returns true) to avoid blocking.
func SatisfiesMin(current, minVersion string) bool {
	if minVersion == "" {
		return true
	}
	if current == "dev" {
		return true
	}

	curParts, curOk := parseSemver(current)
	minParts, minOk := parseSemver(minVersion)

	if !curOk || !minOk {
		// Permissive: don't block on unparseable versions
		return true
	}

	for i := 0; i < 3; i++ {
		if curParts[i] > minParts[i] {
			return true
		}
		if curParts[i] < minParts[i] {
			return false
		}
	}
	return true // equal
}

// parseSemver extracts [major, minor, patch] from a version string.
// Strips leading "v" prefix if present.
func parseSemver(v string) ([3]int, bool) {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return [3]int{}, false
	}

	var result [3]int
	for i, p := range parts {
		// Strip anything after a hyphen (pre-release: 1.0.0-beta)
		if idx := strings.Index(p, "-"); idx >= 0 {
			p = p[:idx]
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return [3]int{}, false
		}
		result[i] = n
	}
	return result, true
}
