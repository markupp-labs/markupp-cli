package tui

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func TestParseCommandSplitsTheWordFromTheArgument(t *testing.T) {
	cases := map[string][2]string{
		"notes":            {"notes", ""},
		"  Notes  v1 ":     {"notes", "v1"},
		"graph a b":        {"graph", "a b"},
		"":                 {"", ""},
		"public my-garden": {"public", "my-garden"},
	}
	for line, want := range cases {
		word, arg := parseCommand(line)
		if word != want[0] || arg != want[1] {
			t.Errorf("parseCommand(%q) = %q, %q, want %q, %q", line, word, arg, want[0], want[1])
		}
	}
}

func TestFindCommandKnowsNamesAndAliases(t *testing.T) {
	for word, want := range map[string]string{"notes": "notes", "n": "notes", "v": "vaults", "g": "graph", "cm": "communities", "pub": "public"} {
		cmd, ok := findCommand(word)
		if !ok || cmd.name != want {
			t.Errorf("findCommand(%q) = %q, %v, want %q", word, cmd.name, ok, want)
		}
	}
	if _, ok := findCommand("pods"); ok {
		t.Error("findCommand(pods) found a command, want none")
	}
}

func TestCompletionsOfferTheCommandsWithThePrefix(t *testing.T) {
	if got, want := completions("no"), []string{"notes", "note", "notifications"}; !slices.Equal(got, want) {
		t.Errorf("completions(no) = %q, want %q", got, want)
	}
	if got := completions("x"); len(got) != 0 {
		t.Errorf("completions(x) = %q, want none", got)
	}
}

func TestEveryCommandHasAUsageAndAnOpener(t *testing.T) {
	for _, cmd := range commandTable() {
		if cmd.open == nil || !strings.HasPrefix(cmd.usage, ":"+cmd.name) {
			t.Errorf("command %q: usage %q, opener set %v", cmd.name, cmd.usage, cmd.open != nil)
		}
	}
}

func rows(cells ...[]string) []record {
	var records []record
	for _, c := range cells {
		records = append(records, record{cells: c})
	}
	return records
}

func firstCells(records []record) []string {
	var firsts []string
	for _, r := range records {
		firsts = append(firsts, r.cells[0])
	}
	return firsts
}

func TestSortRecordsComparesNumbersAsNumbers(t *testing.T) {
	records := rows([]string{"b", "10"}, []string{"a", "9"}, []string{"c", "-"})
	sortRecords(records, sortOrder{column: 1})
	if got := firstCells(records); !slices.Equal(got, []string{"c", "a", "b"}) {
		t.Errorf("ascending by number = %q, want [c a b]", got)
	}
	sortRecords(records, sortOrder{column: 0, descending: true})
	if got := firstCells(records); !slices.Equal(got, []string{"c", "b", "a"}) {
		t.Errorf("descending by text = %q, want [c b a]", got)
	}
}

func TestSortRecordsKeepsTheServerOrderWithoutAColumn(t *testing.T) {
	records := rows([]string{"b"}, []string{"a"})
	sortRecords(records, sortOrder{column: -1})
	if got := firstCells(records); !slices.Equal(got, []string{"b", "a"}) {
		t.Errorf("unsorted = %q, want [b a]", got)
	}
}

func TestSortOrderTogglesTheDirectionOfTheSameColumn(t *testing.T) {
	order := sortOrder{column: -1}.toggle(2)
	if order != (sortOrder{column: 2}) || order.arrow(2) != " ^" || order.arrow(1) != "" {
		t.Fatalf("first toggle = %+v", order)
	}
	if order = order.toggle(2); !order.descending || order.arrow(2) != " v" {
		t.Errorf("second toggle = %+v, want descending", order)
	}
	if order = order.toggle(0); order != (sortOrder{column: 0}) {
		t.Errorf("other column = %+v, want ascending on 0", order)
	}
}

func TestSortRuneIsTheInitialOfTheFirstColumnWithIt(t *testing.T) {
	columns := []string{"PATH", "UPDATED", "RANK", "PREDICATE"}
	if r, ok := sortRune(columns, 0); !ok || r != 'P' {
		t.Errorf("sortRune(PATH) = %q, %v, want P", r, ok)
	}
	if _, ok := sortRune(columns, 3); ok {
		t.Error("PREDICATE got a sort key, but PATH already has P")
	}
}

func TestFilterRecordsMatchesAnyCellIgnoringCase(t *testing.T) {
	records := rows([]string{"work/Meeting.md", "2"}, []string{"zebra.md", "12"})
	if got := firstCells(filterRecords(records, "meet")); !slices.Equal(got, []string{"work/Meeting.md"}) {
		t.Errorf("filter meet = %q", got)
	}
	if got := filterRecords(records, ""); len(got) != 2 {
		t.Errorf("empty filter kept %d rows, want 2", len(got))
	}
}

