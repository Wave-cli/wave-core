package errors

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LogEntry is the JSONL structure written to log files.
type LogEntry struct {
	Timestamp string   `json:"ts"`
	Plugin    string   `json:"plugin"`
	Code      string   `json:"code"`
	Message   string   `json:"message"`
	Args      []string `json:"args"`
	Cwd       string   `json:"cwd,omitempty"`
}

// LogError appends a structured error to the daily log file in logsDir.
func LogError(logsDir, pluginName string, pe *PluginError, args []string) error {
	filename := time.Now().Format("2006-01-02") + ".log"
	logPath := filepath.Join(logsDir, filename)

	cwd, _ := os.Getwd()

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Plugin:    pluginName,
		Code:      pe.Code,
		Message:   pe.Message,
		Args:      args,
		Cwd:       cwd,
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(entry); err != nil {
		return fmt.Errorf("writing log entry: %w", err)
	}

	return nil
}

// jsonUnmarshal is a package-level wrapper around encoding/json.Unmarshal.
func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
