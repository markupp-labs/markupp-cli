package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/session"
)

var loginExpiry = time.Date(2026, 10, 26, 13, 0, 0, 0, time.UTC)

func loggedInMu() *testMu {
	mu := newTestMu()
	mu.auth.cred = config.Credential{Issuer: "https://markupp.example.com", AccessToken: "secret-access", ExpiresAt: loginExpiry}
	return mu
}

func TestLoginWithURLLogsInThere(t *testing.T) {
	mu := loggedInMu()
	run := mu.run(t, "login", "https://markupp.example.com/")
	run.expectCode(t, 0)
	if len(mu.auth.servers) != 1 || mu.auth.servers[0] != "https://markupp.example.com" {
		t.Errorf("logged in to %q, want the normalized URL", mu.auth.servers)
	}
	if run.stdout != "" || !strings.Contains(run.stderr, "Logged in to https://markupp.example.com") || !strings.Contains(run.stderr, "2026-10-26T13:00:00Z") {
		t.Errorf("stdout %q, stderr %q, want the confirmation on stderr only", run.stdout, run.stderr)
	}
	if strings.Contains(run.stderr, "secret-access") {
		t.Errorf("stderr %q shows the token", run.stderr)
	}
}

func TestLoginWithoutURLUsesTheConfiguredServer(t *testing.T) {
	mu := loggedInMu()
	mu.env[config.EnvURL] = "https://env.example.com"
	mu.run(t, "login").expectCode(t, 0)
	if len(mu.auth.servers) != 1 || mu.auth.servers[0] != "https://env.example.com" {
		t.Errorf("logged in to %q, want the server of %s", mu.auth.servers, config.EnvURL)
	}
}

func TestLoginWithoutAnyServerIsAUsageError(t *testing.T) {
	mu := loggedInMu()
	run := mu.run(t, "login")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, "mu login") || len(mu.auth.servers) != 0 {
		t.Errorf("stderr %q, want a hint with mu login and no login attempt", run.stderr)
	}
}

func TestLoginRefusesAnInvalidURL(t *testing.T) {
	mu := loggedInMu()
	run := mu.run(t, "login", "ftp://markupp.example.com")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, "ftp://markupp.example.com") {
		t.Errorf("stderr %q, want the offending URL", run.stderr)
	}
}

func TestLoginTakesAtMostOneURL(t *testing.T) {
	loggedInMu().run(t, "login", "https://a.example.com", "https://b.example.com").expectCode(t, 2)
}

func TestLoginPassesTheSetup(t *testing.T) {
	mu := loggedInMu()
	mu.run(t, "login", "--allow-plain-http", "http://markupp.internal").expectCode(t, 0)
	setup := mu.auth.setups[0]
	if !setup.AllowPlainHTTP || setup.Store != mu.store || setup.Notice == nil {
		t.Errorf("setup = %+v, want the plain http opt in, the store and a notice writer", setup)
	}
}

func TestLoginWarnsThatMarkuppTokenWins(t *testing.T) {
	mu := loggedInMu()
	mu.env[config.EnvToken] = "raw"
	run := mu.run(t, "login", "https://markupp.example.com")
	run.expectCode(t, 0)
	if !strings.Contains(run.stderr, config.EnvToken) {
		t.Errorf("stderr %q, want a warning about %s", run.stderr, config.EnvToken)
	}
}

func TestLoginFailureKeepsItsKind(t *testing.T) {
	mu := loggedInMu()
	mu.auth.err = failure.New(failure.Unauthenticated, "http://127.0.0.1:1/callback", "the server refused the login")
	run := mu.run(t, "login", "https://markupp.example.com")
	run.expectCode(t, 3)
	if !strings.Contains(run.stderr, "refused the login") {
		t.Errorf("stderr %q, want the cause", run.stderr)
	}
}

func TestSessionAuthenticatorCarriesTheSetup(t *testing.T) {
	store := &fakeConfigStore{path: "/home/u/.config/markupp/cli.json"}
	var notice strings.Builder
	built := newSessionAuthenticator(AuthSetup{Store: store, Notice: &notice})
	auth, ok := built.(session.Authenticator)
	if !ok {
		t.Fatalf("newSessionAuthenticator built %T, want a session.Authenticator", built)
	}
	if auth.Store != store || auth.Notice != &notice || auth.OAuth == nil {
		t.Errorf("authenticator = %+v, want the store, the notice writer and an OAuth client", auth)
	}
}
