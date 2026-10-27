package tui

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// startOnNotes runs mu tui on the notes of the vault and waits for them.
func startOnNotes(t *testing.T, api *fakeAPI, configure ...func(*shell)) *harness {
	t.Helper()
	h := startTUI(t, vaultConfig(newFakeConnector(api)), configure...)
	h.waitFor("work/plan.md")
	return h
}

func TestEnterOpensTheNoteSanitizedAndEscapeGoesBack(t *testing.T) {
	h := startOnNotes(t, notesAPI())
	h.press(tcell.KeyEnter)
	screen := h.waitFor("note(work/meeting.md)", "----- end of content -----")
	outputtest.Golden(t, "note", []byte(screen))
	if strings.Contains(screen, "\x1b") || !strings.Contains(screen, "[red]not a tag[-]") {
		t.Errorf("the content was not drawn as plain text:\n%s", screen)
	}
	h.press(tcell.KeyEscape)
	h.waitFor("notes(" + vaultID + ")[3]")
}

func TestNoteCommandOpensANoteById(t *testing.T) {
	api := notesAPI()
	h := startOnNotes(t, api)
	h.command("note " + noteA)
	h.waitFor("note(work/meeting.md)")
	if got := api.callsTo("get"); !slices.Contains(got, noteA) {
		t.Errorf("get calls = %q, want %s", got, noteA)
	}
}

func TestBangFilterSearchesTheServer(t *testing.T) {
	api := notesAPI()
	h := startOnNotes(t, api)
	h.typeText("/!meeting")
	h.press(tcell.KeyEnter)
	h.waitFor("search(meeting)[1]", "weekly [[meeting]] notes")
	if got := api.callsTo("search"); !slices.Equal(got, []string{vaultID + ":meeting"}) {
		t.Errorf("search calls = %q, want the vault and the term", got)
	}
	h.press(tcell.KeyEscape)
	h.waitFor("notes(" + vaultID + ")[3]")
}

func TestSearchOnAListingWithoutSearchFlashes(t *testing.T) {
	h := startTUI(t, testConfig(newFakeConnector(notesAPI())))
	h.waitFor("Acme")
	h.typeText("/!x")
	h.press(tcell.KeyEnter)
	h.waitFor("tenants has no server search")
}

// fakeEditor is a runEditor that writes edited into the file it is given and
// records the file and its mode.
type fakeEditor struct {
	edited string
	err    error
	path   string
	mode   os.FileMode
}

func (e *fakeEditor) run(argv []string) error {
	e.path = argv[len(argv)-1]
	info, err := os.Stat(e.path)
	if err != nil {
		return err
	}
	e.mode = info.Mode().Perm()
	if e.err != nil {
		return e.err
	}
	return os.WriteFile(e.path, []byte(e.edited), 0o600)
}

// editorShell sets $EDITOR to a fake editor working in dir.
func editorShell(editor *fakeEditor, dir string) func(*shell) {
	return func(s *shell) {
		s.cfg.Editor, s.cfg.TempDir, s.runEditor = "vim -n", dir, editor.run
	}
}

func TestEditSavesWhatTheEditorWroteAndRemovesTheFile(t *testing.T) {
	api, dir := notesAPI(), t.TempDir()
	editor := &fakeEditor{edited: "# Meeting\nnew agenda"}
	h := startOnNotes(t, api, editorShell(editor, dir))
	h.typeText("e")
	h.waitFor("saved work/meeting.md")
	want := schema.NoteUpdateRequest{Path: "work/meeting.md", Content: "# Meeting\nnew agenda", LastModifiedAt: "2026-05-15T10:00:00Z"}
	if got := api.sentUpdates(); len(got) != 1 || got[0] != want {
		t.Errorf("updates = %+v, want %+v", got, want)
	}
	assertScratchGone(t, editor, dir)
}

func assertScratchGone(t *testing.T, editor *fakeEditor, dir string) {
	t.Helper()
	if editor.mode != 0o600 || filepath.Dir(editor.path) != dir {
		t.Errorf("the editor got %s with mode %o, want a 0600 file in %s", editor.path, editor.mode, dir)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("the temporary file was left behind: %v", entries)
	}
}

