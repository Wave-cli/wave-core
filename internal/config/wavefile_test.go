package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseWavefile(t *testing.T) {
	content := `
[project]
name = "my-app"
version = "1.0.0"
owner = "bouajila"
category = "backend"
tags = ["api", "microservice", "go"]

[flow]
environment = "staging"
port = 3000
watch = true

[test]
coverage = true
threshold = 80
`
	dir := t.TempDir()
	path := filepath.Join(dir, "Wavefile")
	os.WriteFile(path, []byte(content), 0644)

	wf, err := ParseWavefile(path)
	if err != nil {
		t.Fatalf("ParseWavefile failed: %v", err)
	}

	// Project metadata
	if wf.Project.Name != "my-app" {
		t.Errorf("Name = %q, want my-app", wf.Project.Name)
	}
	if wf.Project.Version != "1.0.0" {
		t.Errorf("Version = %q", wf.Project.Version)
	}
	if wf.Project.Owner != "bouajila" {
		t.Errorf("Owner = %q", wf.Project.Owner)
	}
	if wf.Project.Category != "backend" {
		t.Errorf("Category = %q", wf.Project.Category)
	}
	if len(wf.Project.Tags) != 3 {
		t.Fatalf("Tags len = %d, want 3", len(wf.Project.Tags))
	}
	if wf.Project.Tags[0] != "api" {
		t.Errorf("Tags[0] = %q", wf.Project.Tags[0])
	}

	// Plugin sections
	flowSection, ok := wf.Sections["flow"]
	if !ok {
		t.Fatal("flow section missing")
	}
	if env, ok := flowSection["environment"]; !ok || env != "staging" {
		t.Errorf("flow.environment = %v", flowSection["environment"])
	}

	testSection, ok := wf.Sections["test"]
	if !ok {
		t.Fatal("test section missing")
	}
	if cov, ok := testSection["coverage"]; !ok || cov != true {
		t.Errorf("test.coverage = %v", testSection["coverage"])
	}
}

func TestParseWavefileMinimal(t *testing.T) {
	content := `
[project]
name = "bare"
version = "0.1.0"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "Wavefile")
	os.WriteFile(path, []byte(content), 0644)

	wf, err := ParseWavefile(path)
	if err != nil {
		t.Fatalf("ParseWavefile failed: %v", err)
	}
	if wf.Project.Name != "bare" {
		t.Errorf("Name = %q", wf.Project.Name)
	}
	if wf.Sections == nil {
		t.Error("Sections should be initialized even if empty")
	}
	if len(wf.Sections) != 0 {
		t.Errorf("Sections len = %d, want 0", len(wf.Sections))
	}
}

func TestParseWavefileMissing(t *testing.T) {
	_, err := ParseWavefile("/nonexistent/Wavefile")
	if err == nil {
		t.Error("Should fail for missing file")
	}
}

func TestParseWavefileInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Wavefile")
	os.WriteFile(path, []byte("[[[broken toml"), 0644)

	_, err := ParseWavefile(path)
	if err == nil {
		t.Error("Should fail for invalid TOML")
	}
}

func TestDiscoverWavefile(t *testing.T) {
	// Create nested dirs: root/sub1/sub2
	// Place Wavefile in root
	root := t.TempDir()
	sub1 := filepath.Join(root, "sub1")
	sub2 := filepath.Join(sub1, "sub2")
	os.MkdirAll(sub2, 0755)

	wavefilePath := filepath.Join(root, "Wavefile")
	os.WriteFile(wavefilePath, []byte("[project]\nname=\"test\"\nversion=\"0.1.0\"\n"), 0644)

	// Discover from deepest dir should find it
	found, err := DiscoverWavefile(sub2)
	if err != nil {
		t.Fatalf("DiscoverWavefile failed: %v", err)
	}
	if found != wavefilePath {
		t.Errorf("found %q, want %q", found, wavefilePath)
	}
}

func TestDiscoverWavefileNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := DiscoverWavefile(dir)
	if err == nil {
		t.Error("Should return error when no Wavefile exists up the tree")
	}
}

func TestDiscoverWavefileInCurrentDir(t *testing.T) {
	dir := t.TempDir()
	wavefilePath := filepath.Join(dir, "Wavefile")
	os.WriteFile(wavefilePath, []byte("[project]\nname=\"here\"\nversion=\"1.0.0\"\n"), 0644)

	found, err := DiscoverWavefile(dir)
	if err != nil {
		t.Fatalf("DiscoverWavefile failed: %v", err)
	}
	if found != wavefilePath {
		t.Errorf("found %q, want %q", found, wavefilePath)
	}
}
