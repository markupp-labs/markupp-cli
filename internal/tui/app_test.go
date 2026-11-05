package tui

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

func TestStartShowsTheHeaderTheNotesAndTheCrumbs(t *testing.T) {
	h := startTUI(t, vaultConfig(newFakeConnector(notesAPI())))
	screen := h.waitFor("ana@example.com", "work/plan.md", "notes("+vaultID+")[3]")
	outputtest.Golden(t, "start", []byte(screen))
}

func TestRunFailsBeforeDrawingWhenTheConnectionFails(t *testing.T) {
	connector := newFakeConnector(notesAPI())
	connector.err = failure.New(failure.Unauthenticated, "https://x", "not logged in")
	err := Run(context.Background(), testConfig(connector))
	if failure.KindOf(err) != failure.Unauthenticated {
		t.Errorf("Run = %v, want the unauthenticated error of the connector", err)
	}
}

func TestRunNeedsAConnector(t *testing.T) {
	if err := Run(context.Background(), Config{}); failure.KindOf(err) != failure.Internal {
		t.Errorf("Run without connector = %v, want an internal error", err)
	}
}

func TestStartWithoutAVaultListsTheTenants(t *testing.T) {
	h := startTUI(t, testConfig(newFakeConnector(notesAPI())))
	h.waitFor("tenants[1]", "Acme", "owner", "42")
}

func TestQuitCommandStopsTheInterface(t *testing.T) {
	h := startTUI(t, vaultConfig(newFakeConnector(notesAPI())))
	h.waitFor("work/plan.md")
	h.command("q")
	select {
	case err := <-h.done:
		if err != nil {
			t.Errorf("run after :q = %v, want nil", err)
		}
		h.done <- err
	case <-time.After(waitTimeout):
		t.Fatal(":q did not stop the interface")
	}
}

func TestUnknownCommandFlashesTheCommands(t *testing.T) {
	h := startTUI(t, vaultConfig(newFakeConnector(notesAPI())))
	h.command("pods")
	h.waitFor(`unknown command "pods", expected one of: notes`)
}

func TestFilterKeepsMatchingRowsAndEscapeClearsIt(t *testing.T) {
	h := startTUI(t, vaultConfig(newFakeConnector(notesAPI())))
	h.waitFor("animals/zebra.md")
	h.typeText("/work")
	h.press(tcell.KeyEnter)
	h.waitFor("notes("+vaultID+")[2] /work", "work/plan.md")
	h.waitGone("animals/zebra.md")
	h.press(tcell.KeyEscape)
	h.waitFor("notes("+vaultID+")[3]", "animals/zebra.md")
}

func TestShiftLetterSortsByTheColumn(t *testing.T) {
	h := startTUI(t, vaultConfig(newFakeConnector(notesAPI())))
	h.waitFor("work/plan.md")
	h.typeText("P")
	screen := h.waitFor("PATH ^")
	if strings.Index(screen, "animals/zebra.md") > strings.Index(screen, "work/meeting.md") {
		t.Errorf("sorted by PATH, zebra is not first:\n%s", screen)
	}
	h.typeText("D")
	screen = h.waitFor("DEGREE ^")
	if strings.Index(screen, "work/plan.md") > strings.Index(screen, "work/meeting.md") {
		t.Errorf("sorted by DEGREE, plan (1) is not first:\n%s", screen)
	}
}

func TestCommandAliasOpensTheResource(t *testing.T) {
	h := startTUI(t, testConfig(newFakeConnector(notesAPI())))
	h.waitFor("Acme")
	h.command("n " + vaultID)
	h.waitFor("notes("+vaultID+")[3]", "Vault:   "+vaultID)
}

func TestNotesWithoutAVaultFlashesHowToPickOne(t *testing.T) {
	h := startTUI(t, testConfig(newFakeConnector(notesAPI())))
	h.waitFor("Acme")
	h.command("notes")
	h.waitFor("no vault selected, expected :notes <vault id>")
}

func TestHelpListsTheKeysOfTheViewAndTheCommands(t *testing.T) {
	h := startTUI(t, vaultConfig(newFakeConnector(notesAPI())))
	h.waitFor("work/plan.md")
	h.typeText("?")
	h.waitFor("help(notes)[35]", "<enter>", "sort PATH")
	h.press(tcell.KeyEscape)
	h.waitFor("notes(" + vaultID + ")[3]")
}

func TestCopyWritesTheIDAsOSC52(t *testing.T) {
	h := startTUI(t, vaultConfig(newFakeConnector(notesAPI())))
	h.waitFor("work/meeting.md")
	h.typeText("y")
	h.waitFor("copied " + noteA)
	if got := h.terminal.String(); got != osc52(noteA) {
		t.Errorf("terminal got %q, want the OSC 52 of %s", got, noteA)
	}
}

func TestCtrlRReloadsTheView(t *testing.T) {
	api := notesAPI()
	h := startTUI(t, vaultConfig(newFakeConnector(api)))
	h.waitFor("work/plan.md")
	h.press(tcell.KeyCtrlR)
	eventually(t, "a second listing", func() bool { return len(api.callsTo("list")) >= 2 })
}

func TestAutoRefreshReloadsUntilPaused(t *testing.T) {
	api := notesAPI()
	cfg := vaultConfig(newFakeConnector(api))
	cfg.Refresh = 10 * time.Millisecond
	h := startTUI(t, cfg)
	h.waitFor("Refresh: every 10ms")
	eventually(t, "three listings", func() bool { return len(api.callsTo("list")) >= 3 })
	h.typeText("p")
	h.waitFor("Refresh: paused", "auto refresh paused")
	time.Sleep(30 * time.Millisecond)
	before := len(api.callsTo("list"))
	time.Sleep(50 * time.Millisecond)
	if after := len(api.callsTo("list")); after != before {
		t.Errorf("paused refresh listed %d more times", after-before)
	}
}

func TestErrorsFlashWithTheirHint(t *testing.T) {
	api := notesAPI()
	api.failNext("list", failure.New(failure.Forbidden, "GET /v1/vaults/x/notes", "no access").WithHint("ask an owner"))
	h := startTUI(t, vaultConfig(newFakeConnector(api)))
	h.waitFor("error: GET /v1/vaults/x/notes: no access (hint: ask an owner)")
}

func TestFlashClearsItselfAfterAWhile(t *testing.T) {
	cfg := vaultConfig(newFakeConnector(notesAPI()))
	cfg.FlashFor = 20 * time.Millisecond
	h := startTUI(t, cfg)
	h.command("pods")
	h.waitFor("unknown command")
	h.waitGone("unknown command")
}

func TestEscapeOnTheLastViewKeepsIt(t *testing.T) {
	h := startTUI(t, vaultConfig(newFakeConnector(notesAPI())))
	h.waitFor("work/plan.md")
	h.press(tcell.KeyEscape)
	h.onUI(func() {
		if got := h.shell.stack.crumbs(); !slices.Equal(got, []string{"notes"}) {
			t.Errorf("crumbs after Esc = %q, want [notes]", got)
		}
	})
}
