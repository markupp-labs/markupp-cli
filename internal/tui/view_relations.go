package tui

import (
	"context"
	"io"

	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// allLinkTypes reads every link type.
const allLinkTypes = ""

// newLinksView lists what the note id links to.
func newLinksView(s *shell, id, label string) *listView {
	fetch := func(ctx context.Context) (io.ReadCloser, error) {
		return s.ports.Relations.NoteLinks(ctx, id, allLinkTypes, listPage)
	}
	return s.relationView("links", label, []string{"TYPE", "TARGET", "PREDICATE", "ID"}, arrayOf(fetch, rowsOf(outgoingRow)))
}

// newBacklinksView lists the notes that link to the note id.
func newBacklinksView(s *shell, id, label string) *listView {
	fetch := func(ctx context.Context) (io.ReadCloser, error) {
		return s.ports.Relations.NoteBacklinks(ctx, id, allLinkTypes, listPage)
	}
	return s.relationView("backlinks", label, []string{"TYPE", "SOURCE", "PREDICATE", "ID"}, arrayOf(fetch, rowsOf(incomingRow)))
}

// newMentionsView lists the notes the note id cites by title without a link.
func newMentionsView(s *shell, id, label string) *listView {
	fetch := func(ctx context.Context) (io.ReadCloser, error) {
		return s.ports.Relations.NoteMentions(ctx, id, listPage)
	}
	return s.relationView("mentions", label, []string{"NOTE", "CITED", "PRIOR", "ID"}, arrayOf(fetch, rowsOf(mentionRow)))
}

// newSuggestionsView lists the notes worth linking from the note id.
func newSuggestionsView(s *shell, id, label string) *listView {
	fetch := func(ctx context.Context) (io.ReadCloser, error) {
		return s.ports.Relations.NoteSuggestions(ctx, id, listPage)
	}
	return s.relationView("suggestions", label, []string{"NOTE", "SCORE", "COSINE", "ID"}, arrayOf(fetch, rowsOf(suggestionRow)))
}

// relationView is a listing of notes related to one note, with the note keys.
func (s *shell) relationView(resource, label string, columns []string, load loader) *listView {
	return newListView(s, listing{
		resource: resource, context: label, columns: columns, load: load,
		enter: func(r record) { s.viewNote(r.note) },
		keys:  s.noteKeys,
	})
}

func outgoingRow(link schema.Link) record {
	return noteRecord(linkTarget(link), link.Type, targetLabel(link), dashText(link.Predicate))
}

func incomingRow(link schema.Link) record {
	return noteRecord(&link.Source, link.Type, link.Source.Path, dashText(link.Predicate))
}

func mentionRow(mention schema.Mention) record {
	return noteRecord(&mention.Note, mention.Note.Path, mention.CitedTitle, decimal(mention.Prior))
}

func suggestionRow(suggestion schema.LinkSuggestion) record {
	return noteRecord(&suggestion.Note, suggestion.Note.Path, decimal(suggestion.Score), decimal(suggestion.Cosine))
}

// linkTarget is the target of a link, or the guess of the indexer when unresolved.
func linkTarget(link schema.Link) *schema.NoteSummary {
	if link.Target != nil {
		return link.Target
	}
	return link.InferredTarget
}

// targetLabel is the path of the target, marked when it is a guess or missing.
func targetLabel(link schema.Link) string {
	switch {
	case link.Target != nil:
		return link.Target.Path
	case link.InferredTarget != nil:
		return link.InferredTarget.Path + " (inferred)"
	default:
		return link.CitedTitle + " (unresolved)"
	}
}

// noteRecord is a row about a note, which may be missing: then the note keys skip it.
func noteRecord(note *schema.NoteSummary, cells ...string) record {
	if note == nil {
		return record{cells: append(cells, "-")}
	}
	return record{id: note.ID, note: note.ID, label: note.Path, cells: append(cells, note.ID)}
}
