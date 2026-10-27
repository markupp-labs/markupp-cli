package client

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func readAll(t *testing.T, body io.ReadCloser) string {
	t.Helper()
	defer func() { _ = body.Close() }()
	raw, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	return string(raw)
}

func TestEveryRequestGoesUnderV1WithAuthAndTenant(t *testing.T) {
	server := newFakeServer(t, replyJSON(200, `{"ok":true}`))
	c := newTestClient(t, server, "t1")
	body, err := c.open(context.Background(), request{method: http.MethodGet, path: endpoint("notes", "n1")})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if got := readAll(t, body); got != `{"ok":true}` {
		t.Errorf("body = %q", got)
	}
	want := recordedRequest{method: "GET", path: "/v1/notes/n1", auth: "Bearer tok", tenant: "t1", agent: "mu/test"}
	if got := server.seen()[0]; got != want {
		t.Errorf("request = %+v, want %+v", got, want)
	}
}

func TestBaseURLPathIsKeptBeforeV1(t *testing.T) {
	server := newFakeServer(t, replyJSON(204, ""))
	c, err := New(Options{BaseURL: server.URL + "/markupp/", Tokens: StaticToken("tok")})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.exec(context.Background(), request{method: http.MethodDelete, path: endpoint("notes", "n1")}); err != nil {
		t.Fatalf("exec: %v", err)
	}
	got := server.seen()[0]
	if got.path != "/markupp/v1/notes/n1" || got.tenant != "" {
		t.Errorf("request = %+v, want path /markupp/v1/notes/n1 and no tenant", got)
	}
}

func TestQueryAndBodyAreSent(t *testing.T) {
	server := newFakeServer(t, replyJSON(201, `{}`))
	c := newTestClient(t, server, "")
	query := url.Values{}
	Page{Limit: 5, Offset: 10}.apply(query)
	body, err := c.open(context.Background(), request{
		method: http.MethodPost, path: endpoint("vaults", "v1", "notes"), query: query,
		body: []byte(`{"path":"a.md"}`),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	_ = body.Close()
	got := server.seen()[0]
	if got.query != "limit=5&offset=10" || got.body != `{"path":"a.md"}` {
		t.Errorf("request = %+v", got)
	}
}

func TestEndpointEscapesEverySegment(t *testing.T) {
	cases := map[string][]string{
		"/notes/a%2Fb":           {"notes", "a/b"},
		"/notes/%2E%2E/links":    {"notes", "..", "links"},
		"/notes/a%20b%3Fc":       {"notes", "a b?c"},
		"/public/my-vault/notes": {"public", "my-vault", "notes"},
	}
	for want, segments := range cases {
		if got := endpoint(segments...); got != want {
			t.Errorf("endpoint(%q) = %q, want %q", segments, got, want)
		}
	}
}

func TestEscapedSegmentReachesTheServerUnchanged(t *testing.T) {
	server := newFakeServer(t, replyJSON(204, ""))
	c := newTestClient(t, server, "")
	if err := c.exec(context.Background(), request{method: http.MethodGet, path: endpoint("notes", "../x")}); err != nil {
		t.Fatalf("exec: %v", err)
	}
	if got := server.seen()[0].path; got != "/v1/notes/..%2Fx" && got != "/v1/notes/%2E%2E%2Fx" {
		t.Errorf("path = %q, want the id escaped as one segment", got)
	}
}

func TestNewPublicSendsNoCredentials(t *testing.T) {
	server := newFakeServer(t, replyJSON(204, ""))
	c, err := NewPublic(Options{BaseURL: server.URL, Tenant: "t1", Tokens: StaticToken("tok")})
	if err != nil {
		t.Fatalf("NewPublic: %v", err)
	}
	if err := c.exec(context.Background(), request{method: http.MethodGet, path: endpoint("public", "s")}); err != nil {
		t.Fatalf("exec: %v", err)
	}
	if got := server.seen()[0]; got.auth != "" || got.tenant != "" {
		t.Errorf("public request carried auth %q tenant %q, want none", got.auth, got.tenant)
	}
}

func TestNewWithoutTokensIsUnauthenticated(t *testing.T) {
	_, err := New(Options{BaseURL: "https://markupp.example.com"})
	if failure.KindOf(err) != failure.Unauthenticated {
		t.Fatalf("New without tokens error = %v, want unauthenticated", err)
	}
}

func TestCheckBaseURL(t *testing.T) {
	allowed := []string{
		"https://markupp.example.com", "http://localhost:8080", "http://127.0.0.1:8080",
		"http://[::1]:8080", "http://127.0.0.2",
	}
	for _, raw := range allowed {
		if _, err := CheckBaseURL(raw, false); err != nil {
			t.Errorf("CheckBaseURL(%q) = %v, want ok", raw, err)
		}
	}
	if _, err := CheckBaseURL("http://markupp.example.com", true); err != nil {
		t.Errorf("CheckBaseURL with allowPlainHTTP = %v, want ok", err)
	}
}

func TestCheckBaseURLRefusesPlainHTTPOutsideLocalhost(t *testing.T) {
	for _, raw := range []string{"http://markupp.example.com", "http://localhost.evil.com", "ftp://x", "not a url"} {
		_, err := CheckBaseURL(raw, false)
		if failure.KindOf(err) != failure.Usage || !strings.Contains(err.Error(), raw) {
			t.Errorf("CheckBaseURL(%q) = %v, want a usage error naming the URL", raw, err)
		}
	}
	_, err := CheckBaseURL("http://markupp.example.com", false)
	if !strings.Contains(err.Error(), "--allow-plain-http") {
		t.Errorf("plain http error %q does not name --allow-plain-http", err)
	}
}

func TestRedirectToAnotherHostDropsCredentials(t *testing.T) {
	other := newFakeServer(t, replyJSON(200, `{}`))
	origin := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL+"/landing", http.StatusFound)
	})
	c := newTestClient(t, origin, "t1")
	body, err := c.open(context.Background(), request{method: http.MethodGet, path: endpoint("notes")})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	_ = body.Close()
	if got := other.seen()[0]; got.auth != "" || got.tenant != "" {
		t.Errorf("redirected request carried auth %q tenant %q, want none", got.auth, got.tenant)
	}
}

