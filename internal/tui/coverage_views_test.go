package tui

import (
	"slices"
	"testing"

	"github.com/gdamore/tcell/v2"
)

const relationNote = `{"id":"` + noteC + `","title":"plan","path":"work/plan.md"}`

func TestRelationViewsShowTheirRows(t *testing.T) {
	cases := []struct {
		key, method, body string
		want              []string
	}{
		{"b", "backlinks", `[{"source":` + relationNote + `,"target":null,"cited_title":"meeting","type":"typed","predicate":"author","inferred_target":null}]`,
			[]string{"backlinks(work/meeting.md)[1]", "typed", "work/plan.md", "author"}},
		{"m", "mentions", `[{"note":` + relationNote + `,"cited_title":"Plan","prior":0.5}]`,
			[]string{"mentions(work/meeting.md)[1]", "work/plan.md", "Plan", "0.5"}},
		{"s", "suggestions", `[{"note":` + relationNote + `,"score":0.75,"resource_allocation":0.1,"cosine":0.25}]`,
			[]string{"suggestions(work/meeting.md)[1]", "work/plan.md", "0.75", "0.25"}},
	}
	for _, tc := range cases {
		api := notesAPI()
		api.bodies[tc.method] = tc.body
		h := startOnNotes(t, api)
		h.typeText(tc.key)
		h.waitFor(tc.want...)
		h.stop()
	}
}

func TestTenantKeyListsTheTenantsAndSwitches(t *testing.T) {
	connector := newFakeConnector(notesAPI())
	h := startTUI(t, vaultConfig(connector))
	h.waitFor("work/plan.md")
	h.typeText("t")
	h.waitFor("tenants[1]", "Acme", "owner")
	h.press(tcell.KeyEnter)
	h.waitFor("switched to Acme", "Tenant:  Acme")
	if got := connector.connected(); got[len(got)-1] != (Scope{Tenant: tenantA}) {
		t.Errorf("scopes = %+v, want the tenant last", got)
	}
}

func TestCopyOnARowThatIsNotANote(t *testing.T) {
	connector := newFakeConnector(notesAPI())
	h := startTUI(t, vaultConfig(connector))
	h.waitFor("work/plan.md")
	h.typeText("t")
	h.waitFor("tenants[1]", "Acme")
	h.typeText("y")
	h.waitFor("copied " + tenantA)
}

func TestGraphCommandOpensTheTreeOfTheNote(t *testing.T) {
	api := notesAPI()
	h := startOnNotes(t, api)
	h.command("graph " + noteA)
	h.waitFor("graph("+noteA+")", "-> explicit  work/plan.md")
	if got := api.callsTo("graph"); !slices.Equal(got, []string{noteA}) {
		t.Errorf("graph calls = %q, want %s", got, noteA)
	}
	h.command("graph")
	h.waitFor("missing note id, expected :graph <note id>")
}

func TestGraphNoteActionsFollowTheCurrentNode(t *testing.T) {
	api := notesAPI()
	h := startOnNotes(t, api)
	h.typeText("g")
	h.waitFor("graph(work/meeting.md)", "-> explicit  Budget (unresolved)")
	h.press(tcell.KeyDown)
	h.press(tcell.KeyDown)
	h.typeText("v")
	h.waitFor("unresolved title, expected a node that is a note")
	h.press(tcell.KeyUp)
	h.typeText("v")
	eventually(t, "the neighbor opened", func() bool { return slices.Contains(api.callsTo("get"), noteC) })
}
