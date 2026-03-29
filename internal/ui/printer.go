// Package ui provides consistent terminal output with verbosity levels.
package ui

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
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

// ANSI color codes
const (
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
)

// Error always prints (all levels including quiet).
func (p *Printer) Error(format string, args ...any) {
	fmt.Fprintf(p.out, colorRed+"ERROR:"+colorReset+" "+format+"\n", args...)
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

// WaveProgress displays an animated wave progress bar.
// The wave pattern scrolls horizontally as progress updates.
type WaveProgress struct {
	out     io.Writer
	width   int
	pattern string
	message string
	tape    string
	offset  int
	mu      sync.Mutex
}

// WaveProgressOption configures a WaveProgress instance.
type WaveProgressOption func(*WaveProgress)

// WithWidth sets the visible width of the progress bar.
func WithWidth(w int) WaveProgressOption {
	return func(wp *WaveProgress) {
		if w > 0 {
			wp.width = w
		}
	}
}

// WithPattern sets the wave pattern (e.g., "_.-^-._").
func WithPattern(p string) WaveProgressOption {
	return func(wp *WaveProgress) {
		if p != "" {
			wp.pattern = p
		}
	}
}

// WithMessage sets the message displayed alongside the progress.
func WithMessage(m string) WaveProgressOption {
	return func(wp *WaveProgress) {
		wp.message = m
	}
}

// NewWaveProgress creates a new animated wave progress bar.
func NewWaveProgress(out io.Writer, opts ...WaveProgressOption) *WaveProgress {
	wp := &WaveProgress{
		out:     out,
		width:   30,
		pattern: "_.-^-._",
	}

	for _, opt := range opts {
		opt(wp)
	}

	// Create a long "tape" by repeating the pattern
	wp.tape = strings.Repeat(wp.pattern, 20)

	return wp
}

// getWindow returns the visible portion of the wave at the current offset.
func (wp *WaveProgress) getWindow(offset int) string {
	patternLen := len(wp.pattern)
	if patternLen == 0 {
		return strings.Repeat(" ", wp.width)
	}

	// Ensure offset wraps correctly
	adjustedOffset := offset % patternLen

	// Make sure we have enough tape
	if adjustedOffset+wp.width > len(wp.tape) {
		wp.tape = strings.Repeat(wp.pattern, 20)
	}

	return wp.tape[adjustedOffset : adjustedOffset+wp.width]
}

// Update displays the progress bar at the given percentage (0-100).
func (wp *WaveProgress) Update(percent int) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	// Clamp percent to 0-100
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	// Calculate offset for wave animation
	wp.offset = percent % len(wp.pattern)

	window := wp.getWindow(wp.offset)

	// \r = start of line, \033[K = clear line, \033[36m = Cyan, \033[33m = Yellow
	if wp.message != "" {
		fmt.Fprintf(wp.out, "\r\033[K%s \033[36m(%s)\033[0m \033[33m%3d%%\033[0m", wp.message, window, percent)
	} else {
		fmt.Fprintf(wp.out, "\r\033[K\033[36m(%s)\033[0m \033[33m%3d%%\033[0m", window, percent)
	}
}

// Finish completes the progress bar and moves to a new line.
func (wp *WaveProgress) Finish() {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	fmt.Fprintln(wp.out)
}

// FinishWithMessage completes the progress bar with a final message.
func (wp *WaveProgress) FinishWithMessage(msg string) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	// Clear line and print success message
	fmt.Fprintf(wp.out, "\r\033[K\033[32m%s\033[0m\n", msg)
}

// RunWithProgress executes a task with an animated progress bar.
// The task function receives an update callback to report progress (0-100).
func RunWithProgress(out io.Writer, message string, task func(update func(int)) error) error {
	wp := NewWaveProgress(out, WithMessage(message))

	wp.Update(0)

	err := task(func(percent int) {
		wp.Update(percent)
	})

	if err != nil {
		wp.Finish()
		return err
	}

	wp.FinishWithMessage("done!")
	return nil
}

// Spinner displays an animated spinner for indeterminate operations.
type Spinner struct {
	out      io.Writer
	message  string
	frames   []string
	interval time.Duration
	stopCh   chan struct{}
	doneCh   chan struct{}
	mu       sync.Mutex
	running  bool
}

// NewSpinner creates a new spinner with the given message.
// Uses ASCII wave animation inspired by ocean waves.
func NewSpinner(out io.Writer, message string) *Spinner {
	return &Spinner{
		out:     out,
		message: message,
		frames: []string{
			"~~~~~~~",
			"≈~~~~~~",
			"~≈~~~~~",
			"~~≈~~~~",
			"~~~≈~~~",
			"~~~~≈~~",
			"~~~~~≈~",
			"~~~~~~≈",
			"~~~~~≈~",
			"~~~~≈~~",
			"~~~≈~~~",
			"~~≈~~~~",
			"~≈~~~~~",
			"≈~~~~~~",
		},
		interval: 100 * time.Millisecond,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start begins the spinner animation.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	s.mu.Unlock()

	go func() {
		defer close(s.doneCh)
		frame := 0
		for {
			select {
			case <-s.stopCh:
				return
			default:
				s.mu.Lock()
				msg := s.message
				s.mu.Unlock()

				fmt.Fprintf(s.out, "\r\033[K\033[36m(%s)\033[0m %s", s.frames[frame], msg)
				frame = (frame + 1) % len(s.frames)
				time.Sleep(s.interval)
			}
		}
	}()
}

// Stop halts the spinner animation.
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopCh)
	<-s.doneCh
	fmt.Fprintln(s.out)
}

// SetMessage updates the spinner message.
func (s *Spinner) SetMessage(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = msg
}

// StopWithSuccess stops the spinner and displays a success message.
func (s *Spinner) StopWithSuccess(msg string) {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		fmt.Fprintf(s.out, "\r\033[K\033[32m%s\033[0m\n", msg)
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopCh)
	<-s.doneCh
	fmt.Fprintf(s.out, "\r\033[K\033[32m%s\033[0m\n", msg)
}

// StopWithError stops the spinner and displays an error message.
func (s *Spinner) StopWithError(msg string) {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		fmt.Fprintf(s.out, "\r\033[K\033[31m%s\033[0m\n", msg)
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopCh)
	<-s.doneCh
	fmt.Fprintf(s.out, "\r\033[K\033[31m%s\033[0m\n", msg)
}
