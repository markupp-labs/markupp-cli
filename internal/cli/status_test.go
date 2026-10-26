package cli

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

// fakeAccount serves fixed bodies for the account reads.
type fakeAccount struct {
	health, user, tenants string
	healthErr             error
}

func (f *fakeAccount) GetHealth(context.Context) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(f.health)), f.healthErr
}

func (f *fakeAccount) GetAuthenticatedUser(context.Context) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(f.user)), nil
}

func (f *fakeAccount) ListMyTenants(context.Context) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(f.tenants)), nil
}

func statusMu() *testMu {
	mu := newTestMu()
	mu.store.settings = config.Settings{URL: "https://markupp.example.com", Auth: &config.Credential{
		Issuer: "https://markupp.example.com", AccessToken: "secret-access", RefreshToken: "secret-refresh",
		ExpiresAt: loginExpiry, Scope: "notas:leitura notas:escrita administracao",
	}}
	mu.server.services = Services{Account: &fakeAccount{
		health:  `{"status":"ok"}`,
		user:    `{"id":"0b3c4f8e-2f5a-4d7b-9c1e-8a6f5d4c3b2a","email":"ana@example.com","email_verified":true,"password_login":true}`,
		tenants: `[{"id":"5f1d7c2a-9b3e-4a6f-8d2c-1e7b9a4f6c3d","name":"Acme","role":"owner","notes_count":42,"last_note_at":"2026-10-20T09:30:00Z"},{"id":"a2c4e6f8-1b3d-4f5a-9c7e-2d4f6a8c0e1b","name":"Side project","role":"viewer","notes_count":0,"last_note_at":null}]`,
	}}
	return mu
}

func TestStatusText(t *testing.T) {
	run := statusMu().run(t, "status")
	run.expectCode(t, 0)
	outputtest.Golden(t, "status", []byte(run.stdout))
	if strings.Contains(run.stdout+run.stderr, "secret") {
		t.Errorf("status printed a token:\n%s%s", run.stdout, run.stderr)
	}
}

func TestStatusJSON(t *testing.T) {
	run := statusMu().run(t, "status", "--json")
	run.expectCode(t, 0)
	outputtest.Golden(t, "status_json", []byte(run.stdout))
	if strings.Contains(run.stdout, "secret") {
		t.Errorf("status --json printed a token: %s", run.stdout)
	}
}

func TestWhoamiIsStatus(t *testing.T) {
	run := statusMu().run(t, "whoami")
	run.expectCode(t, 0)
	outputtest.Golden(t, "status", []byte(run.stdout))
}

func TestStatusReadsHealthWithoutCredentials(t *testing.T) {
	mu := statusMu()
	mu.run(t, "status").expectCode(t, 0)
	if len(mu.server.connections) != 2 || !mu.server.connections[0].Public || mu.server.connections[1].Public {
		t.Errorf("connections = %+v, want a public one for the health and then an authenticated one", mu.server.connections)
	}
}

func TestStatusWithMarkuppTokenShowsTheEnvironment(t *testing.T) {
	mu := statusMu()
	mu.env[config.EnvToken] = "secret-raw"
	run := mu.run(t, "status")
	run.expectCode(t, 0)
	if !strings.Contains(run.stdout, config.EnvToken) || strings.Contains(run.stdout, "secret") {
		t.Errorf("stdout %q, want the token source without the token", run.stdout)
	}
}

func TestStatusNotLoggedInPrintsWhatItKnowsAndExits3(t *testing.T) {
	mu := statusMu()
	mu.store.settings.Auth = nil
	mu.server.privateErr = failure.New(failure.Unauthenticated, "https://markupp.example.com", "not logged in").WithHint("run mu login https://markupp.example.com")
	run := mu.run(t, "status")
	run.expectCode(t, 3)
	outputtest.Golden(t, "status_logged_out", []byte(run.stdout))
	if !strings.Contains(run.stderr, "mu login") {
		t.Errorf("stderr %q, want the login hint", run.stderr)
	}
}

func TestStatusWithServerDownExits8(t *testing.T) {
	mu := statusMu()
	mu.server.services.Account.(*fakeAccount).healthErr = failure.New(failure.Unavailable, "GET /v1/healthz", "server answered 503")
	run := mu.run(t, "status")
	run.expectCode(t, 8)
	if !strings.Contains(run.stdout, "https://markupp.example.com") {
		t.Errorf("stdout %q, want at least the server", run.stdout)
	}
}

func TestStatusWithoutServerIsAUsageError(t *testing.T) {
	mu := newTestMu()
	run := mu.run(t, "status")
	run.expectCode(t, 2)
	if run.stdout != "" || len(mu.server.connections) != 0 {
		t.Errorf("stdout %q, connections %d, want nothing", run.stdout, len(mu.server.connections))
	}
}
