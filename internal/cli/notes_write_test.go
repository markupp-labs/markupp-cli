package cli

import (
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// fakeFiles is the filesystem of a test, by file name.
type fakeFiles map[string]string

func (f fakeFiles) read(name string) ([]byte, error) {
	content, ok := f[name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return []byte(content), nil
}

const since = "2026-05-15T12:00:00.5Z"

func editMu(notes *fakeNotes) *testMu {
	mu := notesMu(notes)
	mu.files = fakeFiles{"meeting.md": "# New\n"}
	return mu
}

func TestNotesCreatePrintsTheID(t *testing.T) {
	notes := newFakeNotes(map[string]string{"create": noteBody})
	mu := editMu(notes)
	run := mu.run(t, "notes", "create", "notas/meeting.md", "--file", "meeting.md", "--vault", "v1")
	run.expectCode(t, 0)
	if run.stdout != "n1\n" {
		t.Errorf("stdout = %q, want the id", run.stdout)
	}
	want := noteCall{method: "create", vault: "v1", create: schema.NoteCreateRequest{Path: "notas/meeting.md", Content: "# New\n"}}
	if notes.calls[0] != want {
		t.Errorf("call = %+v, want %+v", notes.calls[0], want)
	}
}

func TestNotesCreateReadsStdinWithADash(t *testing.T) {
	notes := newFakeNotes(map[string]string{"create": noteBody})
	mu := editMu(notes)
	mu.stdin = strings.NewReader("from stdin")
	mu.run(t, "notes", "create", "a.md", "--file", "-", "--vault", "v1").expectCode(t, 0)
	if notes.calls[0].create.Content != "from stdin" {
		t.Errorf("content = %q, want the stdin", notes.calls[0].create.Content)
	}
}

func TestNotesCreatePassesTheJSONThrough(t *testing.T) {
	notes := newFakeNotes(map[string]string{"create": noteBody})
	run := editMu(notes).run(t, "notes", "create", "a.md", "--file", "meeting.md", "--vault", "v1", "--json")
	run.expectCode(t, 0)
	if run.stdout != noteBody+"\n" {
		t.Errorf("stdout = %q, want the body unchanged", run.stdout)
	}
}

func TestNotesCreateNeedsTheFileFlag(t *testing.T) {
	notes := newFakeNotes(nil)
	run := editMu(notes).run(t, "notes", "create", "a.md", "--vault", "v1")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, "--file") || len(notes.calls) != 0 {
		t.Errorf("stderr %q should name --file, calls %v should be empty", run.stderr, notes.calls)
	}
}

func TestNotesCreateNamesAMissingFile(t *testing.T) {
	run := editMu(newFakeNotes(nil)).run(t, "notes", "create", "a.md", "--file", "nope.md", "--vault", "v1")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, "nope.md") {
		t.Errorf("stderr %q should name the file", run.stderr)
	}
}

func TestNotesCreateNeedsAVault(t *testing.T) {
	run := editMu(newFakeNotes(nil)).run(t, "notes", "create", "a.md", "--file", "meeting.md")
	run.expectCode(t, 2)
}

func TestNotesCreateDuplicatePathExitsSix(t *testing.T) {
	notes := newFakeNotes(nil)
	notes.errs["create"] = &failure.Error{Kind: failure.Conflict, Code: "duplicate_path", Message: "path taken"}
	editMu(notes).run(t, "notes", "create", "a.md", "--file", "meeting.md", "--vault", "v1").expectCode(t, 6)
}

func TestNotesEditSendsTheVersionAndPrintsTheNewOne(t *testing.T) {
	notes := newFakeNotes(map[string]string{"update": noteBody})
	run := editMu(notes).run(t, "notes", "edit", "n1", "--file", "meeting.md", "--since", since, "--path", "b.md")
	run.expectCode(t, 0)
	outputtest.Golden(t, "notes_edit", []byte(run.stdout))
	want := noteCall{method: "update", id: "n1", update: schema.NoteUpdateRequest{Path: "b.md", Content: "# New\n", LastModifiedAt: since}}
	if len(notes.calls) != 1 || notes.calls[0] != want {
		t.Errorf("calls = %+v, want only %+v", notes.calls, want)
	}
}

func TestNotesEditWithoutPathKeepsTheCurrentOne(t *testing.T) {
	notes := newFakeNotes(map[string]string{"get": noteBody, "update": noteBody})
	editMu(notes).run(t, "notes", "edit", "n1", "--file", "meeting.md", "--since", since).expectCode(t, 0)
	if len(notes.calls) != 2 || notes.calls[0].method != "get" || notes.calls[1].update.Path != "notas/meeting.md" {
		t.Errorf("calls = %+v, want a get then an update keeping notas/meeting.md", notes.calls)
	}
	if notes.calls[1].update.LastModifiedAt != since {
		t.Errorf("lastModifiedAt = %q, want --since, never the version just read", notes.calls[1].update.LastModifiedAt)
	}
}

