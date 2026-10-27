package cli

import (
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func TestLogoutConfirmsOnStderr(t *testing.T) {
	mu := newTestMu()
	mu.auth.removed = true
	run := mu.run(t, "logout")
	run.expectCode(t, 0)
	if run.stdout != "" || !strings.Contains(run.stderr, "Logged out") || mu.auth.logouts != 1 {
		t.Errorf("stdout %q, stderr %q, logouts %d, want one logout confirmed on stderr", run.stdout, run.stderr, mu.auth.logouts)
	}
}

func TestLogoutWithoutLoginSucceeds(t *testing.T) {
	run := newTestMu().run(t, "logout")
	run.expectCode(t, 0)
	if !strings.Contains(run.stderr, "Not logged in") {
		t.Errorf("stderr %q, want a note that there was no login", run.stderr)
	}
}

func TestLogoutWithFailedRevokeExitsWithItsKind(t *testing.T) {
	mu := newTestMu()
	mu.auth.removed = true
	mu.auth.err = failure.New(failure.Unavailable, "https://markupp.example.com", "revoking failed").WithHint("revoke it in the web panel")
	run := mu.run(t, "logout")
	run.expectCode(t, 8)
	if !strings.Contains(run.stderr, "web panel") {
		t.Errorf("stderr %q, want the hint", run.stderr)
	}
}
