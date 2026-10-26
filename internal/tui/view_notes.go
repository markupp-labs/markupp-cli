package tui

import (
	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// listPage is the page every listing reads: the most the server returns at once.
var listPage = client.Page{Limit: client.MaxLimit}

// openNotes is :notes [vault].
func openNotes(s *shell, vault string) error { return *new(error) }

// newNotesView lists the notes of a vault; /!term searches it on the server.
func newNotesView(s *shell, vault string) *listView { return *new(*listView) }

func noteListRow(note schema.NoteListItem) record { return *new(record) }

// newSearchView lists the notes of vault that match term, most relevant first.
func newSearchView(s *shell, vault, term string) *listView { return *new(*listView) }

func searchRow(hit schema.SearchHit) record { return *new(record) }

// noteKeys are the keys of every row that is a note.
func (s *shell) noteKeys(selected func() (record, bool)) []Hotkey { return *new([]Hotkey) }

// onSelectedNote runs run on the selected row when it is a note.
func (s *shell) onSelectedNote(selected func() (record, bool), run func(r record)) func() {
	return *new(func())
}

// noteActions are the actions on a note, bound by on to how the note is chosen.
func (s *shell) noteActions(on func(run func(r record)) func()) []Hotkey { return *new([]Hotkey) }

func (s *shell) copySelected(selected func() (record, bool)) {}

// pushRelation opens a listing of the notes related to the note of r.
func (s *shell) pushRelation(r record, build func(s *shell, id, label string) *listView) {}

// need fails when a port this build of mu lacks is required.
func (s *shell) need(present bool, resource string) error { return *new(error) }
