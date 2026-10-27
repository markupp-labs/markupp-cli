package cli

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/tui"
)

// fakeTUI is the interactive interface of a test: it records the Config it got and
// returns err instead of drawing.
type fakeTUI struct {
	configs []tui.Config
	err     error
}

func (f *fakeTUI) run(_ context.Context, cfg tui.Config) error {
	f.configs = append(f.configs, cfg)
	return f.err
}

// tuiMu is a mu with a server, a tenant and a vault configured.
func tuiMu() *testMu {
	mu := newTestMu()
	mu.env["MARKUPP_URL"] = "https://markupp.example.com"
	mu.env["MARKUPP_TENANT"] = "t1"
	mu.store.settings.Vault = "v-file"
	return mu
}

func TestTUIStartsWithTheResolvedSettings(t *testing.T) {
	mu := tuiMu()
	mu.env["EDITOR"] = "vim -n"
	run := mu.run(t, "tui", "--vault", "v1", "--refresh", "30s")
	run.expectCode(t, 0)
	if run.stdout != "" {
		t.Errorf("stdout = %q, want nothing", run.stdout)
	}
	cfg := mu.tui.onlyConfig(t)
	if cfg.Vault != "v1" || cfg.Refresh != 30*time.Second || cfg.Scope != (tui.Scope{Tenant: "t1"}) || cfg.Editor != "vim -n" {
		t.Errorf("config = vault %q, refresh %s, scope %+v, editor %q", cfg.Vault, cfg.Refresh, cfg.Scope, cfg.Editor)
	}
	if cfg.Terminal.Out == nil || cfg.Connect == nil || cfg.Session == nil {
		t.Error("the config lacks the terminal, the connector or the session")
	}
}

func (f *fakeTUI) onlyConfig(t *testing.T) tui.Config {
	t.Helper()
	if len(f.configs) != 1 {
		t.Fatalf("the interface ran %d times, want once", len(f.configs))
	}
	return f.configs[0]
}

func TestTUIPrefersVisualToEditor(t *testing.T) {
	mu := tuiMu()
	mu.env["EDITOR"], mu.env["VISUAL"] = "vim", "code -w"
	mu.run(t, "tui").expectCode(t, 0)
	if got := mu.tui.onlyConfig(t).Editor; got != "code -w" {
		t.Errorf("editor = %q, want $VISUAL", got)
	}
}

func TestTUIWithoutAServerExplainsHowToConfigureOne(t *testing.T) {
	mu := newTestMu()
	run := mu.run(t, "tui")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, "no Markupp server configured") || len(mu.tui.configs) != 0 {
		t.Errorf("stderr = %q, runs = %d", run.stderr, len(mu.tui.configs))
	}
}

func TestTUIRefusesANegativeRefresh(t *testing.T) {
	run := tuiMu().run(t, "tui", "--refresh", "-1s")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, "--refresh: invalid interval -1s") {
		t.Errorf("stderr = %q, want the interval named", run.stderr)
	}
}

func TestTUIErrorKeepsItsExitCode(t *testing.T) {
	mu := tuiMu()
	mu.tui.err = failure.New(failure.Unauthenticated, "https://markupp.example.com", "not logged in")
	mu.run(t, "tui").expectCode(t, 3)
}

func TestTUIConnectorSwitchesTheTenant(t *testing.T) {
	mu := tuiMu()
	mu.server.services = Services{Notes: newFakeNotes(nil), Graph: &fakeGraph{}}
	mu.run(t, "tui").expectCode(t, 0)
	ports, err := mu.tui.onlyConfig(t).Connect(context.Background(), tui.Scope{Tenant: "t2"})
	if err != nil {
		t.Fatalf("Connect = %v", err)
	}
	if ports.Notes == nil || ports.Graph == nil || ports.Relations != nil || ports.Vaults != nil {
		t.Errorf("ports = %+v, want notes and graph only", ports)
	}
	conn := mu.server.connections[len(mu.server.connections)-1]
	if conn.Resolved.Tenant.Value != "t2" || conn.Public {
		t.Errorf("connection = tenant %q, public %v, want t2 with credentials", conn.Resolved.Tenant.Value, conn.Public)
	}
}

func TestTUISessionShowsTheLoginExpiryNeverTheToken(t *testing.T) {
	mu := tuiMu()
	expires := time.Date(2026, 5, 15, 13, 0, 0, 0, time.UTC)
	mu.store.settings.Auth = &config.Credential{AccessToken: "secret", ExpiresAt: expires}
	mu.run(t, "tui").expectCode(t, 0)
	session := mu.tui.onlyConfig(t).Session()
	if session.Server != "https://markupp.example.com" || session.Credential != "login" || !session.ExpiresAt.Equal(expires) {
		t.Errorf("session = %+v", session)
	}
}

func TestTUIGraphUsesTheServerDefaults(t *testing.T) {
	graph := &fakeGraph{recordingPort{body: "{}"}}
	adapter, ctx, page := tuiGraph{graph: graph}, context.Background(), client.Page{Limit: 5}
	_, errGraph := adapter.NoteGraph(ctx, "n1", 1)
	_, errCommunities := adapter.VaultCommunities(ctx, "v1", page)
	_, errLinks := adapter.VaultLinks(ctx, "v1", page)
	if err := errors.Join(errGraph, errCommunities, errLinks); err != nil {
		t.Fatal(err)
	}
	want := []portCall{
		{method: "NoteGraph", subject: "n1", depth: 1},
		{method: "VaultCommunities", subject: "v1", page: page},
		{method: "VaultLinks", subject: "v1", page: page},
	}
	if !slices.Equal(graph.calls, want) {
		t.Errorf("calls = %+v, want %+v", graph.calls, want)
	}
}
