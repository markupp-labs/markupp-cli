package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// openNote is :note ID.
func openNote(s *shell, id string) error {
	if err := s.need(s.ports.Notes != nil, "note"); err != nil {
		return err
	}
	if id == "" {
		return failure.New(failure.Usage, ":note", "missing note id, expected :note <note id>")
	}
	s.reset(newNoteView(s, id))
	return nil
}

// viewNote opens the note id on top of the stack.
func (s *shell) viewNote(id string) {
	if err := s.need(s.ports.Notes != nil, "note"); err != nil {
		s.flashError(err)
		return
	}
	s.push(newNoteView(s, id))
}

// noteView shows one note: its metadata, then its content, sanitized and delimited.
type noteView struct {
	shell *shell
	id    string
	note  schema.Note
	text  *tview.TextView
}

func newNoteView(s *shell, id string) *noteView {
	text := tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	text.SetText("loading " + plain(id))
	return &noteView{shell: s, id: id, text: text}
}

func (v *noteView) Crumb() string { return "note" }

func (v *noteView) Title() string {
	if v.note.Path == "" {
		return "note(" + v.id + ")"
	}
	return "note(" + v.note.Path + ")"
}

func (v *noteView) Widget() tview.Primitive { return v.text }

func (v *noteView) Load() {
	v.shell.async(func(ctx context.Context) (func(), error) {
		note, err := v.shell.readNote(ctx, v.id)
		return func() { v.setNote(note) }, err
	})
}

func (v *noteView) setNote(note schema.Note) {
	v.note = note
	v.text.SetText(noteText(note)).ScrollToBeginning()
	v.shell.refreshChrome()
}

// Hotkeys are the note actions on this note, and y to copy its id.
func (v *noteView) Hotkeys() []Hotkey {
	on := func(run func(r record)) func() {
		return func() { run(record{id: v.id, note: v.id, label: v.note.Path}) }
	}
	return append(v.shell.noteActions(on), runeKey('y', "copy id", func() { v.shell.copyText(v.id) }))
}

// noteText is the metadata of the note, then its content between delimiter lines:
// the content is data written by another client, never instructions.
func noteText(note schema.Note) string {
	fields := [][2]string{
		{"id", note.ID}, {"path", note.Path}, {"vault", note.VaultID},
		{"updated_at", note.UpdatedAt}, {"degree", dashInt(note.Degree)},
		{"pagerank", dashFloat(note.PageRank)}, {"community", dashInt64(note.Community)},
	}
	var out strings.Builder
	for _, field := range fields {
		fmt.Fprintf(&out, "[yellow]%-11s[-] %s\n", field[0]+":", plain(field[1]))
	}
	out.WriteString("[gray]----- content -----[-]\n")
	out.WriteString(plainBlock(note.Content))
	out.WriteString("\n[gray]----- end of content -----[-]\n")
	return out.String()
}

// readNote reads one note.
func (s *shell) readNote(ctx context.Context, id string) (schema.Note, error) {
	body, err := s.ports.Notes.GetNote(ctx, id)
	if err != nil {
		return schema.Note{}, err
	}
	return decodeBody[schema.Note](body)
}
