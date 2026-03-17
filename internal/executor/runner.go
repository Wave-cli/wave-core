// Package executor handles plugin execution: fork/exec with stdin config,
// environment variables, and stderr parsing.
package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	waveerrors "github.com/wave-cli/wave-core/internal/errors"
)

// execCommand is a variable so tests can verify command construction.
var execCommand = exec.Command

// Result holds the outcome of a plugin execution.
type Result struct {
	ExitCode    int
	Stdout      string
	Stderr      string
	PluginError *waveerrors.PluginError
}

// Execute runs a plugin binary with the given arguments, config section
// (as stdin JSON), and environment variables.
func Execute(binPath string, args []string, section map[string]any, pluginName, pluginVersion, projectRoot string) (*Result, error) {
	// Verify binary exists
	if _, err := os.Stat(binPath); err != nil {
		return nil, fmt.Errorf("plugin binary not found: %w", err)
	}

	// Build stdin data
	stdinData, err := BuildStdin(section)
	if err != nil {
		return nil, fmt.Errorf("building stdin: %w", err)
	}

	// Build environment
	pluginDir := filepath.Dir(filepath.Dir(binPath)) // bin/<name> -> version dir
	env := BuildEnv(pluginName, pluginVersion, pluginDir, projectRoot)

	// Create command
	cmd := execCommand(binPath, args...)
	cmd.Env = env
	cmd.Stdin = bytes.NewReader(stdinData)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	runErr := cmd.Run()

	result := &Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	// Get exit code
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}

	// Parse stderr for structured wave errors
	if runErr != nil {
		if pe := waveerrors.ParseStderr(stderr.Bytes()); pe != nil {
			result.PluginError = pe
			return result, nil // structured error, not a Go error
		}
		// Non-zero exit without structured error
		return result, nil
	}

	return result, nil
}

// BuildEnv creates the environment variables for a plugin execution.
// It inherits the current OS environment and adds WAVE_* variables.
func BuildEnv(pluginName, pluginVersion, pluginDir, projectRoot string) []string {
	env := os.Environ()

	waveVars := map[string]string{
		"WAVE_PLUGIN_NAME":    pluginName,
		"WAVE_PLUGIN_VERSION": pluginVersion,
		"WAVE_PLUGIN_DIR":     pluginDir,
		"WAVE_PLUGIN_ASSETS":  filepath.Join(pluginDir, "assets"),
		"WAVE_PROJECT_ROOT":   projectRoot,
	}

	for k, v := range waveVars {
		env = append(env, k+"="+v)
	}

	return env
}

// BuildStdin serializes a plugin config section as JSON for stdin.
// Returns "{}" for nil or empty sections.
func BuildStdin(section map[string]any) ([]byte, error) {
	if section == nil {
		section = map[string]any{}
	}
	return json.Marshal(section)
}

// StreamExecute runs a plugin binary and streams output directly to the
// provided writers (typically os.Stdout and os.Stderr).
func StreamExecute(binPath string, args []string, section map[string]any, pluginName, pluginVersion, projectRoot string, stdout, stderr io.Writer) (*Result, error) {
	if _, err := os.Stat(binPath); err != nil {
		return nil, fmt.Errorf("plugin binary not found: %w", err)
	}

	stdinData, err := BuildStdin(section)
	if err != nil {
		return nil, fmt.Errorf("building stdin: %w", err)
	}

	pluginDir := filepath.Dir(filepath.Dir(binPath))
	env := BuildEnv(pluginName, pluginVersion, pluginDir, projectRoot)

	cmd := execCommand(binPath, args...)
	cmd.Env = env
	cmd.Stdin = bytes.NewReader(stdinData)

	// Capture stderr for error parsing while also streaming
	var stderrBuf bytes.Buffer
	cmd.Stdout = stdout
	cmd.Stderr = io.MultiWriter(stderr, &stderrBuf)

	runErr := cmd.Run()

	result := &Result{
		Stderr: stderrBuf.String(),
	}

	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}

	if runErr != nil {
		if pe := waveerrors.ParseStderr(stderrBuf.Bytes()); pe != nil {
			result.PluginError = pe
			return result, nil
		}
	}

	return result, nil
}

// LookupPlugin finds a plugin name in the installed plugins map and returns
// its full name (org/name).
func LookupPlugin(shortName string, plugins map[string]string) (string, string, bool) {
	for fullName, version := range plugins {
		parts := strings.SplitN(fullName, "/", 2)
		if len(parts) == 2 && parts[1] == shortName {
			return fullName, version, true
		}
	}
	return "", "", false
}
