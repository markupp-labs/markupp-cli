package cli

import (
	"io"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

func TestNewServicesWiresTheNotes(t *testing.T) {
	api, err := client.New(client.Options{BaseURL: "https://markupp.example.com", Tokens: client.StaticToken("tok")})
	if err != nil {
		t.Fatalf("client.New: %v", err)
	}
	if newServices(api).Notes == nil {
		t.Error("newServices left Notes nil")
	}
}

func TestNotesAloneShowsItsHelp(t *testing.T) {
	run := newTestMu().run(t, "notes")
	run.expectCode(t, 0)
	if !strings.Contains(run.stdout, "list") || !strings.Contains(run.stdout, "get") {
		t.Errorf("mu notes printed %q, want the subcommands", run.stdout)
	}
}

func TestNotesListPrintsATable(t *testing.T) {
	notes := newFakeNotes(map[string]string{"list": noteListBody})
	run := notesMu(notes).run(t, "notes", "list", "--vault", "v1", "--limit", "5", "--offset", "10")
	run.expectCode(t, 0)
	outputtest.Golden(t, "notes_list", []byte(run.stdout))
	want := noteCall{method: "list", vault: "v1", filter: noFilter, page: client.Page{Limit: 5, Offset: 10}}
	if notes.calls[0] != want {
		t.Errorf("call = %+v, want %+v", notes.calls[0], want)
	}
}

func TestNotesListStreamsNDJSON(t *testing.T) {
	notes := newFakeNotes(map[string]string{"list": noteListBody})
	run := notesMu(notes).run(t, "notes", "list", "--vault", "v1", "--ndjson")
	run.expectCode(t, 0)
	outputtest.Golden(t, "notes_list_ndjson", []byte(run.stdout))
}

func TestNotesListPassesTheJSONThrough(t *testing.T) {
	notes := newFakeNotes(map[string]string{"list": noteListBody})
	run := notesMu(notes).run(t, "notes", "list", "--vault", "v1", "--json")
	run.expectCode(t, 0)
	if run.stdout != noteListBody+"\n" {
		t.Errorf("stdout = %q, want the body unchanged", run.stdout)
	}
}

func TestNotesListTakesTheVaultFromTheEnvironment(t *testing.T) {
	notes := newFakeNotes(map[string]string{"list": `[]`})
	mu := notesMu(notes)
	mu.env["MARKUPP_VAULT"] = "v-env"
	mu.run(t, "notes", "list").expectCode(t, 0)
	if notes.calls[0].vault != "v-env" {
		t.Errorf("vault = %q, want v-env", notes.calls[0].vault)
	}
}

func TestNotesListWithoutVaultIsAUsageError(t *testing.T) {
	notes := newFakeNotes(nil)
	run := notesMu(notes).run(t, "notes", "list")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, "--vault") || len(notes.calls) != 0 {
		t.Errorf("stderr %q should point to --vault, calls %v should be empty", run.stderr, notes.calls)
	}
}

func TestNotesListRefusesANegativeLimit(t *testing.T) {
	run := notesMu(newFakeNotes(nil)).run(t, "notes", "list", "--vault", "v1", "--limit", "-1")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, "-1") {
		t.Errorf("stderr %q should name the value", run.stderr)
	}
}

func TestNotesListHelpDocumentsTheCap(t *testing.T) {
	run := newTestMu().run(t, "notes", "list", "--help")
	run.expectCode(t, 0)
	if !strings.Contains(run.stdout, "100") || !strings.Contains(run.stdout, "10000") {
		t.Errorf("help %q should document the limit and offset caps", run.stdout)
	}
}

func TestNotesGetPrintsMetadataAndDelimitedContent(t *testing.T) {
	notes := newFakeNotes(map[string]string{"get": noteBody})
	run := notesMu(notes).run(t, "notes", "get", "n1")
	run.expectCode(t, 0)
	outputtest.Golden(t, "notes_get", []byte(run.stdout))
	if notes.calls[0] != (noteCall{method: "get", id: "n1"}) {
		t.Errorf("call = %+v", notes.calls[0])
	}
}

func TestNotesGetContentPrintsTheRawContent(t *testing.T) {
	notes := newFakeNotes(map[string]string{"get": noteBody})
	run := notesMu(notes).run(t, "notes", "get", "n1", "--content")
	run.expectCode(t, 0)
	if run.stdout != "# Meeting\n- item \x1b[31mred\n" {
		t.Errorf("stdout = %q, want the content byte for byte", run.stdout)
	}
}

func TestNotesGetContentAndJSONAreExclusive(t *testing.T) {
	notes := newFakeNotes(map[string]string{"get": noteBody})
	run := notesMu(notes).run(t, "notes", "get", "n1", "--content", "--json")
	run.expectCode(t, 2)
	if len(notes.calls) != 0 {
		t.Errorf("calls = %v, want none", notes.calls)
	}
}

func TestNotesGetNotFoundExitsFive(t *testing.T) {
	notes := newFakeNotes(nil)
	notes.errs["get"] = failure.New(failure.NotFound, "GET /v1/notes/n9", "no note")
	run := notesMu(notes).run(t, "notes", "get", "n9")
	run.expectCode(t, 5)
	if run.stdout != "" || !strings.Contains(run.stderr, "n9") {
		t.Errorf("stdout %q should be empty, stderr %q should name the note", run.stdout, run.stderr)
	}
}

func TestNotesGetNeedsAnID(t *testing.T) {
	run := notesMu(newFakeNotes(nil)).run(t, "notes", "get")
	run.expectCode(t, 2)
}

func TestNotesListInterruptedMidStreamExits130(t *testing.T) {
	for _, mode := range [][]string{{}, {"--ndjson"}} {
		notes := newFakeNotes(nil)
		notes.streams = map[string]io.ReadCloser{"list": &interruptedBody{prefix: strings.NewReader(noteListBody[:40])}}
		run := notesMu(notes).run(t, append([]string{"notes", "list", "--vault", "v1"}, mode...)...)
		run.expectCode(t, 130)
		if strings.Contains(run.stderr, "malformed") {
			t.Errorf("%v: stderr %q calls an interrupt a malformed response", mode, run.stderr)
		}
	}
}

func TestNotesGetInterruptedWhileReadingExits130(t *testing.T) {
	notes := newFakeNotes(nil)
	notes.streams = map[string]io.ReadCloser{"get": &interruptedBody{prefix: strings.NewReader(`{"id":`)}}
	notesMu(notes).run(t, "notes", "get", "n1").expectCode(t, 130)
}

func TestNotesListRateLimitedExits7WithTheHint(t *testing.T) {
	notes := newFakeNotes(nil)
	notes.errs["list"] = failure.New(failure.RateLimited, "GET /v1/notes", "too many requests").WithHint("retry after 30 seconds")
	run := notesMu(notes).run(t, "notes", "list", "--vault", "v1")
	run.expectCode(t, 7)
	if run.stdout != "" || !strings.Contains(run.stderr, "Hint: retry after 30 seconds") {
		t.Errorf("stdout %q stderr %q, want nothing on stdout and the Retry-After hint on stderr", run.stdout, run.stderr)
	}
}
