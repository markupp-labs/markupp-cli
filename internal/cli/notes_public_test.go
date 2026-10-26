package cli

import (
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

const publicNoteBody = `{"id":"n1","title":"meeting","path":"notas/meeting.md","content":"# Meeting\n",` +
	`"content_sha256":"ab","created_at":"2026-05-14T12:00:00Z","updated_at":"2026-05-15T12:00:00Z",` +
	`"degree":1,"pagerank":null,"community":null}`

// expectPublicCall asserts that mu made exactly want, without credentials.
func expectPublicCall(t *testing.T, mu *testMu, notes *fakeNotes, want noteCall) {
	t.Helper()
	if len(notes.calls) != 1 || notes.calls[0] != want {
		t.Fatalf("calls = %+v, want only %+v", notes.calls, want)
	}
	if conn := mu.server.connections[0]; !conn.Public {
		t.Errorf("connection = %+v, want a public one", conn)
	}
}

func TestNotesListReadsAPublicVault(t *testing.T) {
	notes := newFakeNotes(map[string]string{"public list": noteListBody})
	mu := notesMu(notes)
	mu.env["MARKUPP_VAULT"] = "v-env"
	run := mu.run(t, "notes", "list", "--public", "garden", "--limit", "5")
	run.expectCode(t, 0)
	outputtest.Golden(t, "notes_list", []byte(run.stdout))
	expectPublicCall(t, mu, notes, noteCall{method: "public list", slug: "garden", page: client.Page{Limit: 5}})
}

func TestNotesGetReadsAPublicNote(t *testing.T) {
	notes := newFakeNotes(map[string]string{"public get": publicNoteBody})
	mu := notesMu(notes)
	run := mu.run(t, "notes", "get", "n1", "--public", "garden", "--content")
	run.expectCode(t, 0)
	if run.stdout != "# Meeting\n" {
		t.Errorf("stdout = %q, want the content", run.stdout)
	}
	expectPublicCall(t, mu, notes, noteCall{method: "public get", slug: "garden", id: "n1"})
}

func TestNotesSearchSearchesAPublicVault(t *testing.T) {
	notes := newFakeNotes(map[string]string{"public search": searchBody})
	mu := notesMu(notes)
	mu.store.settings.Vault = "v-file"
	run := mu.run(t, "notes", "search", "kubernetes", "--public", "garden", "--offset", "10")
	run.expectCode(t, 0)
	expectPublicCall(t, mu, notes, noteCall{method: "public search", slug: "garden", text: "kubernetes", page: client.Page{Offset: 10}})
}

func TestNotesPublicAndVaultAreExclusive(t *testing.T) {
	for _, args := range [][]string{
		{"notes", "list", "--vault", "v1", "--public", "garden"},
		{"notes", "search", "x", "--vault", "v1", "--public", "garden"},
	} {
		notes := newFakeNotes(nil)
		mu := notesMu(notes)
		run := mu.run(t, args...)
		run.expectCode(t, 2)
		if !strings.Contains(run.stderr, `--public "garden"`) || len(mu.server.connections) != 0 {
			t.Errorf("%v: stderr %q, %d connections, want a usage error naming both and none", args, run.stderr, len(mu.server.connections))
		}
	}
}

func TestNotesPublicNamesTheMissingFeature(t *testing.T) {
	notes := newFakeNotes(nil)
	notes.errs["public get"] = errEndpointAbsent()
	run := notesMu(notes).run(t, "notes", "get", "n1", "--public", "garden")
	run.expectCode(t, 5)
	if !strings.Contains(run.stderr, "vault publication") {
		t.Errorf("stderr = %q, want the feature named", run.stderr)
	}
}

func TestNotesReadsDocumentThePublicFlag(t *testing.T) {
	for _, name := range []string{"list", "get", "search"} {
		run := newTestMu().run(t, "notes", name, "--help")
		if !strings.Contains(run.stdout, "--public string") || !strings.Contains(run.stdout, "--public my-garden") {
			t.Errorf("mu notes %s --help lacks --public or an example of it: %q", name, run.stdout)
		}
	}
}