func TestEditorFailureSavesNothingAndRemovesTheFile(t *testing.T) {
	api, dir := notesAPI(), t.TempDir()
	editor := &fakeEditor{err: errors.New("exit status 1")}
	h := startOnNotes(t, api, editorShell(editor, dir))
	h.typeText("e")
	h.waitFor("the editor failed, the note was not saved")
	if got := api.sentUpdates(); len(got) != 0 {
		t.Errorf("updates = %+v, want none", got)
	}
	assertScratchGone(t, editor, dir)
}

func TestEditWithoutChangesSendsNothing(t *testing.T) {
	api, dir := notesAPI(), t.TempDir()
	editor := &fakeEditor{edited: "# Meeting\nagenda \x1b[31mred\x1b[0m [red]not a tag[-]\nend"}
	h := startOnNotes(t, api, editorShell(editor, dir))
	h.typeText("e")
	h.waitFor("no changes to work/meeting.md")
	if got := api.sentUpdates(); len(got) != 0 {
		t.Errorf("updates = %+v, want none", got)
	}
}

func conflictAPI() *fakeAPI {
	api := notesAPI()
	conflict := &failure.Error{Kind: failure.Conflict, Code: "conflict", Message: "note changed", Subject: "PUT /v1/notes/x"}
	api.failNext("update", conflict)
	return api
}

func TestConflictDialogCancelsByDefault(t *testing.T) {
	api := conflictAPI()
	h := startOnNotes(t, api, editorShell(&fakeEditor{edited: "mine"}, t.TempDir()))
	h.typeText("e")
	screen := h.waitFor("Overwrite", "edits made on the server since then are lost")
	outputtest.Golden(t, "conflict", []byte(screen))
	h.press(tcell.KeyEnter)
	h.waitGone("Overwrite")
	if got := api.sentUpdates(); len(got) != 1 {
		t.Errorf("updates = %d, want only the refused one", len(got))
	}
}

func TestConflictCancelKeepsTheEditedText(t *testing.T) {
	api := conflictAPI()
	h := startOnNotes(t, api, editorShell(&fakeEditor{edited: "my unsaved words"}, t.TempDir()))
	h.typeText("e")
	h.waitFor("Overwrite", "Cancel keeps your version")
	h.press(tcell.KeyEnter)
	h.waitFor("edit(work/meeting.md)", "my unsaved words")
	h.press(tcell.KeyCtrlS)
	h.waitFor("saved work/meeting.md")
	if got := api.sentUpdates(); len(got) != 2 || got[1].Content != "my unsaved words" || got[1].Force {
		t.Errorf("updates = %+v, want the kept text sent again without force", got)
	}
}

func TestConflictOverwriteForcesTheWrite(t *testing.T) {
	api := conflictAPI()
	h := startOnNotes(t, api, editorShell(&fakeEditor{edited: "mine"}, t.TempDir()))
	h.typeText("e")
	h.waitFor("Overwrite")
	h.press(tcell.KeyRight)
	h.press(tcell.KeyEnter)
	h.waitFor("saved work/meeting.md")
	got := api.sentUpdates()
	if len(got) != 2 || got[0].Force || !got[1].Force || got[1].Content != "mine" {
		t.Errorf("updates = %+v, want a refused write then a forced one", got)
	}
}

func TestDuplicatePathIsNotAConflictToOverwrite(t *testing.T) {
	api := notesAPI()
	api.failNext("update", &failure.Error{Kind: failure.Conflict, Code: "duplicate_path", Message: "path taken"})
	h := startOnNotes(t, api, editorShell(&fakeEditor{edited: "mine"}, t.TempDir()))
	h.typeText("e")
	h.waitFor("path taken (duplicate_path)")
}

