package tui

import (
	"os"
)

// scratchPattern names the temporary file of the editor; the extension lets the
// editor pick Markdown highlighting.
const scratchPattern = "mu-note-*.md"

// externalEdit hands content to $EDITOR in a 0600 temporary file and returns what
// the editor saved. The file is removed on every path, the error paths included.
func (s *shell) externalEdit(content string) (string, error) { return *new(string), *new(error) }

// writeScratch creates the temporary file, readable only by the user, with content.
func writeScratch(dir, content string) (string, error) { return *new(string), *new(error) }

func fillScratch(file *os.File, content string) error { return *new(error) }

// editorArgs is the command line of $EDITOR, which may carry flags, then path.
func (s *shell) editorArgs(path string) []string { return *new([]string) }

// suspended runs the editor with the terminal given back to it.
func (s *shell) suspended(argv []string) error { return *new(error) }

// execEditor is the real runEditor: the editor gets the terminal of mu tui.
func (s *shell) execEditor(argv []string) error { return *new(error) }
