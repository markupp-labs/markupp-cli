package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// scratchPattern names the temporary file of the editor; the extension lets the
// editor pick Markdown highlighting.
const scratchPattern = "mu-note-*.md"

// externalEdit hands content to $EDITOR in a 0600 temporary file and returns what
// the editor saved. The file is removed on every path, the error paths included.
func (s *shell) externalEdit(content string) (string, error) {
	path, err := writeScratch(s.cfg.TempDir, content)
	if err != nil {
		return "", err
	}
	defer func() { _ = os.Remove(path) }()
	if err := s.suspended(s.editorArgs(path)); err != nil {
		return "", err
	}
	edited, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("reading the edited note from %s: %w", path, err)
	}
	return string(edited), nil
}

// writeScratch creates the temporary file, readable only by the user, with content.
func writeScratch(dir, content string) (string, error) {
	file, err := os.CreateTemp(dir, scratchPattern)
	if err != nil {
		return "", fmt.Errorf("creating the temporary file of the editor in %q: %w", dir, err)
	}
	path := file.Name()
	err = fillScratch(file, content)
	if err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("writing the temporary file %s: %w", path, err)
	}
	return path, nil
}

func fillScratch(file *os.File, content string) error {
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.WriteString(content); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

// editorArgs is the command line of $EDITOR, which may carry flags, then path.
func (s *shell) editorArgs(path string) []string {
	return append(strings.Fields(s.cfg.Editor), path)
}

// suspended runs the editor with the terminal given back to it.
func (s *shell) suspended(argv []string) error {
	var err error
	if !s.app.Suspend(func() { err = s.runEditor(argv) }) {
		return failure.New(failure.Internal, argv[0], "the interface could not hand the terminal to the editor")
	}
	if err != nil {
		return failure.Wrap(failure.Internal, "$EDITOR "+argv[0], fmt.Errorf("the editor failed, the note was not saved: %w", err))
	}
	return nil
}

// execEditor is the real runEditor: the editor gets the terminal of mu tui.
func (s *shell) execEditor(argv []string) error {
	// #nosec G204 -- the command is the $EDITOR of the user, run on purpose.
	cmd := exec.CommandContext(s.ctx, argv[0], argv[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = s.cfg.Terminal.In, s.cfg.Terminal.Out, s.cfg.Terminal.Err
	return cmd.Run()
}