func TestTextAreaEditsWhenThereIsNoEditor(t *testing.T) {
	api := notesAPI()
	h := startOnNotes(t, api)
	h.typeText("e")
	h.waitFor("edit(work/meeting.md) ctrl-s saves, esc discards")
	h.typeText("!")
	h.press(tcell.KeyCtrlS)
	h.waitFor("saved work/meeting.md", "notes("+vaultID+")[3]")
	if got := api.sentUpdates(); len(got) != 1 || !strings.HasPrefix(got[0].Content, "!# Meeting") {
		t.Errorf("updates = %+v, want the typed text first", got)
	}
}

func TestTextAreaEscapeDiscards(t *testing.T) {
	api := notesAPI()
	h := startOnNotes(t, api)
	h.typeText("e")
	h.waitFor("edit(work/meeting.md)")
	h.typeText("q")
	h.press(tcell.KeyEscape)
	h.waitFor("edit of work/meeting.md discarded", "notes("+vaultID+")[3]")
	if got := api.sentUpdates(); len(got) != 0 {
		t.Errorf("updates = %+v, want none", got)
	}
}

func TestDeleteAsksAndCancelIsTheDefault(t *testing.T) {
	api := notesAPI()
	h := startOnNotes(t, api)
	h.press(tcell.KeyCtrlD)
	h.waitFor("Delete the note", "work/meeting.md")
	h.press(tcell.KeyEnter)
	h.waitGone("Delete the note")
	if got := api.callsTo("delete"); len(got) != 0 {
		t.Errorf("delete calls = %q, want none", got)
	}
}

func TestDeleteConfirmedDeletesTheNote(t *testing.T) {
	api := notesAPI()
	h := startOnNotes(t, api)
	h.press(tcell.KeyCtrlD)
	h.waitFor("Delete the note")
	h.press(tcell.KeyRight)
	h.press(tcell.KeyEnter)
	h.waitFor("deleted work/meeting.md")
	if got := api.callsTo("delete"); !slices.Equal(got, []string{noteA}) {
		t.Errorf("delete calls = %q, want %s", got, noteA)
	}
}

func TestLinksListTargetsAndUnresolvedTitles(t *testing.T) {
	api := notesAPI()
	h := startOnNotes(t, api)
	h.typeText("l")
	h.waitFor("links(work/meeting.md)[2]", "work/plan.md", "Budget (unresolved)")
	if got := api.callsTo("links"); !slices.Equal(got, []string{noteA}) {
		t.Errorf("links calls = %q, want %s", got, noteA)
	}
	h.press(tcell.KeyEnter)
	eventually(t, "the target opened", func() bool { return slices.Contains(api.callsTo("get"), noteC) })
}

func TestRelationKeysReadTheirRoute(t *testing.T) {
	for key, method := range map[string]string{"b": "backlinks", "m": "mentions", "s": "suggestions"} {
		api := notesAPI()
		h := startOnNotes(t, api)
		h.typeText(key)
		h.waitFor(method + "(work/meeting.md)[0]")
		if got := api.callsTo(method); !slices.Equal(got, []string{noteA}) {
			t.Errorf("%s calls = %q, want %s", method, got, noteA)
		}
		h.stop()
	}
}

func TestGraphIsATreeThatExpandsOnEnter(t *testing.T) {
	api := notesAPI()
	h := startOnNotes(t, api)
	h.typeText("g")
	h.waitFor("graph(work/meeting.md)", "-> explicit  work/plan.md", "-> explicit  Budget (unresolved)", "<- typed  animals/zebra.md")
	h.press(tcell.KeyDown)
	h.press(tcell.KeyEnter)
	screen := h.waitFor("-> embed  work/roadmap.md")
	outputtest.Golden(t, "graph", []byte(screen))
	if got := api.callsTo("graph"); !slices.Equal(got, []string{noteA, noteC}) {
		t.Errorf("graph calls = %q, want the note then the expanded neighbor", got)
	}
}

