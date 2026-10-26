package client

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func requestError(t *testing.T, reply http.HandlerFunc) *failure.Error {
	t.Helper()
	server := newFakeServer(t, reply)
	c := newTestClient(t, server, "")
	err := c.exec(context.Background(), request{method: http.MethodGet, path: endpoint("notes", "n1")})
	failed, ok := err.(*failure.Error)
	if !ok {
		t.Fatalf("exec error = %#v, want *failure.Error", err)
	}
	return failed
}

func TestStatusMapsToKind(t *testing.T) {
	table := map[int]failure.Kind{
		400: failure.Usage, 401: failure.Unauthenticated, 403: failure.Forbidden,
		404: failure.NotFound, 409: failure.Conflict, 422: failure.Usage,
		429: failure.RateLimited, 500: failure.Unavailable, 503: failure.Unavailable,
		418: failure.Internal,
	}
	for status, want := range table {
		got := requestError(t, replyJSON(status, `{"error":"code_x","message":"server says"}`))
		if got.Kind != want || got.Code != "code_x" || got.Message != "server says" {
			t.Errorf("status %d mapped to %+v, want kind %s with code and message", status, got, want)
		}
		if got.Subject != "GET /v1/notes/n1" {
			t.Errorf("status %d subject = %q", status, got.Subject)
		}
	}
}

func TestPlainTextNotFoundIsAnAbsentEndpoint(t *testing.T) {
	got := requestError(t, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	if got.Kind != failure.NotFound || got.Code != "endpoint_absent" || !strings.Contains(got.Hint, "server") {
		t.Errorf("plain 404 mapped to %+v, want not_found endpoint_absent with a hint", got)
	}
}

func TestJSONNotFoundKeepsTheServerCode(t *testing.T) {
	got := requestError(t, replyJSON(404, `{"error":"note_not_found","message":"no note n1"}`))
	if got.Code != "note_not_found" {
		t.Errorf("code = %q, want note_not_found", got.Code)
	}
}

func TestInsufficientScopeCarriesTheScope(t *testing.T) {
	got := requestError(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Bearer error="insufficient_scope", scope="notes:write"`)
		replyJSON(403, `{"error":"insufficient_scope","message":"missing scope"}`)(w, r)
	})
	if got.Kind != failure.Forbidden || got.Scope != "notes:write" || !strings.Contains(got.Hint, "notes:write") {
		t.Errorf("insufficient scope mapped to %+v", got)
	}
}

func TestFirstPartyOnlyPointsToThePanel(t *testing.T) {
	got := requestError(t, replyJSON(403, `{"error":"first_party_only","message":"login token only"}`))
	if got.Kind != failure.Forbidden || !strings.Contains(got.Hint, "web panel") {
		t.Errorf("first_party_only mapped to %+v, want a hint about the web panel", got)
	}
}

func TestRateLimitCarriesRetryAfter(t *testing.T) {
	got := requestError(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "42")
		replyJSON(429, `{"error":"rate_limited","message":"slow down"}`)(w, r)
	})
	if got.RetryAfter != 42*time.Second || !strings.Contains(got.Hint, "42") {
		t.Errorf("rate limit mapped to %+v, want RetryAfter 42s", got)
	}
}

func TestUnauthorizedHintsLogin(t *testing.T) {
	got := requestError(t, replyJSON(401, `{"error":"unauthorized","message":"expired"}`))
	if !strings.Contains(got.Hint, "mu login") {
		t.Errorf("hint = %q, want one naming mu login", got.Hint)
	}
}

func TestServerErrorWithoutBodyUsesTheStatus(t *testing.T) {
	got := requestError(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})
	if got.Kind != failure.Unavailable || !strings.Contains(got.Message, "502") {
		t.Errorf("502 mapped to %+v, want unavailable naming the status", got)
	}
}

func TestUnreachableServerIsUnavailable(t *testing.T) {
	server := newFakeServer(t, replyJSON(200, `{}`))
	c := newTestClient(t, server, "")
	server.Close()
	err := c.exec(context.Background(), request{method: http.MethodGet, path: "/x"})
	if failure.KindOf(err) != failure.Unavailable {
		t.Errorf("unreachable server error = %v, want unavailable", err)
	}
}

func TestCanceledContextIsAnInterrupt(t *testing.T) {
	server := newFakeServer(t, replyJSON(200, `{}`))
	c := newTestClient(t, server, "")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := c.exec(ctx, request{method: http.MethodGet, path: "/x"})
	if failure.KindOf(err) != failure.Interrupted {
		t.Errorf("canceled request error = %v, want interrupted", err)
	}
}
