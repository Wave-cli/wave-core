package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewPrinter(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, LevelNormal)
	if p == nil {
		t.Fatal("NewPrinter should not return nil")
	}
}

func TestInfoPrintsAtNormalLevel(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, LevelNormal)
	p.Info("hello %s", "world")
	out := buf.String()
	if !strings.Contains(out, "hello world") {
		t.Errorf("Info should print at normal level, got %q", out)
	}
}

func TestInfoSuppressedAtQuietLevel(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, LevelQuiet)
	p.Info("should not appear")
	if buf.Len() != 0 {
		t.Errorf("Info should be suppressed at quiet level, got %q", buf.String())
	}
}

func TestVerbosePrintsAtVerboseLevel(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, LevelVerbose)
	p.Verbose("detail: %d", 42)
	out := buf.String()
	if !strings.Contains(out, "detail: 42") {
		t.Errorf("Verbose should print at verbose level, got %q", out)
	}
}

func TestVerboseSuppressedAtNormalLevel(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, LevelNormal)
	p.Verbose("should not appear")
	if buf.Len() != 0 {
		t.Errorf("Verbose should be suppressed at normal level, got %q", buf.String())
	}
}

func TestDebugPrintsAtDebugLevel(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, LevelDebug)
	p.Debug("trace: %v", true)
	out := buf.String()
	if !strings.Contains(out, "trace: true") {
		t.Errorf("Debug should print at debug level, got %q", out)
	}
}

func TestDebugSuppressedAtVerboseLevel(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, LevelVerbose)
	p.Debug("should not appear")
	if buf.Len() != 0 {
		t.Errorf("Debug should be suppressed at verbose level, got %q", buf.String())
	}
}

func TestErrorAlwaysPrints(t *testing.T) {
	levels := []Level{LevelQuiet, LevelNormal, LevelVerbose, LevelDebug}
	for _, level := range levels {
		var buf bytes.Buffer
		p := NewPrinter(&buf, level)
		p.Error("failure: %s", "boom")
		out := buf.String()
		if !strings.Contains(out, "failure: boom") {
			t.Errorf("Error should always print at level %d, got %q", level, out)
		}
	}
}

func TestWarnPrintsAtNormalAndAbove(t *testing.T) {
	// Should print at normal, verbose, debug
	for _, level := range []Level{LevelNormal, LevelVerbose, LevelDebug} {
		var buf bytes.Buffer
		p := NewPrinter(&buf, level)
		p.Warn("careful: %s", "hot")
		if !strings.Contains(buf.String(), "careful: hot") {
			t.Errorf("Warn should print at level %d", level)
		}
	}
	// Should NOT print at quiet
	var buf bytes.Buffer
	p := NewPrinter(&buf, LevelQuiet)
	p.Warn("careful")
	if buf.Len() != 0 {
		t.Errorf("Warn should be suppressed at quiet level, got %q", buf.String())
	}
}

func TestSuccessPrintsAtNormalLevel(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, LevelNormal)
	p.Success("done: %s", "install")
	out := buf.String()
	if !strings.Contains(out, "done: install") {
		t.Errorf("Success should print at normal level, got %q", out)
	}
}

func TestLevelOrdering(t *testing.T) {
	if LevelQuiet >= LevelNormal {
		t.Error("LevelQuiet should be less than LevelNormal")
	}
	if LevelNormal >= LevelVerbose {
		t.Error("LevelNormal should be less than LevelVerbose")
	}
	if LevelVerbose >= LevelDebug {
		t.Error("LevelVerbose should be less than LevelDebug")
	}
}
