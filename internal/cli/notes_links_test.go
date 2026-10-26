package cli

import (
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

const linksBody = `[
{"source":{"id":"n1","title":"a","path":"a.md"},"target":{"id":"n2","title":"b","path":"notes/b.md"},"cited_title":"B","type":"explicit","predicate":null,"inferred_target":null},
{"source":{"id":"n1","title":"a","path":"a.md"},"target":{"id":"n3","title":"Ana","path":"people/ana.md"},"cited_title":"Ana","type":"typed","predicate":"author","inferred_target":null},
{"source":{"id":"n1","title":"a","path":"a.md"},"target":null,"cited_title":"Meting","type":"explicit","predicate":null,"inferred_target":{"id":"n4","title":"meeting","path":"meeting.md"}}
]`

const mentionsBody = `[{"note":{"id":"n2","title":"kubernetes","path":"k8s/kubernetes.md"},"cited_title":"kubernetes","prior":0.3333}]`

const suggestionsBody = `[{"note":{"id":"n5","title":"helm","path":"k8s/helm.md"},"score":0.0325,"resource_allocation":0.5,"cosine":0.38}]`

func expectCalls(t *testing.T, got []portCall, want ...portCall) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("calls = %+v, want %+v", got, want)
	}
	for i := range want {
		if !equalCalls(got[i], want[i]) {
			t.Errorf("call %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// equalCalls compares two calls, the community pointer of the graph filter by value.
func equalCalls(a, b portCall) bool {
	ac, bc := a.graph.Community, b.graph.Community
	a.graph.Community, b.graph.Community = nil, nil
	if a != b || (ac == nil) != (bc == nil) {
		return false
	}
	return ac == nil || *ac == *bc
}

func TestNotesLinksText(t *testing.T) {
	mu, links := linkedMu(linksBody)
	run := mu.run(t, "notes", "links", "n1")
	run.expectCode(t, 0)
	outputtest.Golden(t, "notes_links", []byte(run.stdout))
	expectCalls(t, links.calls, portCall{method: "NoteLinks", subject: "n1"})
}

func TestNotesLinksBackShowsTheSource(t *testing.T) {
	mu, links := linkedMu(linksBody)
	run := mu.run(t, "notes", "links", "n2", "--back", "--type", "embed", "--limit", "5", "--offset", "10")
	run.expectCode(t, 0)
	outputtest.Golden(t, "notes_links_back", []byte(run.stdout))
	page := client.Page{Limit: 5, Offset: 10}
	expectCalls(t, links.calls, portCall{method: "NoteBacklinks", subject: "n2", linkType: "embed", page: page})
}

func TestNotesLinksJSONIsTheBodyUnchanged(t *testing.T) {
	mu, _ := linkedMu(linksBody)
	run := mu.run(t, "notes", "links", "n1", "--json")
	run.expectCode(t, 0)
	if run.stdout != linksBody+"\n" {
		t.Errorf("--json printed %q, want the body unchanged", run.stdout)
	}
}

func TestNotesLinksRefusesAnUnknownType(t *testing.T) {
	mu, links := linkedMu(linksBody)
	run := mu.run(t, "notes", "links", "n1", "--type", "wiki")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, `"wiki"`) || !strings.Contains(run.stderr, "explicit, embed, typed, inferred") {
		t.Errorf("stderr %q should name the value and the choices", run.stderr)
	}
	if len(links.calls) != 0 || len(mu.server.connections) != 0 {
		t.Error("an invalid --type reached the server")
	}
}

func TestNotesLinksNeedsAnID(t *testing.T) {
	mu, _ := linkedMu(linksBody)
	mu.run(t, "notes", "links").expectCode(t, 2)
}

func TestNotesLinksOfAMissingNoteExitsNotFound(t *testing.T) {
	mu, links := linkedMu("")
	links.err = failure.New(failure.NotFound, "GET /v1/notes/n9/links", "note n9 not found")
	run := mu.run(t, "notes", "links", "n9")
	run.expectCode(t, 5)
	if run.stdout != "" || !strings.Contains(run.stderr, "n9") {
		t.Errorf("stdout %q should be empty and stderr %q name the note", run.stdout, run.stderr)
	}
}

func TestNotesLinksNeedsAServer(t *testing.T) {
	mu, _ := linkedMu(linksBody)
	delete(mu.env, "MARKUPP_URL")
	mu.run(t, "notes", "links", "n1").expectCode(t, 2)
}

func TestNotesMentionsText(t *testing.T) {
	mu, links := linkedMu(mentionsBody)
	run := mu.run(t, "notes", "mentions", "n1", "--limit", "3")
	run.expectCode(t, 0)
	outputtest.Golden(t, "notes_mentions", []byte(run.stdout))
	expectCalls(t, links.calls, portCall{method: "NoteMentions", subject: "n1", page: client.Page{Limit: 3}})
}

func TestNotesMentionsBackNDJSON(t *testing.T) {
	mu, links := linkedMu(mentionsBody)
	run := mu.run(t, "notes", "mentions", "n1", "--back", "--ndjson")
	run.expectCode(t, 0)
	outputtest.Golden(t, "notes_mentions_back_ndjson", []byte(run.stdout))
	expectCalls(t, links.calls, portCall{method: "NoteBackmentions", subject: "n1"})
}

func TestNotesMentionsRefusesANegativeLimit(t *testing.T) {
	mu, links := linkedMu(mentionsBody)
	mu.run(t, "notes", "mentions", "n1", "--limit", "-1").expectCode(t, 2)
	if len(links.calls) != 0 {
		t.Error("a negative --limit reached the server")
	}
}

func TestNotesSuggestText(t *testing.T) {
	mu, links := linkedMu(suggestionsBody)
	run := mu.run(t, "notes", "suggest", "n1")
	run.expectCode(t, 0)
	outputtest.Golden(t, "notes_suggest", []byte(run.stdout))
	expectCalls(t, links.calls, portCall{method: "NoteSuggestions", subject: "n1"})
}

func TestNotesSuggestStopsOnAMalformedRecord(t *testing.T) {
	mu, _ := linkedMu(`[{"note":"n5"}]`)
	run := mu.run(t, "notes", "suggest", "n1")
	run.expectCode(t, 1)
	if !strings.Contains(run.stderr, `"note":"n5"`) {
		t.Errorf("stderr %q should name the malformed record", run.stderr)
	}
}

func TestNotesGroupListsTheLinkCommands(t *testing.T) {
	run := newTestMu().run(t, "notes")
	run.expectCode(t, 0)
	for _, name := range []string{"links", "mentions", "suggest"} {
		if !strings.Contains(run.stdout, name) {
			t.Errorf("mu notes help %q does not list %s", run.stdout, name)
		}
	}
	newTestMu().run(t, "notes", "bogus").expectCode(t, 2)
}