func TestNotesEditNeedsSinceUnlessForced(t *testing.T) {
	notes := newFakeNotes(nil)
	run := editMu(notes).run(t, "notes", "edit", "n1", "--file", "meeting.md", "--path", "a.md")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, "--since") || len(notes.calls) != 0 {
		t.Errorf("stderr %q should name --since, calls %v should be empty", run.stderr, notes.calls)
	}
}

func TestNotesEditRefusesAMalformedSince(t *testing.T) {
	run := editMu(newFakeNotes(nil)).run(t, "notes", "edit", "n1", "--file", "meeting.md", "--path", "a.md", "--since", "yesterday")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, `"yesterday"`) || !strings.Contains(run.stderr, "RFC 3339") {
		t.Errorf("stderr %q should name the value and the expected format", run.stderr)
	}
}

func TestNotesEditForceOverwritesWithoutSince(t *testing.T) {
	notes := newFakeNotes(map[string]string{"update": noteBody})
	editMu(notes).run(t, "notes", "edit", "n1", "--file", "meeting.md", "--path", "a.md", "--force").expectCode(t, 0)
	got := notes.calls[0].update
	if !got.Force || got.LastModifiedAt != "0001-01-01T00:00:00Z" {
		t.Errorf("update = %+v, want force with the zero time", got)
	}
}

func TestNotesEditConflictExitsSixWithoutRetry(t *testing.T) {
	notes := newFakeNotes(nil)
	notes.errs["update"] = &failure.Error{Kind: failure.Conflict, Code: "conflict", Subject: "PUT /v1/notes/n1", Message: "changed"}
	run := editMu(notes).run(t, "notes", "edit", "n1", "--file", "meeting.md", "--path", "a.md", "--since", since)
	run.expectCode(t, 6)
	if len(notes.calls) != 1 {
		t.Errorf("calls = %+v, want one update and no retry", notes.calls)
	}
	if !strings.Contains(run.stderr, "Hint:") || !strings.Contains(run.stderr, "--force") || run.stdout != "" {
		t.Errorf("stderr %q should hint at merging or --force, stdout %q should be empty", run.stderr, run.stdout)
	}
}

func TestNotesEditNeverReadsStdinImplicitly(t *testing.T) {
	notes := newFakeNotes(nil)
	mu := editMu(notes)
	mu.stdin = strings.NewReader("implicit")
	mu.run(t, "notes", "edit", "n1", "--since", since, "--path", "a.md").expectCode(t, 2)
	if len(notes.calls) != 0 {
		t.Errorf("calls = %+v, want none", notes.calls)
	}
}

func TestNotesEditForceHelpSaysWhatIsLost(t *testing.T) {
	run := newTestMu().run(t, "notes", "edit", "--help")
	run.expectCode(t, 0)
	if !strings.Contains(run.stdout, "lost") {
		t.Errorf("edit help should say what --force loses:\n%s", run.stdout)
	}
}

func TestNotesDeleteRequiresYes(t *testing.T) {
	notes := newFakeNotes(nil)
	run := notesMu(notes).run(t, "notes", "delete", "n1")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, "--yes") || len(notes.calls) != 0 {
		t.Errorf("stderr %q should name --yes, calls %v should be empty", run.stderr, notes.calls)
	}
}

func TestNotesDeleteWithYesDeletesAndConfirmsOnStderr(t *testing.T) {
	notes := newFakeNotes(nil)
	run := notesMu(notes).run(t, "notes", "delete", "n1", "--yes")
	run.expectCode(t, 0)
	if notes.calls[0] != (noteCall{method: "delete", id: "n1"}) {
		t.Errorf("call = %+v", notes.calls[0])
	}
	if run.stdout != "" || !strings.Contains(run.stderr, "n1") {
		t.Errorf("stdout %q should be empty, stderr %q should confirm", run.stdout, run.stderr)
	}
}

func TestNotesDeleteNotFoundExitsFive(t *testing.T) {
	notes := newFakeNotes(nil)
	notes.errs["delete"] = failure.New(failure.NotFound, "DELETE /v1/notes/n9", "no note")
	notesMu(notes).run(t, "notes", "delete", "n9", "--yes").expectCode(t, 5)
}

func TestReadContentReportsAStdinFailure(t *testing.T) {
	app := newApp(Deps{Stdin: failingReader{}})
	if _, err := app.readContent("-"); failure.KindOf(err) != failure.Usage {
		t.Errorf("readContent error = %v, want usage", err)
	}
}

// failingReader fails every read, like a closed stdin.
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("stdin closed") }
