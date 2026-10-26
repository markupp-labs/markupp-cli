package client

import (
	"context"
	"net/url"
	"testing"
)

func TestPublicReadsSendTheirRequests(t *testing.T) {
	orphan, community := true, int64(7)
	checkCalls(t, []apiCall{
		{"notes", func(ctx context.Context, c *Client) error {
			filter := NoteFilter{Orphan: &orphan, Community: &community, Sort: "pagerank", Direction: "asc"}
			return opened(c.PublicNotes(ctx, "docs", filter, Page{Limit: 5}))
		}, recordedRequest{method: "GET", path: "/v1/public/vaults/docs/notes", query: "community=7&direction=asc&limit=5&orphan=true&sort=pagerank"}},
		{"note", func(ctx context.Context, c *Client) error { return opened(c.PublicNote(ctx, "docs", "n/1")) },
			recordedRequest{method: "GET", path: "/v1/public/vaults/docs/notes/n%2F1"}},
		{"search", func(ctx context.Context, c *Client) error {
			return opened(c.PublicSearch(ctx, "docs", "graph theory", Page{Offset: 10}))
		}, recordedRequest{method: "GET", path: "/v1/public/vaults/docs/search", query: "offset=10&q=graph+theory"}},
		{"slug escaped", func(ctx context.Context, c *Client) error { return opened(c.PublicNote(ctx, "../x", "n1")) },
			recordedRequest{method: "GET", path: "/v1/public/vaults/%2E%2E%2Fx/notes/n1"}},
	})
}

func TestPublicReadsSendNoCredentials(t *testing.T) {
	server := newFakeServer(t, replyJSON(200, `[]`))
	c, err := NewPublic(Options{BaseURL: server.URL, Tenant: "t1", Tokens: StaticToken("tok")})
	if err != nil {
		t.Fatalf("NewPublic: %v", err)
	}
	if err := opened(c.PublicNotes(context.Background(), "docs", NoteFilter{}, Page{})); err != nil {
		t.Fatalf("PublicNotes: %v", err)
	}
	if got := server.seen()[0]; got.auth != "" || got.tenant != "" {
		t.Errorf("public request carried auth %q and tenant %q", got.auth, got.tenant)
	}
}

func TestEmptyFiltersSendNothing(t *testing.T) {
	empty := url.Values{}
	GraphFilter{}.apply(empty)
	NoteFilter{}.apply(empty)
	LinkFilter{}.apply(empty)
	if len(empty) != 0 {
		t.Errorf("empty filters sent %q", empty.Encode())
	}
}
