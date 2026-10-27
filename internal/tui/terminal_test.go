package tui

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func TestRunRefusesAStdoutThatIsNotATerminal(t *testing.T) {
	log, err := os.Create(filepath.Join(t.TempDir(), "log"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = log.Close() }()
	err = Run(context.Background(), Config{Terminal: Terminal{In: strings.NewReader(""), Out: log, Err: &bytes.Buffer{}}})
	if failure.KindOf(err) != failure.Usage || !strings.Contains(err.Error(), "stdout") {
		t.Fatalf("Run with stdout on a file = %v, want a usage error naming stdout", err)
	}
	if info, _ := log.Stat(); info.Size() != 0 {
		t.Errorf("Run wrote %d bytes to the redirected stdout, want none", info.Size())
	}
}

func TestCheckTerminalAcceptsWritersThatAreNotFiles(t *testing.T) {
	if err := checkTerminal(Terminal{In: strings.NewReader(""), Out: &bytes.Buffer{}}); err != nil {
		t.Errorf("checkTerminal with in memory streams = %v, want nil", err)
	}
}
