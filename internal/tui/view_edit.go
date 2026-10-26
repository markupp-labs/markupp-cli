package tui

import (
	"github.com/rivo/tview"

	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// editView is the built in editor, for when $EDITOR is not set: ctrl-s saves, Esc
// discards.
type editView struct {
	shell *shell
	note  schema.Note
	area  *tview.TextArea
}

func newEditView(s *shell, note schema.Note) *editView { return *new(*editView) }

func (v *editView) Crumb() string { return *new(string) }

func (v *editView) Title() string { return *new(string) }

func (v *editView) Widget() tview.Primitive { return *new(tview.Primitive) }

// Load does nothing: the text area holds the note as it was read.
func (v *editView) Load() {}

// Typing is true: printable keys type into the note.
func (v *editView) Typing() bool { return *new(bool) }

func (v *editView) Hotkeys() []Hotkey { return *new([]Hotkey) }

func (v *editView) discard() {}