func TestRedirectOfABodyToAnotherHostIsRefused(t *testing.T) {
	other := newFakeServer(t, replyJSON(200, `{}`))
	origin := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL+"/landing", http.StatusTemporaryRedirect)
	})
	c := newTestClient(t, origin, "t1")
	err := c.exec(context.Background(), request{method: http.MethodPut, path: endpoint("notes", "n1"), body: []byte(`{"content":"secret"}`)})
	if err == nil {
		t.Fatal("exec followed a 307 with the body to another host, want an error")
	}
	if got := other.seen(); len(got) != 0 {
		t.Errorf("the other host received %d requests, want none: %+v", len(got), got)
	}
}

func TestRedirectOfABodyOnTheSameHostIsFollowed(t *testing.T) {
	server := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/old" {
			http.Redirect(w, r, "/v1/new", http.StatusTemporaryRedirect)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	c := newTestClient(t, server, "t1")
	if err := c.exec(context.Background(), request{method: http.MethodPost, path: "/old", body: []byte(`{"a":1}`)}); err != nil {
		t.Fatalf("exec: %v", err)
	}
	if got := server.seen(); len(got) != 2 || got[1].body != `{"a":1}` {
		t.Errorf("same host 307 requests = %+v, want the body replayed to /v1/new", got)
	}
}

func TestRedirectOnTheSameHostKeepsCredentials(t *testing.T) {
	server := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/old" {
			http.Redirect(w, r, "/v1/new", http.StatusMovedPermanently)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	c := newTestClient(t, server, "t1")
	if err := c.exec(context.Background(), request{method: http.MethodGet, path: "/old"}); err != nil {
		t.Fatalf("exec: %v", err)
	}
	if got := server.seen()[1]; got.auth != "Bearer tok" || got.tenant != "t1" {
		t.Errorf("same host redirect lost credentials: %+v", got)
	}
}

func TestCrossesTrust(t *testing.T) {
	parse := func(raw string) *url.URL { u, _ := url.Parse(raw); return u }
	cases := []struct {
		from, to string
		want     bool
	}{
		{"https://a.com/x", "https://a.com/y", false},
		{"https://a.com/x", "https://b.com/y", true},
		{"https://a.com/x", "http://a.com/y", true},
		{"https://a.com/x", "https://a.com:8443/y", true},
		{"http://localhost:8080/x", "https://localhost:8080/y", false},
	}
	for _, tc := range cases {
		if got := crossesTrust(parse(tc.from), parse(tc.to)); got != tc.want {
			t.Errorf("crossesTrust(%s, %s) = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestUnauthorizedRetriesOnceWithAFreshToken(t *testing.T) {
	server := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer new" {
			replyJSON(401, `{"error":"unauthorized","message":"expired"}`)(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	tokens := &rotatingTokens{first: "old", fresh: "new"}
	c, _ := New(Options{BaseURL: server.URL, Tokens: tokens})
	err := c.exec(context.Background(), request{method: http.MethodPut, path: "/x", body: []byte(`{"a":1}`)})
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	seen := server.seen()
	if len(seen) != 2 || seen[1].body != `{"a":1}` {
		t.Errorf("requests = %+v, want two with the body replayed", seen)
	}
}

func TestUnauthorizedAfterRetryIsReturned(t *testing.T) {
	server := newFakeServer(t, replyJSON(401, `{"error":"unauthorized","message":"expired"}`))
	tokens := &rotatingTokens{first: "old", fresh: "new"}
	c, _ := New(Options{BaseURL: server.URL, Tokens: tokens})
	err := c.exec(context.Background(), request{method: http.MethodGet, path: "/x"})
	if failure.KindOf(err) != failure.Unauthenticated {
		t.Fatalf("exec error = %v, want unauthenticated", err)
	}
	if len(server.seen()) != 2 || len(tokens.rejected) != 1 {
		t.Errorf("requests %d, rejections %d, want 2 and 1", len(server.seen()), len(tokens.rejected))
	}
}

func TestVerboseLogNamesRequestWithoutSecrets(t *testing.T) {
	server := newFakeServer(t, replyJSON(200, `{"content":"secret note"}`))
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	c, _ := New(Options{BaseURL: server.URL, Tokens: StaticToken("secret-token"), Logger: logger})
	body, err := c.open(context.Background(), request{
		method: http.MethodPost, path: "/notes", query: url.Values{"q": {"secret-query"}}, body: []byte(`{"secret":"body"}`),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	_ = readAll(t, body)
	text := logs.String()
	for _, want := range []string{`"method":"POST"`, `"path":"/v1/notes"`, `"status":200`, `"duration_ms"`} {
		if !strings.Contains(text, want) {
			t.Errorf("log %s lacks %s", text, want)
		}
	}
	if strings.Contains(text, "secret") {
		t.Errorf("log %s leaks a token, query or body", text)
	}
}
