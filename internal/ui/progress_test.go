package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestNewWaveProgress(t *testing.T) {
	var buf bytes.Buffer
	wp := NewWaveProgress(&buf)

	if wp == nil {
		t.Fatal("NewWaveProgress returned nil")
	}
	if wp.width != 30 {
		t.Errorf("Default width = %d, want 30", wp.width)
	}
	if wp.pattern != "_.-^-._" {
		t.Errorf("Default pattern = %q, want %q", wp.pattern, "_.-^-._")
	}
}

func TestWaveProgressWithWidth(t *testing.T) {
	var buf bytes.Buffer
	wp := NewWaveProgress(&buf, WithWidth(50))

	if wp.width != 50 {
		t.Errorf("Width = %d, want 50", wp.width)
	}
}

func TestWaveProgressWithPattern(t *testing.T) {
	var buf bytes.Buffer
	wp := NewWaveProgress(&buf, WithPattern("~-~"))

	if wp.pattern != "~-~" {
		t.Errorf("Pattern = %q, want %q", wp.pattern, "~-~")
	}
}

func TestWaveProgressWithMessage(t *testing.T) {
	var buf bytes.Buffer
	wp := NewWaveProgress(&buf, WithMessage("Loading..."))

	if wp.message != "Loading..." {
		t.Errorf("Message = %q, want %q", wp.message, "Loading...")
	}
}

func TestWaveProgressUpdate(t *testing.T) {
	var buf bytes.Buffer
	wp := NewWaveProgress(&buf)

	// Update to 50%
	wp.Update(50)

	output := buf.String()
	if !strings.Contains(output, "50%") {
		t.Errorf("Output should contain '50%%', got %q", output)
	}
}

func TestWaveProgressUpdateZero(t *testing.T) {
	var buf bytes.Buffer
	wp := NewWaveProgress(&buf)

	wp.Update(0)

	output := buf.String()
	if !strings.Contains(output, "0%") {
		t.Errorf("Output should contain '0%%', got %q", output)
	}
}

func TestWaveProgressUpdate100(t *testing.T) {
	var buf bytes.Buffer
	wp := NewWaveProgress(&buf)

	wp.Update(100)

	output := buf.String()
	if !strings.Contains(output, "100%") {
		t.Errorf("Output should contain '100%%', got %q", output)
	}
}

func TestWaveProgressFinish(t *testing.T) {
	var buf bytes.Buffer
	wp := NewWaveProgress(&buf)

	wp.Update(50)
	wp.Finish()

	output := buf.String()
	// Should contain newline at the end
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("Finish should end with newline")
	}
}

func TestWaveProgressFinishWithMessage(t *testing.T) {
	var buf bytes.Buffer
	wp := NewWaveProgress(&buf)

	wp.Update(100)
	wp.FinishWithMessage("done!")

	output := buf.String()
	if !strings.Contains(output, "done!") {
		t.Errorf("Output should contain 'done!', got %q", output)
	}
}

func TestWaveProgressWindowSlide(t *testing.T) {
	var buf bytes.Buffer
	wp := NewWaveProgress(&buf, WithWidth(7)) // Small width for testing

	// Get window at different offsets
	win1 := wp.getWindow(0)
	win2 := wp.getWindow(1)

	// Windows should be different (offset by 1)
	if win1 == win2 {
		t.Errorf("Windows at different offsets should differ: %q vs %q", win1, win2)
	}

	// Window length should equal width
	if len(win1) != 7 {
		t.Errorf("Window length = %d, want 7", len(win1))
	}
}

func TestWaveProgressClampsPercent(t *testing.T) {
	var buf bytes.Buffer
	wp := NewWaveProgress(&buf)

	// Should clamp to 0-100
	wp.Update(-10)
	output1 := buf.String()
	if strings.Contains(output1, "-10") {
		t.Error("Negative percent should be clamped to 0")
	}

	buf.Reset()
	wp.Update(150)
	output2 := buf.String()
	if strings.Contains(output2, "150") {
		t.Error("Percent over 100 should be clamped")
	}
}

func TestRunWithProgress(t *testing.T) {
	var buf bytes.Buffer

	completed := false
	err := RunWithProgress(&buf, "Testing", func(update func(int)) error {
		update(50)
		time.Sleep(10 * time.Millisecond)
		update(100)
		completed = true
		return nil
	})

	if err != nil {
		t.Fatalf("RunWithProgress returned error: %v", err)
	}
	if !completed {
		t.Error("Task should have completed")
	}
}

func TestRunWithProgressError(t *testing.T) {
	var buf bytes.Buffer

	err := RunWithProgress(&buf, "Failing", func(update func(int)) error {
		update(25)
		return &testError{msg: "task failed"}
	})

	if err == nil {
		t.Fatal("RunWithProgress should return error")
	}
	if err.Error() != "task failed" {
		t.Errorf("Error = %q, want %q", err.Error(), "task failed")
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestSpinnerStart(t *testing.T) {
	var buf bytes.Buffer
	spinner := NewSpinner(&buf, "Loading")

	spinner.Start()
	time.Sleep(150 * time.Millisecond)
	spinner.Stop()

	output := buf.String()
	// Should have some output
	if len(output) == 0 {
		t.Error("Spinner should produce output")
	}
}

func TestSpinnerMessage(t *testing.T) {
	var buf bytes.Buffer
	spinner := NewSpinner(&buf, "Processing")

	spinner.Start()
	time.Sleep(50 * time.Millisecond)
	spinner.SetMessage("Still processing")
	time.Sleep(50 * time.Millisecond)
	spinner.Stop()

	// Message should have been updated
	if spinner.message != "Still processing" {
		t.Errorf("Message = %q, want %q", spinner.message, "Still processing")
	}
}

func TestSpinnerStopWithResult(t *testing.T) {
	var buf bytes.Buffer
	spinner := NewSpinner(&buf, "Working")

	spinner.Start()
	time.Sleep(50 * time.Millisecond)
	spinner.StopWithSuccess("Complete!")

	output := buf.String()
	if !strings.Contains(output, "Complete!") {
		t.Errorf("Output should contain success message, got %q", output)
	}
}

func TestSpinnerStopWithError(t *testing.T) {
	var buf bytes.Buffer
	spinner := NewSpinner(&buf, "Working")

	spinner.Start()
	time.Sleep(50 * time.Millisecond)
	spinner.StopWithError("Failed!")

	output := buf.String()
	if !strings.Contains(output, "Failed!") {
		t.Errorf("Output should contain error message, got %q", output)
	}
}
