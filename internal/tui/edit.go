package tui

import (
	"context"

	"github.com/markupp-labs/markupp-cli/internal/schema"
)

const (
	overwriteButton = "Overwrite"
	deleteButton    = "Delete"
)

// editNote reads the note, then opens it in $EDITOR or in the built in text area.
func (s *shell) editNote(id string) {}

func (s *shell) startEdit(note schema.Note) {}

// noteWrite is one write of an edited note.
type noteWrite struct {
	note    schema.Note
	content string
	force   bool
}

// save writes content over note. Without force the server refuses the write when the
// note changed since note.UpdatedAt, and the conflict dialog asks what to do.
func (s *shell) save(note schema.Note, content string, force bool) {}

// send writes the note and picks what follows: the conflict dialog on a 409.
func (s *shell) send(ctx context.Context, w noteWrite) (func(), error) {
	return *new(func()), *new(error)
}

func (s *shell) update(ctx context.Context, w noteWrite) error { return *new(error) }

func (s *shell) saved(note schema.Note) {}

// leaveEditor pops the text area when it is the current view.
func (s *shell) leaveEditor() {}

// conflict asks whether to overwrite a note that changed on the server. Overwriting
// is never the default: Cancel has the focus.
func (s *shell) conflict(w noteWrite) {}

// isVersionConflict reports the 409 of a note that changed since it was read, not
// the 409 of a duplicate path, which force would not solve.
func isVersionConflict(err error) bool { return *new(bool) }

// conflictCode is the error code of the server for a stale lastModifiedAt.
const conflictCode = "conflict"

// deleteNote asks, then deletes the note.
func (s *shell) deleteNote(id, label string) {}

func (s *shell) remove(ctx context.Context, id, label string) (func(), error) {
	return *new(func()), *new(error)
}

func (s *shell) deleted(id, label string) {}
