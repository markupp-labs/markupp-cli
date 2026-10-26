package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// recordedRequest is what a fake server saw of one request.
type recordedRequest struct {
	method string
	path   string
	query  string
	auth   string
	tenant string
	agent  string
	body   string
}

// fakeServer is an httptest server that records every request and answers with reply.
type fakeServer struct {
	*httptest.Server
	mu       sync.Mutex
	requests []recordedRequest
}

func newFakeServer(t *testing.T, reply http.HandlerFunc) *fakeServer {
	t.Helper()
	fake := &fakeServer{}
	fake.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fake.record(r)
		reply(w, r)
	}))
	t.Cleanup(fake.Close)
	return fake
}

func (f *fakeServer) record(r *http.Request) {
	body := make([]byte, r.ContentLength)
	if r.ContentLength > 0 {
		_, _ = r.Body.Read(body)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requests = append(f.requests, recordedRequest{
		method: r.Method, path: r.URL.EscapedPath(), query: r.URL.RawQuery,
		auth: r.Header.Get("Authorization"), tenant: r.Header.Get("X-Markupp-Tenant"),
		agent: r.Header.Get("User-Agent"), body: string(body),
	})
}

func (f *fakeServer) seen() []recordedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]recordedRequest(nil), f.requests...)
}

func replyJSON(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

// rotatingTokens hands out first, and fresh once after the server rejects first.
type rotatingTokens struct {
	first, fresh string
	rejected     []string
}

func (r *rotatingTokens) Token(context.Context) (string, error) { return r.first, nil }

func (r *rotatingTokens) Rejected(_ context.Context, token string) (string, bool, error) {
	r.rejected = append(r.rejected, token)
	if token != r.first {
		return "", false, nil
	}
	return r.fresh, true, nil
}

func newTestClient(t *testing.T, server *fakeServer, tenant string) *Client {
	t.Helper()
	c, err := New(Options{BaseURL: server.URL, Tenant: tenant, Tokens: StaticToken("tok"), UserAgent: "mu/test"})
	if err != nil {
		t.Fatalf("New(%s): %v", server.URL, err)
	}
	return c
}
