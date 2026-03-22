package version

import (
	"testing"
)

func TestDefaultVersion(t *testing.T) {
	// When no ldflags are set, version should be "dev"
	info := Get()
	if info.Version == "" {
		t.Error("Version should never be empty, expected 'dev' as default")
	}
}

func TestGetReturnsConsistentInfo(t *testing.T) {
	a := Get()
	b := Get()
	if a.Version != b.Version || a.Commit != b.Commit || a.Date != b.Date {
		t.Error("Get() should return consistent values across calls")
	}
}

func TestStringFormat(t *testing.T) {
	info := Get()
	s := info.String()
	if s == "" {
		t.Error("String() should not be empty")
	}
	// Should contain the version somewhere
	if !containsSubstring(s, info.Version) {
		t.Errorf("String() output %q should contain version %q", s, info.Version)
	}
}

func TestFullFormat(t *testing.T) {
	info := Get()
	full := info.Full()
	if full == "" {
		t.Error("Full() should not be empty")
	}
	// Full should contain version, commit, date
	if !containsSubstring(full, info.Version) {
		t.Errorf("Full() %q should contain version %q", full, info.Version)
	}
	if !containsSubstring(full, info.Commit) {
		t.Errorf("Full() %q should contain commit %q", full, info.Commit)
	}
	if !containsSubstring(full, info.Date) {
		t.Errorf("Full() %q should contain date %q", full, info.Date)
	}
}

func TestSatisfiesMinVersion(t *testing.T) {
	tests := []struct {
		name       string
		current    string
		minVersion string
		want       bool
	}{
		{"same version", "1.0.0", "1.0.0", true},
		{"higher major", "2.0.0", "1.0.0", true},
		{"higher minor", "1.2.0", "1.1.0", true},
		{"higher patch", "1.0.1", "1.0.0", true},
		{"lower major", "0.9.0", "1.0.0", false},
		{"lower minor", "1.0.0", "1.1.0", false},
		{"lower patch", "1.0.0", "1.0.1", false},
		{"dev satisfies anything", "dev", "1.0.0", true},
		{"empty min always satisfied", "1.0.0", "", true},
		{"complex versions", "2.10.3", "2.9.15", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SatisfiesMin(tt.current, tt.minVersion)
			if got != tt.want {
				t.Errorf("SatisfiesMin(%q, %q) = %v, want %v", tt.current, tt.minVersion, got, tt.want)
			}
		})
	}
}

func TestSatisfiesMinInvalidVersions(t *testing.T) {
	// Invalid versions should not panic, should return true (permissive)
	tests := []struct {
		name    string
		current string
		min     string
	}{
		{"garbage current", "not-a-version", "1.0.0"},
		{"garbage min", "1.0.0", "not-a-version"},
		{"both garbage", "abc", "xyz"},
		{"v-prefix current", "v1.0.0", "1.0.0"},
		{"v-prefix min", "1.0.0", "v1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			_ = SatisfiesMin(tt.current, tt.min)
		})
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
