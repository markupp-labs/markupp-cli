package tui

import (
	"context"
	"fmt"
	"io"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// listPage is the page every listing reads: the most the server returns at once.
var listPage = client.Page{Limit: client.MaxLimit}

// openNotes is :notes [vault].
func openNotes(s *shell, vault string) error {
	if vault == "" {
		vault = s.vault
	}
	if err := s.need(s.ports.Notes != nil, "notes"); err != nil {
		return err
	}
	if vault == "" {
		return failure.New(failure.Usage, ":notes", "no vault selected, expected :notes <vault id>, :vaults or mu tui --vault")
	}
	s.vault = vault
	s.reset(newNotesView(s, vault))
	return nil
}

// newNotesView lists the notes of a vault; /!term searches it on the server.
func newNotesView(s *shell, vault string) *listView {
	return newListView(s, listing{
		resource: "notes",
		context:  vault,
		columns:  []string{"PATH", "UPDATED", "DEGREE", "RANK", "COMMUNITY", "ID"},
		load:     arrayOf(func(ctx context.Context) (io.ReadCloser, error) { return s.ports.Notes.ListNotes(ctx, vault, listPage) }, rowsOf(noteListRow)),
		enter:    func(r record) { s.viewNote(r.note) },
		keys:     s.noteKeys,
		search:   func(term string) { s.push(newSearchView(s, vault, term)) },
	})
}

func noteListRow(note schema.NoteListItem) record {
	cells := []string{note.Path, note.UpdatedAt, dashInt(note.Degree), dashFloat(note.PageRank), dashInt64(note.Community), note.ID}
	return record{id: note.ID, note: note.ID, label: note.Path, cells: cells}
}

// newSearchView lists the notes of vault that match term, most relevant first.
func newSearchView(s *shell, vault, term string) *listView {
	return newListView(s, listing{
		resource: "search",
		context:  term,
		columns:  []string{"PATH", "SNIPPET", "RANK", "ID"},
		load: arrayOf(func(ctx context.Context) (io.ReadCloser, error) {
			return s.ports.Notes.SearchNotes(ctx, term, vault, listPage)
		}, rowsOf(searchRow)),
		enter:  func(r record) { s.viewNote(r.note) },
		keys:   s.noteKeys,
		search: func(next string) { s.push(newSearchView(s, vault, next)) },
	})
}

func searchRow(hit schema.SearchHit) record {
	cells := []string{hit.Note.Path, output.Snippet(hit.Snippet), decimal(hit.Rank), hit.Note.ID}
	return record{id: hit.Note.ID, note: hit.Note.ID, label: hit.Note.Path, cells: cells}
}

// noteKeys are the keys of every row that is a note.
func (s *shell) noteKeys(selected func() (record, bool)) []Hotkey {
	onNote := func(run func(r record)) func() { return s.onSelectedNote(selected, run) }
	return append(s.noteActions(onNote),
		runeKey('y', "copy id", func() { s.copySelected(selected) }))
}

// onSelectedNote runs run on the selected row when it is a note.
func (s *shell) onSelectedNote(selected func() (record, bool), run func(r record)) func() {
	return func() {
		row, ok := selected()
		if !ok || row.note == "" {
			s.flashError(fmt.Errorf("no note selected, expected a row that is a note"))
			return
		}
		run(row)
	}
}

// noteActions are the actions on a note, bound by on to how the note is chosen.
func (s *shell) noteActions(on func(run func(r record)) func()) []Hotkey {
	return []Hotkey{
		runeKey('v', "view", on(func(r record) { s.viewNote(r.note) })),
		runeKey('e', "edit", on(func(r record) { s.editNote(r.note) })),
		ctrlKey(ctrlDelete, "delete", on(func(r record) { s.deleteNote(r.note, r.label) })),
		runeKey('l', "links", on(func(r record) { s.pushRelation(r, newLinksView) })),
		runeKey('b', "backlinks", on(func(r record) { s.pushRelation(r, newBacklinksView) })),
		runeKey('m', "mentions", on(func(r record) { s.pushRelation(r, newMentionsView) })),
		runeKey('s', "suggestions", on(func(r record) { s.pushRelation(r, newSuggestionsView) })),
		runeKey('g', "graph", on(func(r record) { s.pushGraph(r.note, r.label) })),
	}
}

func (s *shell) copySelected(selected func() (record, bool)) {
	row, _ := selected()
	s.copyText(row.id)
}

// pushRelation opens a listing of the notes related to the note of r.
func (s *shell) pushRelation(r record, build func(s *shell, id, label string) *listView) {
	if err := s.need(s.ports.Relations != nil, "links"); err != nil {
		s.flashError(err)
		return
	}
	s.push(build(s, r.note, r.label))
}

// need fails when a port this build of mu lacks is required.
func (s *shell) need(present bool, resource string) error {
	if present {
		return nil
	}
	return failure.New(failure.NotFound, ":"+resource, "not available in this build of mu or in this scope")
}