func TestClampRowKeepsTheSelectionBelowTheHeader(t *testing.T) {
	for _, c := range [][3]int{{0, 3, 1}, {2, 3, 2}, {9, 3, 3}, {5, 0, 5}} {
		if got := clampRow(c[0], c[1]); got != c[2] {
			t.Errorf("clampRow(%d, %d) = %d, want %d", c[0], c[1], got, c[2])
		}
	}
}

// staticView is a View with only a crumb, for the stack.
type staticView struct{ crumb string }

func (v staticView) Crumb() string { return v.crumb }

func (v staticView) Title() string { return v.crumb }

func (staticView) Widget() tview.Primitive { return nil }

func (staticView) Hotkeys() []Hotkey { return nil }

func (staticView) Load() {}

func TestStackKeepsItsLastView(t *testing.T) {
	var s stack
	if s.top() != nil || s.pop() {
		t.Fatal("an empty stack has a top or popped")
	}
	s.reset(staticView{"notes"})
	s.push(staticView{"note"})
	if got := s.crumbs(); !slices.Equal(got, []string{"notes", "note"}) {
		t.Errorf("crumbs = %q", got)
	}
	if !s.pop() || s.pop() || s.top().Crumb() != "notes" {
		t.Errorf("pop left %q, want notes kept", s.crumbs())
	}
}

func TestCrumbTextStartsWithTheScope(t *testing.T) {
	got := crumbText("Acme", []string{"notes", "note"})
	for _, part := range []string{" Acme ", " notes ", "[black:orange] note "} {
		if !strings.Contains(got, part) {
			t.Errorf("crumbText = %q, missing %q", got, part)
		}
	}
}

func TestTokenTextSaysWhenTheLoginExpires(t *testing.T) {
	soon, past := fixedNow.Add(90*time.Minute), fixedNow.Add(-time.Minute)
	cases := []struct {
		session Session
		want    string
	}{
		{Session{Credential: "env"}, "env"},
		{Session{Credential: "login", ExpiresAt: &soon}, "login, expires in 1h30m0s"},
		{Session{Credential: "login", ExpiresAt: &past}, "login, expired, renewed on the next request"},
	}
	for _, c := range cases {
		if got := tokenText(c.session, fixedNow); got != c.want {
			t.Errorf("tokenText(%+v) = %q, want %q", c.session, got, c.want)
		}
	}
}

func TestHotkeyColumnsLeaveOutTheMinorKeys(t *testing.T) {
	sortKey := runeKey('P', "sort PATH", nil)
	sortKey.minor = true
	keys := []Hotkey{runeKey('e', "edit", nil), ctrlKey(tcell.KeyCtrlD, "delete", nil), runeKey('y', "copy id", nil), sortKey}
	got := hotkeyColumns(keys, 2)
	lines := strings.Split(got, "\n")
	if len(lines) != 2 || !strings.Contains(lines[0], "<e>") || !strings.Contains(lines[0], "<y>") || !strings.Contains(lines[1], "<ctrl-d>") {
		t.Errorf("hotkeyColumns = %q", got)
	}
	if strings.Contains(got, "sort") {
		t.Errorf("hotkeyColumns shows a minor key: %q", got)
	}
}

func TestHotkeyMatchesItsKeyAndRune(t *testing.T) {
	key := runeKey('e', "edit", nil)
	if !key.matches(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone)) || key.matches(tcell.NewEventKey(tcell.KeyRune, 'E', tcell.ModNone)) {
		t.Error("runeKey e does not match e alone")
	}
	if got := ctrlKey(tcell.KeyCtrlD, "delete", nil).Name(); got != "<ctrl-d>" {
		t.Errorf("Name = %q, want <ctrl-d>", got)
	}
}

func TestOSC52EncodesTheTextInBase64(t *testing.T) {
	if got, want := osc52("n1"), "\x1b]52;c;bjE=\a"; got != want {
		t.Errorf("osc52 = %q, want %q", got, want)
	}
}

func TestErrorTextCarriesTheHint(t *testing.T) {
	err := failure.New(failure.Unauthenticated, "https://x", "not logged in").WithHint("run mu login https://x")
	if got, want := errorText(err), "error: https://x: not logged in (hint: run mu login https://x)"; got != want {
		t.Errorf("errorText = %q, want %q", got, want)
	}
	if got := errorText(errors.New("boom")); got != "error: boom" {
		t.Errorf("errorText = %q", got)
	}
}

func TestPlainEscapesStyleTagsAndControlCharacters(t *testing.T) {
	if got, want := plain("a\x1b[31m [red]b\nc"), "a[31m [red[]b c"; got != want {
		t.Errorf("plain = %q, want %q", got, want)
	}
}
