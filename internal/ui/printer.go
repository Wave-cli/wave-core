// Package ui provides consistent terminal output with verbosity levels.
package ui

import (
	"fmt"
	"io"
)

// Level controls which messages are printed.
type Level int

const (
	LevelQuiet   Level = 0
	LevelNormal  Level = 1
	LevelVerbose Level = 2
	LevelDebug   Level = 3
)

// Printer writes formatted output respecting the configured verbosity level.
type Printer struct {
	out   io.Writer
	level Level
}

// NewPrinter creates a Printer that writes to out at the given verbosity level.
func NewPrinter(out io.Writer, level Level) *Printer {
	return &Printer{out: out, level: level}
}

// Error always prints (all levels including quiet).
func (p *Printer) Error(format string, args ...any) {
	fmt.Fprintf(p.out, "ERROR: "+format+"\n", args...)
}

// Warn prints at normal level and above.
func (p *Printer) Warn(format string, args ...any) {
	if p.level >= LevelNormal {
		fmt.Fprintf(p.out, "WARN:  "+format+"\n", args...)
	}
}

// Info prints at normal level and above.
func (p *Printer) Info(format string, args ...any) {
	if p.level >= LevelNormal {
		fmt.Fprintf(p.out, format+"\n", args...)
	}
}

// Success prints at normal level and above.
func (p *Printer) Success(format string, args ...any) {
	if p.level >= LevelNormal {
		fmt.Fprintf(p.out, "OK:    "+format+"\n", args...)
	}
}

// Verbose prints at verbose level and above.
func (p *Printer) Verbose(format string, args ...any) {
	if p.level >= LevelVerbose {
		fmt.Fprintf(p.out, format+"\n", args...)
	}
}

// Debug prints at debug level only.
func (p *Printer) Debug(format string, args ...any) {
	if p.level >= LevelDebug {
		fmt.Fprintf(p.out, "DEBUG: "+format+"\n", args...)
	}
}
