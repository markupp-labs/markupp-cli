package output

import (
	"bytes"
	"errors"
	"testing"
)

func TestRawWritesTheTextUnchanged(t *testing.T) {
	var out bytes.Buffer
	const content = "# A\r\n\ttab \x1b[31mred\nno final newline"
	if err := NewPrinter(&out).Raw(content); err != nil {
		t.Fatalf("Raw: %v", err)
	}
	if out.String() != content {
		t.Errorf("Raw wrote %q, want %q", out.String(), content)
	}
}

func TestRawReportsAWriteError(t *testing.T) {
	if err := NewPrinter(failingWriter{}).Raw("x"); err == nil {
		t.Error("Raw on a failing writer returned nil")
	}
}

// failingWriter refuses every write, like a closed stdout.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errClosedStdout }

var errClosedStdout = errors.New("stdout closed")
