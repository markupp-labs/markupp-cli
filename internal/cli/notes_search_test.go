package cli

import (
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

const searchBody = `[{"note":{"id":"3b241101-e2bb-4255-8caf-4136c566a962","title":"meeting","path":"notas/meeting.md"},` +
	`"vault_id":"v1","rank":0.0607,"snippet":"about the \u0002kubernetes\u0003 cluster\nnew"},` +
	`{"note":{"id":"n2","title":"b","path":"b.md"},"vault_id":"v2","rank":0.01,"snippet":""}]`

func TestNotesSearchPrintsHitsWithMarkedSnippets(t *testing.T) {
	notes := newFakeNotes(map[string]string{"search": searchBody})
	run := notesMu(notes).run(t, "notes", "search", "kubernetes -draft", "--vault", "v1", "--limit", "5")
	run.expectCode(t, 0)
	outputtest.Golden(t, "notes_search", []byte(run.stdout))
	want := noteCall{method: "search", text: "kubernetes -draft", vault: "v1", page: client.Page{Limit: 5}}
	if notes.calls[0] != want {
		t.Errorf("call = %+v, want %+v", notes.calls[0], want)
	}
}

func TestNotesSearchKeepsTheMarkersInJSON(t *testing.T) {
	notes := newFakeNotes(map[string]string{"search": searchBody})
	run := notesMu(notes).run(t, "notes", "search", "kubernetes", "--json")
	run.expectCode(t, 0)
	if run.stdout != searchBody+"\n" {
		t.Errorf("stdout = %q, want the body unchanged", run.stdout)
	}
}

func TestNotesSearchStreamsNDJSON(t *testing.T) {
	notes := newFakeNotes(map[string]string{"search": searchBody})
	run := notesMu(notes).run(t, "notes", "search", "kubernetes", "--ndjson")
	run.expectCode(t, 0)
	outputtest.Golden(t, "notes_search_ndjson", []byte(run.stdout))
}

func TestNotesSearchWithoutVaultSearchesEveryVault(t *testing.T) {
	notes := newFakeNotes(map[string]string{"search": `[]`})
	notesMu(notes).run(t, "notes", "search", "x").expectCode(t, 0)
	if notes.calls[0].vault != "" {
		t.Errorf("vault = %q, want empty for every vault", notes.calls[0].vault)
	}
}

func TestNotesSearchUsesTheConfiguredVault(t *testing.T) {
	notes := newFakeNotes(map[string]string{"search": `[]`})
	mu := notesMu(notes)
	mu.store.settings.Vault = "v-file"
	mu.run(t, "notes", "search", "x").expectCode(t, 0)
	if notes.calls[0].vault != "v-file" {
		t.Errorf("vault = %q, want v-file", notes.calls[0].vault)
	}
}

func TestNotesSearchNeedsAQuery(t *testing.T) {
	notes := newFakeNotes(nil)
	run := notesMu(notes).run(t, "notes", "search")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, "QUERY") || len(notes.calls) != 0 {
		t.Errorf("stderr %q should name QUERY, calls %v should be empty", run.stderr, notes.calls)
	}
}

func TestNotesSearchRefusesAnEmptyQuery(t *testing.T) {
	notes := newFakeNotes(nil)
	run := notesMu(notes).run(t, "notes", "search", "  ")
	run.expectCode(t, 2)
	if len(notes.calls) != 0 {
		t.Errorf("calls = %v, want none", notes.calls)
	}
}

func TestNotesSearchHelpSaysContentIsASecondCall(t *testing.T) {
	run := newTestMu().run(t, "notes", "search", "--help")
	run.expectCode(t, 0)
	for _, want := range []string{"100", "10000", "mu notes get", "per result"} {
		if !strings.Contains(run.stdout, want) {
			t.Errorf("help does not mention %q:\n%s", want, run.stdout)
		}
	}
}
