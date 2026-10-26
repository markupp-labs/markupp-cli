package tui

import (
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// allLinkTypes reads every link type.
const allLinkTypes = ""

// newLinksView lists what the note id links to.
func newLinksView(s *shell, id, label string) *listView { return *new(*listView) }

// newBacklinksView lists the notes that link to the note id.
func newBacklinksView(s *shell, id, label string) *listView { return *new(*listView) }

// newMentionsView lists the notes the note id cites by title without a link.
func newMentionsView(s *shell, id, label string) *listView { return *new(*listView) }

// newSuggestionsView lists the notes worth linking from the note id.
func newSuggestionsView(s *shell, id, label string) *listView { return *new(*listView) }

// relationView is a listing of notes related to one note, with the note keys.
func (s *shell) relationView(resource, label string, columns []string, load loader) *listView {
	return *new(*listView)
}

func outgoingRow(link schema.Link) record { return *new(record) }

func incomingRow(link schema.Link) record { return *new(record) }

func mentionRow(mention schema.Mention) record { return *new(record) }

func suggestionRow(suggestion schema.LinkSuggestion) record { return *new(record) }

// linkTarget is the target of a link, or the guess of the indexer when unresolved.
func linkTarget(link schema.Link) *schema.NoteSummary { return *new(*schema.NoteSummary) }

// targetLabel is the path of the target, marked when it is a guess or missing.
func targetLabel(link schema.Link) string { return *new(string) }

// noteRecord is a row about a note, which may be missing: then the note keys skip it.
func noteRecord(note *schema.NoteSummary, cells ...string) record { return *new(record) }
