package tui

import (
	"github.com/gdamore/tcell/v2"
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

func newEditView(s *shell, note schema.Note) *editView {
	area := tview.NewTextArea().SetText(note.Content, false)
	return &editView{shell: s, note: note, area: area}
}

func (v *editView) Crumb() string { return "edit" }

func (v *editView) Title() string { return "edit(" + v.note.Path + ") ctrl-s saves, esc discards" }

func (v *editView) Widget() tview.Primitive { return v.area }

// Load does nothing: the text area holds the note as it was read.
func (v *editView) Load() {}

// Typing is true: printable keys type into the note.
func (v *editView) Typing() bool { return true }

func (v *editView) Hotkeys() []Hotkey {
	return []Hotkey{
		ctrlKey(tcell.KeyCtrlS, "save", func() { v.shell.save(v.note, v.area.GetText(), false) }),
		ctrlKey(tcell.KeyEscape, "discard", v.discard),
	}
}

func (v *editView) discard() {
	v.shell.flashInfo("edit of " + v.note.Path + " discarded")
	v.shell.back()
}
