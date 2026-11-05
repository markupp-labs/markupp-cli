package cli

import (
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/client"
)

func TestNotesListSendsTheFilters(t *testing.T) {
	notes := newFakeNotes(map[string]string{"list": `[]`})
	mu := notesMu(notes)
	mu.run(t, "notes", "list", "--vault", "v1", "--orphan", "--community", "7", "--sort", "pagerank", "--direction", "asc").expectCode(t, 0)
	want := `orphan=true community=7 sort="pagerank" direction="asc"`
	if got := notes.calls[0].filter; got != want {
		t.Errorf("filter = %s, want %s", got, want)
	}
}

func TestNotesListSendsAnExplicitFalseOrphan(t *testing.T) {
	notes := newFakeNotes(map[string]string{"list": `[]`})
	notesMu(notes).run(t, "notes", "list", "--vault", "v1", "--orphan=false", "--sort", "updated").expectCode(t, 0)
	want := `orphan=false community=- sort="updated" direction=""`
	if got := notes.calls[0].filter; got != want {
		t.Errorf("filter = %s, want %s", got, want)
	}
}

func TestNotesListFiltersAPublicVault(t *testing.T) {
	notes := newFakeNotes(map[string]string{"public list": `[]`})
	notesMu(notes).run(t, "notes", "list", "--public", "garden", "--sort", "degree", "--direction", "desc").expectCode(t, 0)
	want := noteCall{method: "public list", slug: "garden", filter: `orphan=- community=- sort="degree" direction="desc"`}
	if notes.calls[0] != want {
		t.Errorf("call = %+v, want %+v", notes.calls[0], want)
	}
}

func TestNotesListRefusesAnUnknownOrder(t *testing.T) {
	for flag, value := range map[string]string{"--sort": "title", "--direction": "up"} {
		notes := newFakeNotes(nil)
		run := notesMu(notes).run(t, "notes", "list", "--vault", "v1", flag, value)
		run.expectCode(t, 2)
		if !strings.Contains(run.stderr, flag+`: invalid value "`+value+`"`) || len(notes.calls) != 0 {
			t.Errorf("%s %s: stderr %q, calls %v", flag, value, run.stderr, notes.calls)
		}
	}
}

func TestNotesListHelpDocumentsTheFilters(t *testing.T) {
	run := newTestMu().run(t, "notes", "list", "--help")
	for _, want := range []string{"--orphan", "--community", "--sort", "pagerank, degree or updated", "--direction", "--sort pagerank"} {
		if !strings.Contains(run.stdout, want) {
			t.Errorf("mu notes list --help lacks %q", want)
		}
	}
}

func TestNoteFilterFlagsLeaveUnsetFiltersOut(t *testing.T) {
	notes := newFakeNotes(map[string]string{"list": `[]`})
	notesMu(notes).run(t, "notes", "list", "--vault", "v1").expectCode(t, 0)
	if got := notes.calls[0].filter; got != describeFilter(client.NoteFilter{}) {
		t.Errorf("filter = %s, want none", got)
	}
}
