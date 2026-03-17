package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wave-cli/wave-core/internal/config"
)

func TestInitCreatesWavefile(t *testing.T) {
	root := t.TempDir()
	setTestHome(t, root)

	// Create a project directory
	projectDir := filepath.Join(root, "myproject")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Change to project directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	resetCmdState()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"init"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check Wavefile was created
	wavefilePath := filepath.Join(projectDir, "Wavefile")
	data, err := os.ReadFile(wavefilePath)
	if err != nil {
		t.Fatalf("Wavefile not created: %v", err)
	}

	// Check project name is directory name
	if !strings.Contains(string(data), `name = "myproject"`) {
		t.Errorf("Wavefile should contain project name 'myproject', got:\n%s", data)
	}
}

func TestInitWithProjectNameArg(t *testing.T) {
	root := t.TempDir()
	setTestHome(t, root)

	// Create a project directory
	projectDir := filepath.Join(root, "somefolder")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Change to project directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	resetCmdState()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"init", "custom-project-name"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check Wavefile was created with custom name
	wavefilePath := filepath.Join(projectDir, "Wavefile")
	data, err := os.ReadFile(wavefilePath)
	if err != nil {
		t.Fatalf("Wavefile not created: %v", err)
	}

	// Check project name is the custom name, not directory name
	if !strings.Contains(string(data), `name = "custom-project-name"`) {
		t.Errorf("Wavefile should contain project name 'custom-project-name', got:\n%s", data)
	}
}

func TestInitAddsProjectToGlobalConfig(t *testing.T) {
	root := t.TempDir()
	setTestHome(t, root)

	// Create a project directory
	projectDir := filepath.Join(root, "testproject")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Change to project directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	resetCmdState()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"init"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check global config was updated with project folder
	configPath := filepath.Join(root, ".wave", "config")
	gc, err := config.ParseGlobalConfig(configPath)
	if err != nil {
		t.Fatalf("ParseGlobalConfig failed: %v", err)
	}

	found := false
	for _, folder := range gc.Projects.Folders {
		if folder == projectDir {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Project folder %q not found in global config folders: %v", projectDir, gc.Projects.Folders)
	}
}

func TestInitUsesDefaultOrgFromConfig(t *testing.T) {
	root := t.TempDir()
	setTestHome(t, root)

	// Create global config with default org
	waveDir := filepath.Join(root, ".wave")
	if err := os.MkdirAll(waveDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	gc := config.DefaultGlobalConfig(root)
	gc.User.Org = "wave-org"
	gc.User.Name = "testuser"
	configPath := filepath.Join(waveDir, "config")
	if err := config.WriteGlobalConfig(configPath, gc); err != nil {
		t.Fatalf("WriteGlobalConfig failed: %v", err)
	}

	// Create a project directory
	projectDir := filepath.Join(root, "orgproject")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Change to project directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	resetCmdState()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"init"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check Wavefile uses default org as owner
	wavefilePath := filepath.Join(projectDir, "Wavefile")
	data, err := os.ReadFile(wavefilePath)
	if err != nil {
		t.Fatalf("Wavefile not created: %v", err)
	}

	if !strings.Contains(string(data), `owner = "wave-org"`) {
		t.Errorf("Wavefile should contain owner 'wave-org', got:\n%s", data)
	}
}

func TestInitDoesNotDuplicateProjectFolder(t *testing.T) {
	root := t.TempDir()
	setTestHome(t, root)

	// Create a project directory
	projectDir := filepath.Join(root, "duplicatetest")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Change to project directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	// Pre-populate global config with the project folder
	waveDir := filepath.Join(root, ".wave")
	if err := os.MkdirAll(waveDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	gc := config.DefaultGlobalConfig(root)
	gc.Projects.Folders = []string{projectDir}
	configPath := filepath.Join(waveDir, "config")
	if err := config.WriteGlobalConfig(configPath, gc); err != nil {
		t.Fatalf("WriteGlobalConfig failed: %v", err)
	}

	// Remove Wavefile if it exists (to allow init to run)
	os.Remove(filepath.Join(projectDir, "Wavefile"))

	resetCmdState()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"init"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check global config doesn't have duplicate entries
	gc, err := config.ParseGlobalConfig(configPath)
	if err != nil {
		t.Fatalf("ParseGlobalConfig failed: %v", err)
	}

	count := 0
	for _, folder := range gc.Projects.Folders {
		if folder == projectDir {
			count++
		}
	}

	if count != 1 {
		t.Errorf("Project folder appears %d times in config, want 1: %v", count, gc.Projects.Folders)
	}
}
