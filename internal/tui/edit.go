package tui

import (
	"context"
	"errors"
	"fmt"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

const (
	overwriteButton = "Overwrite"
	deleteButton    = "Delete"
)

// editNote reads the note, then opens it in $EDITOR or in the built in text area.
func (s *shell) editNote(id string) {
	s.async(func(ctx context.Context) (func(), error) {
		note, err := s.readNote(ctx, id)
		return func() { s.startEdit(note) }, err
	})
}

func (s *shell) startEdit(note schema.Note) {
	if s.cfg.Editor == "" {
		s.push(newEditView(s, note))
		return
	}
	content, err := s.externalEdit(note.Content)
	if err != nil {
		s.flashError(err)
		return
	}
	s.save(note, content, false)
}

// noteWrite is one write of an edited note.
type noteWrite struct {
	note    schema.Note
	content string
	force   bool
}

// save writes content over note. Without force the server refuses the write when the
// note changed since note.UpdatedAt, and the conflict dialog asks what to do.
func (s *shell) save(note schema.Note, content string, force bool) {
	if content == note.Content && !force {
		s.flashInfo("no changes to " + note.Path)
		s.leaveEditor()
		return
	}
	write := noteWrite{note: note, content: content, force: force}
	s.async(func(ctx context.Context) (func(), error) { return s.send(ctx, write) })
}

// send writes the note and picks what follows: the conflict dialog on a 409.
func (s *shell) send(ctx context.Context, w noteWrite) (func(), error) {
	err := s.update(ctx, w)
	if isVersionConflict(err) {
		return func() { s.conflict(w) }, nil
	}
	return func() { s.saved(w.note) }, err
}

func (s *shell) update(ctx context.Context, w noteWrite) error {
	update := schema.NoteUpdateRequest{
		Path: w.note.Path, Content: w.content, LastModifiedAt: w.note.UpdatedAt, Force: w.force,
	}
	body, err := s.ports.Notes.UpdateNote(ctx, w.note.ID, update)
	if err != nil {
		return err
	}
	return body.Close()
}

func (s *shell) saved(note schema.Note) {
	s.leaveEditor()
	s.flashInfo("saved " + note.Path)
	s.reload()
}

// leaveEditor pops the text area when it is the current view.
func (s *shell) leaveEditor() {
	if _, ok := s.stack.top().(*editView); ok {
		s.back()
	}
}

// conflict asks whether to overwrite a note that changed on the server. Overwriting
// is never the default: Cancel has the focus.
func (s *shell) conflict(w noteWrite) {
	text := fmt.Sprintf("%s changed on the server after %s, when you opened it.\n\n"+
		"Overwrite replaces it with your version: the edits made on the server since then are lost.",
		plain(w.note.Path), plain(w.note.UpdatedAt))
	s.confirm(text, overwriteButton, func() { s.save(w.note, w.content, true) })
}

// isVersionConflict reports the 409 of a note that changed since it was read, not
// the 409 of a duplicate path, which force would not solve.
func isVersionConflict(err error) bool {
	var classified *failure.Error
	return errors.As(err, &classified) && classified.Kind == failure.Conflict && classified.Code == conflictCode
}

// conflictCode is the error code of the server for a stale lastModifiedAt.
const conflictCode = "conflict"

// deleteNote asks, then deletes the note.
func (s *shell) deleteNote(id, label string) {
	text := fmt.Sprintf("Delete the note %s (%s)?\n\nThe note and its content are removed from the server.", plain(label), plain(id))
	s.confirm(text, deleteButton, func() { s.async(func(ctx context.Context) (func(), error) { return s.remove(ctx, id, label) }) })
}

func (s *shell) remove(ctx context.Context, id, label string) (func(), error) {
	err := s.ports.Notes.DeleteNote(ctx, id)
	return func() { s.deleted(id, label) }, err
}

func (s *shell) deleted(id, label string) {
	if v, ok := s.stack.top().(*noteView); ok && v.id == id {
		s.back()
	}
	s.flashInfo("deleted " + label)
	s.reload()
}
