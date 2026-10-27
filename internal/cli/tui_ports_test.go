package cli

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/schema"
	"github.com/markupp-labs/markupp-cli/internal/tui"
)

// connectTUI starts mu tui on mu and connects the interface to scope.
func connectTUI(t *testing.T, mu *testMu, scope tui.Scope) tui.Ports {
	t.Helper()
	mu.run(t, "tui").expectCode(t, 0)
	ports, err := mu.tui.onlyConfig(t).Connect(context.Background(), scope)
	if err != nil {
		t.Fatalf("Connect(%+v) = %v", scope, err)
	}
	return ports
}

// adminServices are the fakes of the vault, team and notification ports.
type adminServices struct {
	vaults        *fakeVaults
	vaultMembers  *fakeVaultMembers
	team          *fakeAdminServer
	notifications *fakeNotifications
}

func newAdminServices() adminServices {
	return adminServices{&fakeVaults{}, &fakeVaultMembers{}, &fakeAdminServer{}, &fakeNotifications{}}
}

func (a adminServices) services() Services {
	return Services{
		Vaults: a.vaults, VaultMembers: a.vaultMembers, Members: a.team, Invites: a.team,
		Notifications: a.notifications,
	}
}

func TestTUIConnectorWiresVaultsTeamAndNotifications(t *testing.T) {
	mu, fakes := tuiMu(), newAdminServices()
	mu.server.services = fakes.services()
	ports := connectTUI(t, mu, tui.Scope{Tenant: "t2"})
	if ports.Vaults == nil || ports.Team == nil || ports.Notifications == nil {
		t.Fatalf("ports = %+v, want vaults, team and notifications", ports)
	}
	ctx := context.Background()
	_, errVaults := ports.Vaults.ListVaults(ctx)
	_, errVaultMembers := ports.Vaults.ListVaultMembers(ctx, "v1")
	_, errMembers := ports.Team.ListMembers(ctx)
	_, errInvites := ports.Team.ListMyInvites(ctx)
	_, errFeed := ports.Notifications.Notifications(ctx)
	if err := errors.Join(errVaults, errVaultMembers, errMembers, errInvites, errFeed); err != nil {
		t.Fatal(err)
	}
	got := [][]string{fakes.vaults.calls, fakes.vaultMembers.calls, fakes.team.calls, fakes.notifications.calls}
	want := [][]string{{"list"}, {"list v1"}, {"ListMembers", "ListMyInvites"}, {"feed"}}
	if !slices.EqualFunc(got, want, slices.Equal) {
		t.Errorf("calls = %q, want %q", got, want)
	}
}

func TestTUIConnectorNamesTheFeatureAServerLacks(t *testing.T) {
	mu, fakes := tuiMu(), newAdminServices()
	fakes.vaults.err, fakes.notifications.err = errEndpointAbsent(), errEndpointAbsent()
	mu.server.services = fakes.services()
	ports := connectTUI(t, mu, tui.Scope{})
	if ports.Vaults == nil || ports.Notifications == nil {
		t.Fatalf("ports = %+v, want vaults and notifications", ports)
	}
	_, errVaults := ports.Vaults.ListVaults(context.Background())
	_, errFeed := ports.Notifications.Notifications(context.Background())
	if !strings.Contains(failure.HintOf(errVaults), featureVaults) || !strings.Contains(failure.HintOf(errFeed), featureNotifications) {
		t.Errorf("hints = %q and %q, want the features named", failure.HintOf(errVaults), failure.HintOf(errFeed))
	}
}

func TestTUIConnectorLeavesOutAHalfWiredPort(t *testing.T) {
	mu := tuiMu()
	mu.server.services = Services{Vaults: &fakeVaults{}, Members: &fakeAdminServer{}}
	ports := connectTUI(t, mu, tui.Scope{})
	if ports.Vaults != nil || ports.Team != nil || ports.Notifications != nil {
		t.Errorf("ports = %+v, want no vaults, team or notifications", ports)
	}
}

func TestTUIConnectorReadsAPublicVaultWithoutCredentials(t *testing.T) {
	mu, notes := tuiMu(), newFakeNotes(map[string]string{"public list": "[]", "public get": "{}", "public search": "[]"})
	mu.server.services = Services{Public: notes, Notes: newFakeNotes(nil), Vaults: &fakeVaults{}}
	ports := connectTUI(t, mu, tui.Scope{PublicSlug: "garden"})
	if ports.Notes == nil || ports.Vaults != nil || ports.Account != nil || ports.Team != nil {
		t.Fatalf("ports = %+v, want the public notes only", ports)
	}
	if conn := mu.server.connections[len(mu.server.connections)-1]; !conn.Public {
		t.Errorf("connection = %+v, want a public one", conn)
	}
	ctx, page := context.Background(), client.Page{Limit: 100}
	_, errList := ports.Notes.ListNotes(ctx, "garden", client.NoteFilter{}, page)
	_, errGet := ports.Notes.GetNote(ctx, "n1")
	_, errSearch := ports.Notes.SearchNotes(ctx, "kubernetes", "garden", page)
	if err := errors.Join(errList, errGet, errSearch); err != nil {
		t.Fatal(err)
	}
	want := []noteCall{
		{method: "public list", slug: "garden", filter: noFilter, page: page},
		{method: "public get", slug: "garden", id: "n1"},
		{method: "public search", slug: "garden", text: "kubernetes", page: page},
	}
	if !slices.Equal(notes.calls, want) {
		t.Errorf("calls = %+v, want %+v", notes.calls, want)
	}
}

func TestTUIPublicVaultIsReadOnly(t *testing.T) {
	mu, notes := tuiMu(), newFakeNotes(nil)
	mu.server.services = Services{Public: notes}
	ports := connectTUI(t, mu, tui.Scope{PublicSlug: "garden"})
	_, errUpdate := ports.Notes.UpdateNote(context.Background(), "n1", schema.NoteUpdateRequest{Content: "x"})
	errDelete := ports.Notes.DeleteNote(context.Background(), "n1")
	for _, err := range []error{errUpdate, errDelete} {
		if failure.KindOf(err) != failure.Forbidden || !strings.Contains(err.Error(), "garden") {
			t.Errorf("write on a public vault = %v, want forbidden naming the slug", err)
		}
	}
	if len(notes.calls) != 0 {
		t.Errorf("calls = %+v, want none", notes.calls)
	}
}

func TestTUIPublicVaultNamesTheMissingFeature(t *testing.T) {
	mu, notes := tuiMu(), newFakeNotes(nil)
	notes.errs["public list"] = errEndpointAbsent()
	mu.server.services = Services{Public: notes}
	ports := connectTUI(t, mu, tui.Scope{PublicSlug: "garden"})
	_, err := ports.Notes.ListNotes(context.Background(), "garden", client.NoteFilter{}, client.Page{})
	if !strings.Contains(failure.HintOf(err), featurePublication) {
		t.Errorf("hint = %q, want the feature named", failure.HintOf(err))
	}
}