func TestEnterOnATenantReconnectsAsIt(t *testing.T) {
	connector := newFakeConnector(notesAPI())
	h := startTUI(t, testConfig(connector))
	h.waitFor("Acme")
	h.press(tcell.KeyEnter)
	h.waitFor("switched to Acme", "Tenant:  Acme")
	if got := connector.connected(); len(got) != 2 || got[1] != (Scope{Tenant: tenantA}) {
		t.Errorf("scopes = %+v, want a second connection as the tenant", got)
	}
}

func TestResourcesThisBuildLacksSaySo(t *testing.T) {
	h := startOnNotes(t, notesAPI())
	for _, command := range []string{"vaults", "members", "invites", "notifications"} {
		h.command(command)
		h.waitFor("error: :" + command + ": not available in this build of mu")
	}
}

func TestVaultsOpenTheirNotesAndMembers(t *testing.T) {
	api := notesAPI()
	h := startTUI(t, testConfig(newFakeConnector(api).withAll()))
	h.waitFor("vaults[1]", "Personal")
	h.typeText("m")
	h.waitFor("vault members(Personal)[1]", "ana@example.com")
	h.press(tcell.KeyEscape)
	h.press(tcell.KeyEnter)
	h.waitFor("notes("+vaultID+")[3]", "Vault:   "+vaultID)
}

func TestAdminListings(t *testing.T) {
	h := startTUI(t, testConfig(newFakeConnector(notesAPI()).withAll()))
	h.waitFor("Personal")
	h.command("members")
	h.waitFor("members[1]", "ana@example.com", "owner")
	h.command("invites")
	h.waitFor("invites[1]", "Beta", "member")
	h.command("notifications")
	h.waitFor("notifications[1]", "invite", "Beta")
}

func TestCommunitiesAndVaultLinksUseTheCurrentVault(t *testing.T) {
	api := notesAPI()
	api.bodies["communities"] = `[{"id":2,"size":7,"hub":{"id":"` + noteA + `","title":"meeting","path":"work/meeting.md"}}]`
	api.bodies["vault-links"] = linksBody
	h := startOnNotes(t, api)
	h.command("cm")
	h.waitFor("communities("+vaultID+")[1]", "work/meeting.md")
	h.command("links")
	h.waitFor("links("+vaultID+")[2]", "Budget (unresolved)")
	if got := api.callsTo("communities"); !slices.Equal(got, []string{vaultID}) {
		t.Errorf("communities calls = %q", got)
	}
}

func TestPublicOpensTheNotesOfTheSlug(t *testing.T) {
	connector := newFakeConnector(notesAPI())
	h := startTUI(t, vaultConfig(connector))
	h.waitFor("work/plan.md")
	h.command("public garden")
	h.waitFor("public:garden", "notes(garden)[3]")
	if got := connector.connected(); got[len(got)-1] != (Scope{PublicSlug: "garden"}) {
		t.Errorf("scopes = %+v, want the public slug last", got)
	}
}

func TestPublicWithoutSlugIsAUsageError(t *testing.T) {
	h := startOnNotes(t, notesAPI())
	h.command("public")
	h.waitFor("missing slug, expected :public <slug>")
}

func TestNeighborsFollowLinksBothWays(t *testing.T) {
	graph, err := decodeRaw[schema.Graph]([]byte(graphBodyA))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, n := range neighbors(graph, noteA) {
		got = append(got, n.text()+"|"+n.noteID())
	}
	want := []string{"-> explicit  work/plan.md|" + noteC, "-> explicit  Budget (unresolved)|", "<- typed  animals/zebra.md|" + noteB}
	if !slices.Equal(got, want) {
		t.Errorf("neighbors = %q, want %q", got, want)
	}
	if got := neighbors(graph, "missing"); got != nil {
		t.Errorf("neighbors of a missing note = %v, want none", got)
	}
}

func TestMalformedRecordsNameTheRecord(t *testing.T) {
	_, err := rowsOf(noteListRow)([]byte(`{"id":5}`))
	if err == nil || !strings.Contains(err.Error(), `malformed server record {"id":5}`) {
		t.Errorf("err = %v, want the record named", err)
	}
}
