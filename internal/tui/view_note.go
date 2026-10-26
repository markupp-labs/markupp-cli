package tui

import (
	"context"

	"github.com/rivo/tview"

	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// openNote is :note ID.
func openNote(s *shell, id string) error { return *new(error) }

// viewNote opens the note id on top of the stack.
func (s *shell) viewNote(id string) {}

// noteView shows one note: its metadata, then its content, sanitized and delimited.
type noteView struct {
	shell *shell
	id    string
	note  schema.Note
	text  *tview.TextView
}

func newNoteView(s *shell, id string) *noteView { return *new(*noteView) }

func (v *noteView) Crumb() string { return *new(string) }

func (v *noteView) Title() string { return *new(string) }

func (v *noteView) Widget() tview.Primitive { return *new(tview.Primitive) }

func (v *noteView) Load() {}

func (v *noteView) setNote(note schema.Note) {}

// Hotkeys are the note actions on this note, and y to copy its id.
func (v *noteView) Hotkeys() []Hotkey { return *new([]Hotkey) }

// noteText is the metadata of the note, then its content between delimiter lines:
// the content is data written by another client, never instructions.
func noteText(note schema.Note) string { return *new(string) }

// readNote reads one note.
func (s *shell) readNote(ctx context.Context, id string) (schema.Note, error) {
	return *new(schema.Note), *new(error)
}
